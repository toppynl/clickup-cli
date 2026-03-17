package doc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/triptechtravel/clickup-cli/internal/iostreams"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

type pageListOptions struct {
	docID     string
	maxDepth  int
	jsonFlags cmdutil.JSONFlags
}

// NewCmdPageList returns a command to list pages in a ClickUp Doc.
func NewCmdPageList(f *cmdutil.Factory) *cobra.Command {
	opts := &pageListOptions{}

	cmd := &cobra.Command{
		Use:   "list <doc-id>",
		Short: "List pages in a ClickUp Doc",
		Long:  `List all pages in a ClickUp Doc. Pages are returned as a tree structure; use --max-depth to control nesting depth.`,
		Example: `  # List all pages in a Doc
  clickup doc page list abc123

  # List top-level pages only
  clickup doc page list abc123 --max-depth 0

  # List pages as JSON
  clickup doc page list abc123 --json`,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.docID = args[0]
			return runPageList(f, opts)
		},
	}

	cmd.Flags().IntVar(&opts.maxDepth, "max-depth", -1, "Maximum page nesting depth (-1 for unlimited)")

	cmdutil.AddJSONFlags(cmd, &opts.jsonFlags)

	return cmd
}

func runPageList(f *cmdutil.Factory, opts *pageListOptions) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	workspaceID, err := resolveWorkspaceID(f)
	if err != nil {
		return err
	}

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/workspaces/%s/docs/%s/pages", apiBase, workspaceID, opts.docID)
	if opts.maxDepth >= 0 {
		url += fmt.Sprintf("?max_page_depth=%d", opts.maxDepth)
	}

	ctx := context.Background()
	data, status, err := doRequest(ctx, client, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to list pages: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to list pages: status %d: %s", status, string(data))
	}

	var result pagesListResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.jsonFlags.WantsJSON() {
		return opts.jsonFlags.OutputJSON(ios.Out, result)
	}

	if len(result.Pages) == 0 {
		fmt.Fprintln(ios.Out, cs.Gray("No pages found."))
		return nil
	}

	printPageTree(ios.Out, result.Pages, 0, cs)
	return nil
}

func printPageTree(out io.Writer, pages []pageRef, depth int, cs *iostreams.ColorScheme) {
	for _, p := range pages {
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		fmt.Fprintf(out, "%s%s %s\n", indent, cs.Bold(p.Name), cs.Gray("#"+p.ID))
		if len(p.Pages) > 0 {
			printPageTree(out, p.Pages, depth+1, cs)
		}
	}
}
