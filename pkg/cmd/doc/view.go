package doc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

type viewOptions struct {
	docID     string
	jsonFlags cmdutil.JSONFlags
}

// NewCmdView returns a command to view a single ClickUp Doc.
func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <doc-id>",
		Short: "View a ClickUp Doc",
		Long:  `Display details about a ClickUp Doc including its metadata and parent location.`,
		Example: `  # View a Doc
  clickup doc view abc123

  # View as JSON
  clickup doc view abc123 --json`,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.docID = args[0]
			return runView(f, opts)
		},
	}

	cmdutil.AddJSONFlags(cmd, &opts.jsonFlags)

	return cmd
}

func runView(f *cmdutil.Factory, opts *viewOptions) error {
	ios := f.IOStreams

	workspaceID, err := resolveWorkspaceID(f)
	if err != nil {
		return err
	}

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/workspaces/%s/docs/%s", apiBase, workspaceID, opts.docID)

	ctx := context.Background()
	data, status, err := doRequest(ctx, client, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch doc: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to fetch doc: status %d: %s", status, string(data))
	}

	var d docCore
	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.jsonFlags.WantsJSON() {
		return opts.jsonFlags.OutputJSON(ios.Out, d)
	}

	return printDocView(f, &d)
}

func printDocView(f *cmdutil.Factory, d *docCore) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()
	out := ios.Out

	fmt.Fprintf(out, "%s %s\n", cs.Bold(d.Name), cs.Gray("#"+d.ID))
	fmt.Fprintf(out, "%s %s\n", cs.Bold("Visibility:"), strings.ToLower(d.Visibility))

	if d.Creator.Username != "" {
		fmt.Fprintf(out, "%s %s\n", cs.Bold("Creator:"), d.Creator.Username)
	}

	if d.Parent.ID != "" {
		fmt.Fprintf(out, "%s %s (type %d)\n", cs.Bold("Parent:"), d.Parent.ID, d.Parent.Type)
	}

	if d.Deleted {
		fmt.Fprintf(out, "%s %s\n", cs.Bold("Status:"), cs.Red("deleted"))
	} else if d.Archived {
		fmt.Fprintf(out, "%s %s\n", cs.Bold("Status:"), cs.Gray("archived"))
	}

	if d.DateCreated != "" {
		fmt.Fprintf(out, "%s %s\n", cs.Bold("Created:"), d.DateCreated)
	}
	if d.DateUpdated != "" {
		fmt.Fprintf(out, "%s %s\n", cs.Bold("Updated:"), d.DateUpdated)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, cs.Gray("---"))
	fmt.Fprintln(out, cs.Gray("Quick actions:"))
	fmt.Fprintf(out, "  %s  clickup doc page list %s\n", cs.Gray("Pages:"), d.ID)
	fmt.Fprintf(out, "  %s  clickup doc page create %s --name \"My Page\"\n", cs.Gray("Add page:"), d.ID)
	fmt.Fprintf(out, "  %s  clickup doc view %s --json\n", cs.Gray("JSON:"), d.ID)

	return nil
}
