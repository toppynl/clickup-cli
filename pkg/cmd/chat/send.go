package chat

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	clickupv3 "github.com/triptechtravel/clickup-cli/api/clickupv3"
	"github.com/triptechtravel/clickup-cli/internal/apiv3"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

type sendOptions struct {
	channelID string
	message   string
	jsonFlags cmdutil.JSONFlags
}

// NewCmdSend returns the "chat send" command.
func NewCmdSend(f *cmdutil.Factory) *cobra.Command {
	opts := &sendOptions{}

	cmd := &cobra.Command{
		Use:   "send <channel-id> <message>",
		Short: "Send a message to a ClickUp Chat channel",
		Long: `Send a message to a ClickUp Chat channel using the v3 API.

The channel ID can be found in the ClickUp Chat URL or via the API.`,
		Example: `  # Send a message to a channel
  clickup chat send 12345abc "Hello team!"

  # Send and get JSON response
  clickup chat send 12345abc "Deploy complete" --json`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.channelID = args[0]
			opts.message = args[1]
			return runSend(f, opts)
		},
	}

	cmdutil.AddJSONFlags(cmd, &opts.jsonFlags)

	return cmd
}

func runSend(f *cmdutil.Factory, opts *sendOptions) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	cfg, err := f.Config()
	if err != nil {
		return err
	}
	if cfg.Workspace == "" {
		return fmt.Errorf("no workspace configured; run 'clickup auth' first")
	}

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	req := &clickupv3.CommentCreateChatMessage{
		Type:    "message",
		Content: opts.message,
	}
	req.ApplyDefaults()

	resp, err := apiv3.CreateChatMessage(context.Background(), client, cfg.Workspace, opts.channelID, req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if opts.jsonFlags.WantsJSON() {
		return opts.jsonFlags.OutputJSON(ios.Out, resp)
	}

	fmt.Fprintf(ios.Out, "%s Message sent to channel %s\n", cs.Green("!"), cs.Bold(opts.channelID))

	return nil
}
