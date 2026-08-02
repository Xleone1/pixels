package tools

import (
	"context"

	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkcommand "github.com/niflaot/pixels/sdk/command"
	"go.minekube.com/brigodier"
)

// commandScope owns the built-in Creator Tools command root.
var commandScope = pluginruntime.NewScope("wired-tools")

// RegisterCommands installs every Creator Tools chat entrypoint.
func RegisterCommands(tree *plugincommand.Tree, service *Service) error {
	access := plugincommand.NewAccess(tree, commandScope)
	commands := [][2]string{
		{"wired", "monitor"},
		{"wf", "monitor"},
		{"var", "variables"},
		{"inspect", "inspection"},
	}
	for _, command := range commands {
		if err := registerCommand(access, service, command[0], command[1]); err != nil {
			return err
		}
	}
	return nil
}

// registerCommand installs one tab-focused Creator Tools command.
func registerCommand(access *plugincommand.Access, service *Service, name string, tab string) error {
	return access.Register(brigodier.Literal(name).
		Requires(func(ctx context.Context) bool {
			sender, found := sdkcommand.SenderFrom(ctx)
			return found && sender.Kind() == sdkcommand.SenderKindPlayer
		}).
		Executes(brigodier.CommandFunc(func(call *brigodier.CommandContext) error {
			sender, found := sdkcommand.SenderFrom(call.Context)
			if !found {
				return nil
			}
			return service.Open(call.Context, sender.Name(), tab)
		})))
}
