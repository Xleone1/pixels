// Package bypass implements the privileged catalog pricing toggle.
package bypass

import (
	"github.com/niflaot/pixels/internal/permission"
	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	catalogpricing "github.com/niflaot/pixels/internal/realm/catalog/pricing"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkcommand "github.com/niflaot/pixels/sdk/command"
	"go.minekube.com/brigodier"
)

var (
	// TogglePermission allows changing the process-local free catalog mode.
	TogglePermission = permission.RegisterNode("catalog.bypass.toggle", "")
	// commandScope owns the built-in catalog command root.
	commandScope = pluginruntime.NewScope("catalog")
)

// Register claims and registers the catalog bypass command.
func Register(tree *plugincommand.Tree, state *catalogpricing.State, translations i18n.Translator) error {
	access := plugincommand.NewAccess(tree, commandScope)

	return access.Register(command(state, translations))
}

// command builds the catalog bypass command tree.
func command(state *catalogpricing.State, translations i18n.Translator) brigodier.LiteralNodeBuilder {
	return brigodier.Literal("catalogbypass").
		Requires(sdkcommand.RequiresPermission(string(TogglePermission))).
		Executes(brigodier.CommandFunc(func(call *brigodier.CommandContext) error {
			sender, found := sdkcommand.SenderFrom(call.Context)
			if !found {
				return nil
			}
			enabled := state.ToggleFreeItems()
			if enabled {
				return sender.Reply(call.Context, message(translations, "catalog.command.bypass.enabled", "Catálogo gratuito activado. Reabre la tienda para actualizar los precios."))
			}

			return sender.Reply(call.Context, message(translations, "catalog.command.bypass.disabled", "Catálogo gratuito desactivado. Reabre la tienda para restaurar los precios."))
		}))
}

// message resolves localized command feedback.
func message(translations i18n.Translator, key string, fallback string) string {
	if translations == nil {
		return fallback
	}
	translated := translations.Default(i18n.Key(key))
	if translated == key {
		return fallback
	}

	return translated
}
