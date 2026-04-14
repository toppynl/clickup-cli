package task

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/toppynl/clickup-cli/api/clickupv2"
	"github.com/toppynl/clickup-cli/internal/apiv2"
	"github.com/toppynl/clickup-cli/internal/git"
	"github.com/toppynl/clickup-cli/pkg/cmdutil"
)

// NewCmdChecklist returns the "task checklist" parent command.
func NewCmdChecklist(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checklist <command>",
		Short: "Manage task checklists",
		Long: `Add, remove, and manage checklists and their items on ClickUp tasks.

To find checklist and item IDs, use: clickup task view <task-id> --json`,
	}

	cmd.AddCommand(newCmdChecklistAdd(f))
	cmd.AddCommand(newCmdChecklistRemove(f))
	cmd.AddCommand(newCmdChecklistItem(f))

	return cmd
}

// --- Checklist Add ---

func newCmdChecklistAdd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "add <task-id> <checklist-name>",
		Short:             "Create a checklist on a task",
		Example:           `  clickup task checklist add 86abc123 "Deploy Steps"`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecklistAdd(f, args[0], args[1])
		},
	}
	return cmd
}

func runChecklistAdd(f *cmdutil.Factory, taskID string, name string) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	parsed := git.ParseTaskID(taskID)
	taskID = parsed.ID

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	resp, err := apiv2.CreateChecklist(ctx, client, taskID, &clickupv2.CreateChecklistJSONRequest{
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("failed to create checklist: %w", err)
	}

	fmt.Fprintf(ios.Out, "%s Created checklist %s on task %s\n",
		cs.Green("!"), cs.Bold(name), cs.Bold(taskID))
	if resp != nil {
		fmt.Fprintf(ios.Out, "  Checklist ID: %s\n", resp.Checklist.ID)
	}

	return nil
}

// --- Checklist Remove ---

func newCmdChecklistRemove(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "remove <checklist-id>",
		Short:             "Delete a checklist",
		Example:           `  clickup task checklist remove b955c4dc-b8ee-4488-b0c1-example`,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecklistRemove(f, args[0])
		},
	}
	return cmd
}

func runChecklistRemove(f *cmdutil.Factory, checklistID string) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = apiv2.DeleteChecklist(ctx, client, checklistID)
	if err != nil {
		return fmt.Errorf("failed to delete checklist: %w", err)
	}

	fmt.Fprintf(ios.Out, "%s Deleted checklist %s\n",
		cs.Green("!"), cs.Bold(checklistID))

	return nil
}

// --- Checklist Item ---

func newCmdChecklistItem(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "item <command>",
		Short: "Manage checklist items",
		Long:  "Add, resolve, and remove items from a checklist.",
	}

	cmd.AddCommand(newCmdChecklistItemAdd(f))
	cmd.AddCommand(newCmdChecklistItemEdit(f))
	cmd.AddCommand(newCmdChecklistItemResolve(f))
	cmd.AddCommand(newCmdChecklistItemRemove(f))

	return cmd
}

func newCmdChecklistItemAdd(f *cmdutil.Factory) *cobra.Command {
	var assignee int

	cmd := &cobra.Command{
		Use:   "add <checklist-id> <item-name>",
		Short: "Add an item to a checklist",
		Example: `  # Add an item
  clickup task checklist item add b955c4dc-example "Run migrations"

  # Add an item assigned to a user (use 'clickup member list' to find IDs)
  clickup task checklist item add b955c4dc-example "Run migrations" --assignee 54874661`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecklistItemAdd(f, args[0], args[1], assignee)
		},
	}

	cmd.Flags().IntVar(&assignee, "assignee", 0, "User ID to assign the item to (see 'clickup member list')")

	return cmd
}

func runChecklistItemAdd(f *cmdutil.Factory, checklistID string, itemName string, assignee int) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	req := &clickupv2.CreateChecklistItemJSONRequest{
		Name: &itemName,
	}
	if assignee != 0 {
		req.Assignee = &assignee
	}

	ctx := context.Background()
	_, err = apiv2.CreateChecklistItem(ctx, client, checklistID, req)
	if err != nil {
		return fmt.Errorf("failed to add checklist item: %w", err)
	}

	msg := fmt.Sprintf("%s Added item %s to checklist", cs.Green("!"), cs.Bold(itemName))
	if assignee != 0 {
		msg += fmt.Sprintf(" (assigned to %d)", assignee)
	}
	fmt.Fprintln(ios.Out, msg)

	return nil
}

