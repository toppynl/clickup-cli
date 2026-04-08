# patch-v2-spec.jq — Fixes known type mismatches in the ClickUp V2 OpenAPI spec.
#
# The official spec declares several response fields with incorrect types:
#   - time_spent / time_estimate: declared as string|null, API returns integer (ms)
#   - assignees: declared as string[], API returns object[] with {id, username, email, ...}
#   - tags: declared as string[], API returns object[] with {name, tag_fg, tag_bg}
#
# Usage: jq -f patch-v2-spec.jq clickup-v2.json > clickup-v2-patched.json
#
# Reported to ClickUp: https://feedback.clickup.com/public-api

# Helper: patch time_spent and time_estimate in a properties object
def fix_time_fields:
  if .time_spent then
    .time_spent = {"type": ["integer", "null"], "description": "Time spent in milliseconds"}
  else . end
  | if .time_estimate then
    .time_estimate = {"type": ["integer", "null"], "description": "Time estimate in milliseconds"}
  else . end;

# Helper: patch assignees from string[] to object[]
def fix_assignees:
  if .assignees.items.type == "string" then
    .assignees.items = {
      "type": "object",
      "properties": {
        "id": {"type": "integer"},
        "username": {"type": "string"},
        "email": {"type": "string"},
        "color": {"type": ["string", "null"]},
        "initials": {"type": ["string", "null"]},
        "profilePicture": {"type": ["string", "null"]}
      }
    }
  else . end;

# Helper: patch tags from string[] to object[]
def fix_tags:
  if .tags.items.type == "string" then
    .tags.items = {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "tag_fg": {"type": "string"},
        "tag_bg": {"type": "string"}
      }
    }
  else . end;

# Walk all schema properties objects and apply fixes
(.. | objects | select(has("properties")) | .properties) |= (
  fix_time_fields | fix_assignees | fix_tags
)
