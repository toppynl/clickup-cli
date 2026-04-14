---
title: "clickup chat send"
description: "Auto-generated reference for clickup chat send"
---

Send a message to a ClickUp Chat channel

### Synopsis

Send a message to a ClickUp Chat channel using the v3 API.

The channel ID can be found in the ClickUp Chat URL or via the API.

```
clickup chat send <channel-id> <message> [flags]
```

### Examples

```
  # Send a message to a channel
  clickup chat send 12345abc "Hello team!"

  # Send and get JSON response
  clickup chat send 12345abc "Deploy complete" --json
```

### Options

```
  -h, --help              help for send
      --jq string         Filter JSON output using a jq expression
      --json              Output JSON
  -r, --raw               Output raw strings instead of JSON-encoded (use with --jq)
      --template string   Format JSON output using a Go template
```

### SEE ALSO

* [clickup chat](/clickup-cli/reference/clickup_chat/)	 - Manage ClickUp Chat messages

