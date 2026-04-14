package field

import (
	"github.com/spf13/cobra"
	"github.com/toppynl/clickup-cli/pkg/cmdutil"
)

// NewCmdField returns the top-level "field" command.
func NewCmdField(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "field <command>",
		Short: "Manage custom fields",
		Long:  "Discover and inspect custom fields available in your ClickUp lists.",
	}

	cmd.AddCommand(NewCmdFieldList(f))

	return cmd
}