func newCmdChecklistItemEdit(f *cmdutil.Factory) *cobra.Command {
	var assignee int
	var name string

	cmd := &cobra.Command{
		Use:   "edit <checklist-id> <item-id> [<item-id>...]",
		Short: "Edit one or more checklist items (rename or assign)",
		Example: `  # Assign a checklist item to a user
  clickup task checklist item edit b955c4dc-example 21e08dc8-example --assignee 54874661

  # Bulk assign all items in a checklist
  clickup task checklist item edit b955c4dc-example item1 item2 item3 --assignee 54874661

  # Rename a checklist item
  clickup task checklist item edit b955c4dc-example 21e08dc8-example --name "Updated name"

  # Unassign (set assignee to 0)
  clickup task checklist item edit b955c4dc-example 21e08dc8-example --assignee 0`,
		Args:              cobra.MinimumNArgs(2),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecklistItemEdit(f, args[0], args[1:], name, assignee, cmd)
		},
	}

	cmd.Flags().IntVar(&assignee, "assignee", -1, "User ID to assign the item to (0 to unassign; see 'clickup member list')")
	cmd.Flags().StringVar(&name, "name", "", "New name for the item (single item only)")

	return cmd
}

func runChecklistItemEdit(f *cmdutil.Factory, checklistID string, itemIDs []string, name string, assignee int, cmd *cobra.Command) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	assigneeChanged := cmd.Flags().Changed("assignee")

	if !assigneeChanged && name == "" {
		return fmt.Errorf("at least one of --assignee or --name must be provided")
	}

	if name != "" && len(itemIDs) > 1 {
		return fmt.Errorf("--name can only be used with a single item ID")
	}

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	bulk := len(itemIDs) > 1
	total := len(itemIDs)
	var updated int

	for i, itemID := range itemIDs {
		req := &clickupv2.EditChecklistItemJSONRequest{}
		if name != "" {
			req.Name = &name
		}
		if assigneeChanged {
			// The v2 API spec defines assignee as Nullable[string].
			s := strconv.Itoa(assignee)
			if assignee == 0 {
				req.Assignee.SetNull()
			} else {
				req.Assignee.Set(s)
			}
		}

		_, err = apiv2.EditChecklistItem(ctx, client, checklistID, itemID, req)
		if err != nil {
			if bulk {
				fmt.Fprintf(ios.ErrOut, "%s (%d/%d) failed to edit %s: %v\n", cs.Red("✗"), i+1, total, itemID, err)
				continue
			}
			return fmt.Errorf("failed to edit checklist item: %w", err)
		}

		updated++
		if bulk {
			fmt.Fprintf(ios.Out, "(%d/%d) Updated checklist item %s\n", i+1, total, cs.Bold(itemID))
		} else {
			fmt.Fprintf(ios.Out, "%s Updated checklist item %s\n", cs.Green("!"), cs.Bold(itemID))
		}
	}

	if bulk {
		fmt.Fprintf(ios.Out, "\n%s Updated %d/%d items\n", cs.Green("!"), updated, total)
	}

	return nil
}

func newCmdChecklistItemResolve(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resolve <checklist-id> <item-id>",
		Short:             "Mark a checklist item as resolved",
		Example:           `  clickup task checklist item resolve b955c4dc-example 21e08dc8-example`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecklistItemResolve(f, args[0], args[1])
		},
	}
	return cmd
}

func runChecklistItemResolve(f *cmdutil.Factory, checklistID string, itemID string) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	resolved := true
	ctx := context.Background()
	_, err = apiv2.EditChecklistItem(ctx, client, checklistID, itemID, &clickupv2.EditChecklistItemJSONRequest{
		Resolved: &resolved,
	})
	if err != nil {
		return fmt.Errorf("failed to resolve checklist item: %w", err)
	}

	fmt.Fprintf(ios.Out, "%s Marked item %s as resolved\n",
		cs.Green("!"), cs.Bold(itemID))

	return nil
}

func newCmdChecklistItemRemove(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "remove <checklist-id> <item-id>",
		Short:             "Remove an item from a checklist",
		Example:           `  clickup task checklist item remove b955c4dc-example 21e08dc8-example`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: cmdutil.NeedsAuth(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecklistItemRemove(f, args[0], args[1])
		},
	}
	return cmd
}

func runChecklistItemRemove(f *cmdutil.Factory, checklistID string, itemID string) error {
	ios := f.IOStreams
	cs := ios.ColorScheme()

	client, err := f.ApiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = apiv2.DeleteChecklistItem(ctx, client, checklistID, itemID)
	if err != nil {
		return fmt.Errorf("failed to remove checklist item: %w", err)
	}

	fmt.Fprintf(ios.Out, "%s Removed item %s\n",
		cs.Green("!"), cs.Bold(itemID))

	return nil
}
