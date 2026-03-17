package doc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/triptechtravel/clickup-cli/internal/api"
	"github.com/triptechtravel/clickup-cli/pkg/cmdutil"
)

const apiBase = "https://api.clickup.com/api/v3"

// docCore holds the fields common to list and detail responses.
type docCore struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Deleted    bool   `json:"deleted"`
	Archived   bool   `json:"archived"`
	Visibility string `json:"visibility"`
	Creator    struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"creator"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
	Parent      struct {
		ID   string `json:"id"`
		Type int    `json:"type"`
	} `json:"parent"`
	Workspace struct {
		ID string `json:"id"`
	} `json:"workspace"`
}

type docsListResponse struct {
	Docs       []docCore `json:"docs"`
	NextCursor string    `json:"next_cursor"`
}

// pageRef is a summary entry in the page listing response.
type pageRef struct {
	ID        string `json:"id"`
	DocID     string `json:"doc_id"`
	Name      string `json:"name"`
	SubTitle  string `json:"sub_title"`
	OrderIndex int   `json:"order_index"`
	Pages     []pageRef `json:"pages"`
}

type pagesListResponse struct {
	Pages []pageRef `json:"pages"`
}

// pageDetail is the full page returned by the get-page endpoint.
type pageDetail struct {
	ID            string `json:"id"`
	DocID         string `json:"doc_id"`
	Name          string `json:"name"`
	SubTitle      string `json:"sub_title"`
	Content       string `json:"content"`
	ContentFormat string `json:"content_format"`
	OrderIndex    int    `json:"order_index"`
	DateCreated   string `json:"date_created"`
	DateUpdated   string `json:"date_updated"`
	Creator       struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"creator"`
	Pages []pageRef `json:"pages"`
}

// validParentTypes are the accepted string values for --parent-type.
var validParentTypes = map[string]int{
	"SPACE":     4,
	"FOLDER":    5,
	"LIST":      6,
	"WORKSPACE": 7,
	"EVERYTHING": 12,
}

// validVisibility are the accepted string values for --visibility.
var validVisibility = []string{"PUBLIC", "PRIVATE", "PERSONAL", "HIDDEN"}

// validContentFormats are the accepted values for --content-format.
var validContentFormats = []string{"text/md", "text/plain"}

// validEditModes are the accepted values for --content-edit-mode.
var validEditModes = []string{"replace", "append", "prepend"}

// resolveWorkspaceID returns the workspace ID from config.
func resolveWorkspaceID(f *cmdutil.Factory) (string, error) {
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	if cfg.Workspace == "" {
		return "", fmt.Errorf("no workspace configured; run 'clickup auth login' to set it up")
	}
	return cfg.Workspace, nil
}

// doRequest performs a JSON API request against the ClickUp v3 Docs API.
func doRequest(ctx context.Context, client *api.Client, method, url string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.DoRequest(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return data, resp.StatusCode, nil
}

// parseParentType converts a string like "SPACE" or "4" to the int type used by the API.
func parseParentType(s string) (int, error) {
	upper := strings.ToUpper(s)
	if v, ok := validParentTypes[upper]; ok {
		return v, nil
	}
	// Try numeric fallback
	var n int
	if _, err := fmt.Sscan(s, &n); err == nil {
		return n, nil
	}
	keys := make([]string, 0, len(validParentTypes))
	for k := range validParentTypes {
		keys = append(keys, k)
	}
	return 0, fmt.Errorf("invalid parent type %q; valid values: %s", s, strings.Join(keys, "|"))
}

// containsString returns true if s is in the slice (case-insensitive).
func containsString(slice []string, s string) bool {
	upper := strings.ToUpper(s)
	for _, v := range slice {
		if strings.ToUpper(v) == upper {
			return true
		}
	}
	return false
}
