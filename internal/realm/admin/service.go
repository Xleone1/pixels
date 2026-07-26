// Package admin implements first-party operational chat commands.
package admin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/internal/plugin/loader"
	admintrace "github.com/niflaot/pixels/internal/realm/admin/trace"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outnotification "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkcommand "github.com/niflaot/pixels/sdk/command"
	"go.uber.org/zap"
)

const effectConfirmationSound = "sound_respect_received"

// Service implements first-party administrative command behavior.
type Service struct {
	// config controls command compatibility behavior.
	config Config
	// build stores current release metadata.
	build build.Info
	// plugins reports loaded dynamic plugins.
	plugins *loader.Loader
	// players stores current connected players.
	players *playerlive.Registry
	// bindings maps players to active connections.
	bindings *binding.Registry
	// connections sends protocol packets.
	connections *netconn.Registry
	// tracer captures selected bidirectional traffic.
	tracer TraceManager
	// effects selects and clears player-owned avatar effects.
	effects playereffect.Manager
	// translations localizes command feedback.
	translations i18n.Translator
	// log records accountable administrative actions.
	log *zap.Logger
}

// TraceManager owns the command-facing packet trace lifecycle.
type TraceManager interface {
	// Activate starts or returns one player trace.
	Activate(context.Context, int64, string) (admintrace.Session, bool, error)
	// Active returns one active trace snapshot.
	Active(int64) (admintrace.Session, bool)
	// Finalize uploads and removes one active trace.
	Finalize(context.Context, int64, string) (admintrace.Result, error)
}

// New creates the first-party administrative command service.
func New(config Config, info build.Info, plugins *loader.Loader, players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, tracer *admintrace.Tracer, effects playereffect.Manager, translations i18n.Translator, log *zap.Logger) *Service {
	if log == nil {
		log = zap.NewNop()
	}

	return &Service{config: config, build: info, plugins: plugins, players: players, bindings: bindings, connections: connections, tracer: tracer, effects: effects, translations: translations, log: log}
}

// Alert sends one popup to an exact connected username.
func (service *Service) Alert(ctx context.Context, sender sdkcommand.Sender, username string, reason string) error {
	target, found := service.findPlayer(username)
	if !found {
		return sender.Reply(ctx, service.message("admin.command.alert.offline", "El jugador {player} no está conectado.", i18n.Params{"player": strings.TrimSpace(username)}))
	}
	issuerID := service.senderID(sender)
	if target.ID() == issuerID {
		return sender.Reply(ctx, service.message("admin.command.alert.self", "No puedes enviarte una alerta a ti mismo."))
	}
	packet, err := outalert.Encode(strings.TrimSpace(reason))
	if err != nil {
		return service.replyFailure(ctx, sender, err)
	}
	connection, err := service.connection(target.ID())
	if err != nil {
		return sender.Reply(ctx, service.message("admin.command.alert.offline", "El jugador {player} no está conectado.", i18n.Params{"player": target.Username()}))
	}
	if err = connection.Send(ctx, packet); err != nil {
		return service.replyFailure(ctx, sender, err)
	}
	service.log.Info("admin alert sent", zap.Int64("issuer_id", issuerID), zap.String("issuer_name", sender.Name()), zap.Int64("target_id", target.ID()), zap.String("target_name", target.Username()), zap.String("reason", strings.TrimSpace(reason)))

	return sender.Reply(ctx, service.message("admin.command.alert.sent", "Alerta enviada a {player}.", i18n.Params{"player": target.Username()}))
}

// HotelAlert sends one popup to every currently connected player.
func (service *Service) HotelAlert(ctx context.Context, sender sdkcommand.Sender, reason string) error {
	packet, err := outalert.Encode(strings.TrimSpace(reason))
	if err != nil {
		return service.replyFailure(ctx, sender, err)
	}
	issuerID := service.senderID(sender)
	delivered := 0
	failed := 0
	for _, target := range service.players.Snapshot() {
		if target.ID() == issuerID {
			continue
		}
		connection, connectionErr := service.connection(target.ID())
		if connectionErr != nil || connection.Send(ctx, packet) != nil {
			failed++
			continue
		}
		delivered++
	}
	service.log.Info("admin hotel alert sent", zap.Int64("issuer_id", issuerID), zap.String("issuer_name", sender.Name()), zap.String("reason", strings.TrimSpace(reason)), zap.Int("delivered", delivered), zap.Int("failed", failed))

	return sender.Reply(ctx, service.message("admin.command.halert.sent", "Alerta enviada a {delivered} jugadores; {failed} fallaron.", i18n.Params{"delivered": strconv.Itoa(delivered), "failed": strconv.Itoa(failed)}))
}

// About replies with current build and loaded plugin metadata.
func (service *Service) About(ctx context.Context, sender sdkcommand.Sender) error {
	report := service.plugins.Report()
	names := strings.Join(report.Loaded, ", ")
	if names == "" {
		names = service.message("admin.command.about.none", "ninguno")
	}
	message := service.message(
		"admin.command.about.result",
		"Pixels {version} (commit {commit}) — {count} plugin(s) cargado(s): {plugins}",
		i18n.Params{"version": service.build.Version, "commit": service.build.Commit, "count": strconv.Itoa(len(report.Loaded)), "plugins": names},
	)

	return sender.Reply(ctx, message)
}

