package doc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

type createOptions struct {
	name        string
	parentID    string
	parentType  string
	visibility  string
	createPage  bool
	jsonFlags   cmdutil.JSONFlags
}

// NewCmdCreate returns a command to create a new ClickUp Doc.
func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{
		createPage: true,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ClickUp Doc",
		Long: `Create a new Doc in the configured ClickUp workspace.

Optionally scope the Doc to a parent space, folder, or list.
The --create-page flag (default true) creates an initial empty page.`,
		Example: `  # Create a Doc with default visibility
  clickup doc create --name "Project Runbook"

  # Create a Doc in a specific space
  clickup doc create --name "Team Wiki" --parent-id 123456 --parent-type SPACE

  # Create a Doc with public visibility and no initial page
  clickup doc create --name "Public Docs" --visibility PUBLIC --create-page=false

  # Create and output JSON
  clickup doc create --name "API Reference" --json`,
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.name == "" {
				return fmt.Errorf("--name is required")
			}
			if opts.parentType != "" {
				if _, err := parseParentType(opts.parentType); err != nil {
					return err
				}
			}
			if opts.visibility != "" && !containsString(validVisibility, opts.visibility) {
				return fmt.Errorf("invalid visibility %q; valid values: %s", opts.visibility, strings.Join(validVisibility, "|"))
			}
			return runCreate(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.name, "name", "", "Doc name (required)")
	cmd.Flags().StringVar(&opts.parentID, "parent-id", "", "Parent ID (space, folder, or list)")
	cmd.Flags().StringVar(&opts.parentType, "parent-type", "", "Parent type (SPACE|FOLDER|LIST|WORKSPACE|EVERYTHING)")
	cmd.Flags().StringVar(&opts.visibility, "visibility", "", "Visibility (PUBLIC|PRIVATE|PERSONAL|HIDDEN)")
	cmd.Flags().BoolVar(&opts.createPage, "create-page", true, "Create an initial empty page in the Doc")

	cmdutil.AddJSONFlags(cmd, &opts.jsonFlags)

	return cmd
}

func runCreate(f *cmdutil.Factory, opts *createOptions) error {
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

	body := map[string]interface{}{
		"name":        opts.name,
		"create_page": opts.createPage,
	}

	if opts.parentID != "" {
		pt, _ := parseParentType(opts.parentType)
		body["parent"] = map[string]interface{}{
			"id":   opts.parentID,
			"type": pt,
		}
	}

	if opts.visibility != "" {
		body["visibility"] = strings.ToUpper(opts.visibility)
	}

	url := fmt.Sprintf("%s/workspaces/%s/docs", apiBase, workspaceID)

	ctx := context.Background()
	data, status, err := doRequest(ctx, client, "POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create doc: %w", err)
	}
	if status != 200 && status != 201 {
		return fmt.Errorf("failed to create doc: status %d: %s", status, string(data))
	}

	var d docCore
	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.jsonFlags.WantsJSON() {
		return opts.jsonFlags.OutputJSON(ios.Out, d)
	}

	fmt.Fprintf(ios.Out, "%s Created Doc %s %s\n", cs.Green("!"), cs.Bold(d.Name), cs.Gray("#"+d.ID))

	fmt.Fprintln(ios.Out)
	fmt.Fprintln(ios.Out, cs.Gray("---"))
	fmt.Fprintln(ios.Out, cs.Gray("Quick actions:"))
	fmt.Fprintf(ios.Out, "  %s  clickup doc view %s\n", cs.Gray("View:"), d.ID)
	fmt.Fprintf(ios.Out, "  %s  clickup doc page list %s\n", cs.Gray("Pages:"), d.ID)
	fmt.Fprintf(ios.Out, "  %s  clickup doc page create %s --name \"My Page\"\n", cs.Gray("Add page:"), d.ID)

	return nil
}
