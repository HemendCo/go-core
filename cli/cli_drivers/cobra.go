package cli_drivers

import (
	"HemendCo/go-core/cli/cli_models"

	"github.com/spf13/cobra"
)

type CobraCommand struct {
	cmd           *cobra.Command
	defaultAction *func()
}

func (c *CobraCommand) Name() string {
	return "cobra"
}

func (c *CobraCommand) Init(info cli_models.Info) error {
	if c.cmd != nil {
		return nil
	}

	c.cmd = &cobra.Command{
		Use:   info.Name,
		Short: info.Short,
		Long:  info.Long,
		CompletionOptions: cobra.CompletionOptions{
			// Disable the default completion command for the root command.
			DisableDefaultCmd: true,
			// Hide the default completion command from the command help output.
			HiddenDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			if c.defaultAction != nil {
				(*c.defaultAction)()
			} else {
				cmd.Help()
			}
		},
	}

	if c.defaultAction != nil {
		c.cmd.Run = func(cmd *cobra.Command, args []string) {
			(*c.defaultAction)()
		}
	}

	return nil
}

func (c *CobraCommand) SetDefaultAction(action func()) {
	c.defaultAction = &action
}

func (c *CobraCommand) AddCommand(command cli_models.Command) error {
	cmd := &cobra.Command{
		Use:   command.Name,
		Short: command.Short,
		Run: func(cmd *cobra.Command, args []string) {
			command.Action()
		},
	}
	c.cmd.AddCommand(cmd)
	return nil
}

func (c *CobraCommand) AddCommands(commands ...*cli_models.Command) error {
	for _, command := range commands {
		if err := c.AddCommand(*command); err != nil {
			return err
		}
	}
	return nil
}

func (c *CobraCommand) Execute() error {
	return c.cmd.Execute()
}
