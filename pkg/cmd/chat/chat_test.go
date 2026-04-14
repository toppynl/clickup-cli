package chat

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/triptechtravel/clickup-cli/internal/testutil"
)

func TestNewCmdChat_HasSendSubcommand(t *testing.T) {
	cmd := NewCmdChat(nil)
	assert.Equal(t, "chat <command>", cmd.Use)

	sub, _, err := cmd.Find([]string{"send"})
	require.NoError(t, err)
	assert.Equal(t, "send", sub.Name())
}

func TestChatSend_SendsCorrectRequest(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	var capturedBody map[string]interface{}
	var capturedMethod string
	var capturedPath string

	tf.HandleFuncV3("workspaces/12345/chat/channels/chan-abc/messages", func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"content": "Hello team!",
			"type": "message",
			"date": 1700000000000,
			"user_id": "user1",
			"resolved": false,
			"links": {"reactions": "", "replies": "", "tagged_users": ""},
			"replies_count": 0
		}`))
	})

	cmd := NewCmdSend(tf.Factory)
	err := testutil.RunCommand(t, cmd, "chan-abc", "Hello team!")
	require.NoError(t, err)

	assert.Equal(t, "POST", capturedMethod)
	assert.Contains(t, capturedPath, "/workspaces/12345/chat/channels/chan-abc/messages")
	assert.Equal(t, "message", capturedBody["type"])
	assert.Equal(t, "Hello team!", capturedBody["content"])
	assert.Equal(t, "text/md", capturedBody["content_format"])

	out := tf.OutBuf.String()
	assert.Contains(t, out, "Message sent")
	assert.Contains(t, out, "chan-abc")
}

func TestChatSend_JSONOutput(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	tf.HandleFuncV3("workspaces/12345/chat/channels/chan-abc/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"content": "Hi",
			"type": "message",
			"date": 1700000000000,
			"user_id": "user1",
			"resolved": false,
			"links": {"reactions": "", "replies": "", "tagged_users": ""},
			"replies_count": 0
		}`))
	})

	cmd := NewCmdSend(tf.Factory)
	err := testutil.RunCommand(t, cmd, "chan-abc", "Hi", "--json")
	require.NoError(t, err)

	out := tf.OutBuf.String()
	// Should be valid JSON
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	assert.Equal(t, "Hi", parsed["content"])
}

func TestChatSend_RequiresArgs(t *testing.T) {
	cmd := NewCmdSend(nil)
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"chan-only"}))
	assert.NoError(t, cmd.Args(cmd, []string{"chan-id", "message"}))
	assert.Error(t, cmd.Args(cmd, []string{"a", "b", "c"}))
}
