package cli

import "github.com/HemendCo/go-core/cli/cli_models"

type CLIDriver interface {
	Name() string
	Init(info cli_models.Info) error
	SetDefaultAction(func())
	AddCommand(command cli_models.Command) error
	AddCommands(commands ...*cli_models.Command) error
	Execute() error
}
