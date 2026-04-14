package chat

import (
	"github.com/spf13/cobra"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

// NewCmdChat returns the top-level "chat" command that groups chat subcommands.
func NewCmdChat(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat <command>",
		Short: "Manage ClickUp Chat messages",
		Long:  "Send messages to ClickUp Chat channels.",
	}

	cmd.AddCommand(NewCmdSend(f))

	return cmd
}