// ToggleTrace starts or finalizes the issuing player's packet trace.
func (service *Service) ToggleTrace(ctx context.Context, sender sdkcommand.Sender) error {
	player, found := service.findPlayer(sender.Name())
	if !found {
		return sender.Reply(ctx, service.message("admin.command.trace.unavailable", "No se pudo resolver tu sesión conectada."))
	}
	if _, active := service.tracer.Active(player.ID()); active {
		result, err := service.tracer.Finalize(ctx, player.ID(), "manual")
		if err != nil {
			return service.replyFailure(ctx, sender, err)
		}

		return sender.Reply(ctx, service.message("admin.command.trace.saved", "Trace guardado: {url}", i18n.Params{"url": result.URL}))
	}
	session, _, err := service.tracer.Activate(ctx, player.ID(), player.Username())
	if err != nil {
		return service.replyFailure(ctx, sender, err)
	}

	return sender.Reply(ctx, service.message("admin.command.trace.started", "Trace activado hasta {expires}.", i18n.Params{"expires": session.ExpiresAt.Format("2006-01-02 15:04:05 MST")}))
}

// Effect selects one owned effect or clears the current selection with id zero.
func (service *Service) Effect(ctx context.Context, sender sdkcommand.Sender, effectID int32) error {
	permitted := sender.HasPermission(string(EffectPermission))
	if !permitted && (!service.config.AllowUnpermittedEffectClear || effectID != 0) {
		return sender.Reply(ctx, service.message("admin.command.effect.denied", "No tienes permiso para seleccionar ese efecto."))
	}
	if effectID < 0 {
		return sender.Reply(ctx, service.message("admin.command.effect.invalid", "El identificador del efecto no es válido."))
	}
	player, found := service.findPlayer(sender.Name())
	if !found || service.effects == nil {
		return sender.Reply(ctx, service.message("admin.command.effect.unavailable", "No se pudo resolver tu sesión o el sistema de efectos."))
	}
	if err := service.effects.Enable(ctx, player.ID(), effectID); err != nil {
		if errors.Is(err, playereffect.ErrEffectNotFound) {
			return sender.Reply(ctx, service.message("admin.command.effect.not_owned", "No tienes ese efecto disponible."))
		}
		return service.replyFailure(ctx, sender, err)
	}
	service.log.Info("avatar effect command used", zap.Int64("player_id", player.ID()), zap.String("player_name", player.Username()), zap.Int32("effect_id", effectID), zap.Bool("permission", permitted))
	if err := service.sendEffectConfirmation(ctx, player.ID()); err != nil {
		service.log.Warn("avatar effect confirmation sound failed", zap.Int64("player_id", player.ID()), zap.Error(err))
	}

	return nil
}

// sendEffectConfirmation sends a sound-only Nitro notification.
func (service *Service) sendEffectConfirmation(ctx context.Context, playerID int64) error {
	packet, err := outnotification.Encode(
		"effect.confirmation",
		"",
		outnotification.WithParam("display", "SOUND"),
		outnotification.WithParam("sound", effectConfirmationSound),
	)
	if err != nil {
		return err
	}
	connection, err := service.connection(playerID)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// canUseEffectCommand exposes the root to staff or to self-clear compatibility.
func (service *Service) canUseEffectCommand(ctx context.Context) bool {
	sender, found := sdkcommand.SenderFrom(ctx)
	return found && (sender.HasPermission(string(EffectPermission)) || service.config.AllowUnpermittedEffectClear)
}

// senderID resolves the connected sender id for audit logs.
func (service *Service) senderID(sender sdkcommand.Sender) int64 {
	player, found := service.findPlayer(sender.Name())
	if !found {
		return 0
	}

	return player.ID()
}

// replyFailure logs one internal failure and returns localized feedback.
func (service *Service) replyFailure(ctx context.Context, sender sdkcommand.Sender, err error) error {
	service.log.Error("admin command failed", zap.String("issuer_name", sender.Name()), zap.Error(err))

	return sender.Reply(ctx, service.message("admin.command.failed", "No se pudo completar el comando. Inténtalo de nuevo."))
}

// message resolves one localized command value with a fallback.
func (service *Service) message(key string, fallback string, params ...i18n.Params) string {
	if service.translations == nil {
		return interpolate(fallback, params...)
	}
	value := service.translations.Default(i18n.Key(key), params...)
	if value == key {
		return interpolate(fallback, params...)
	}

	return value
}

// interpolate applies fallback translation parameters.
func interpolate(value string, params ...i18n.Params) string {
	for _, replacements := range params {
		for key, replacement := range replacements {
			value = strings.ReplaceAll(value, fmt.Sprintf("{%s}", key), replacement)
		}
	}

	return value
}
