package doc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

type listOptions struct {
	creator    int
	deleted    bool
	archived   bool
	parentID   string
	parentType string
	limit      int
	cursor     string
	jsonFlags  cmdutil.JSONFlags
}

// NewCmdList returns a command to list ClickUp Docs in the workspace.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ClickUp Docs in the workspace",
		Long: `List Docs in the configured ClickUp workspace.

Supports filtering by creator, status, parent location, and pagination.`,
		Example: `  # List all Docs
  clickup doc list

  # List non-deleted, non-archived Docs in JSON
  clickup doc list --json

  # List Docs in a specific space
  clickup doc list --parent-id 123456 --parent-type SPACE

  # Paginate
  clickup doc list --limit 10 --cursor <cursor>`,
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.parentType != "" {
				if _, err := parseParentType(opts.parentType); err != nil {
					return err
				}
			}
			return runList(f, opts)
		},
	}

	cmd.Flags().IntVar(&opts.creator, "creator", 0, "Filter by creator user ID")
	cmd.Flags().BoolVar(&opts.deleted, "deleted", false, "Include deleted Docs")
	cmd.Flags().BoolVar(&opts.archived, "archived", false, "Include archived Docs")
	cmd.Flags().StringVar(&opts.parentID, "parent-id", "", "Filter by parent ID")
	cmd.Flags().StringVar(&opts.parentType, "parent-type", "", "Parent type (SPACE|FOLDER|LIST|WORKSPACE|EVERYTHING)")
	cmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of Docs to return")
	cmd.Flags().StringVar(&opts.cursor, "cursor", "", "Pagination cursor from a previous response")

	cmdutil.AddJSONFlags(cmd, &opts.jsonFlags)

	return cmd
}

func runList(f *cmdutil.Factory, opts *listOptions) error {
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

	url := fmt.Sprintf("%s/workspaces/%s/docs", apiBase, workspaceID)

	// Build query params
	var params []string
	if opts.deleted {
		params = append(params, "deleted=true")
	}
	if opts.archived {
		params = append(params, "archived=true")
	}
	if opts.creator != 0 {
		params = append(params, fmt.Sprintf("creator=%d", opts.creator))
	}
	if opts.parentID != "" {
		params = append(params, "parent_id="+opts.parentID)
		if opts.parentType != "" {
			pt, _ := parseParentType(opts.parentType)
			params = append(params, fmt.Sprintf("parent_type=%d", pt))
		}
	}
	if opts.limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", opts.limit))
	}
	if opts.cursor != "" {
		params = append(params, "cursor="+opts.cursor)
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	ctx := context.Background()
	data, status, err := doRequest(ctx, client, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to list docs: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to list docs: status %d: %s", status, string(data))
	}

	var result docsListResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.jsonFlags.WantsJSON() {
		return opts.jsonFlags.OutputJSON(ios.Out, result)
	}

	if len(result.Docs) == 0 {
		fmt.Fprintln(ios.Out, cs.Gray("No Docs found."))
		return nil
	}

	for _, d := range result.Docs {
		status := ""
		if d.Deleted {
			status = cs.Red(" [deleted]")
		} else if d.Archived {
			status = cs.Gray(" [archived]")
		}
		vis := cs.Gray(fmt.Sprintf(" (%s)", strings.ToLower(d.Visibility)))
		fmt.Fprintf(ios.Out, "%s %s%s%s\n", cs.Bold(d.Name), cs.Gray("#"+d.ID), vis, status)
	}

	if result.NextCursor != "" {
		fmt.Fprintf(ios.Out, "\n%s  clickup doc list --cursor %s\n", cs.Gray("Next page:"), result.NextCursor)
	}

	return nil
}
