package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triptechtravel/clickup-cli/internal/testutil"
)

func TestNewCmdCreate_Flags(t *testing.T) {
	cmd := NewCmdCreate(nil)

	assert.NotNil(t, cmd.Flags().Lookup("list-id"))
	assert.NotNil(t, cmd.Flags().Lookup("list-name"))
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("status"))
	assert.NotNil(t, cmd.Flags().Lookup("priority"))
	assert.NotNil(t, cmd.Flags().Lookup("assignee"))
	assert.NotNil(t, cmd.Flags().Lookup("tags"))
	assert.NotNil(t, cmd.Flags().Lookup("due-date"))
	assert.NotNil(t, cmd.Flags().Lookup("start-date"))
	assert.NotNil(t, cmd.Flags().Lookup("time-estimate"))
	assert.NotNil(t, cmd.Flags().Lookup("points"))
}

func TestNewCmdCreate_ListNameMutuallyExclusive(t *testing.T) {
	tf := testutil.NewTestFactory(t)
	cmd := NewCmdCreate(tf.Factory)
	err := testutil.RunCommand(t, cmd, "--list-id", "123", "--list-name", "My List", "--name", "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "if any flags in the group [list-id list-name current] are set none of the others can be")
}
