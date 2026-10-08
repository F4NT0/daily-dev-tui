#!/usr/bin/env bash
set -euo pipefail

path="$ENDPOINT_PATH"
inputs="${INPUTS:-}"
{ [ -z "$inputs" ] || [ "$inputs" = "null" ]; } && inputs='{}'
query=()
[ -n "${DAILY_DEV_TOKEN:-}" ] && echo "::add-mask::$DAILY_DEV_TOKEN"
while IFS=$'\t' read -r k v; do
  [ -z "$v" ] && continue
  [ "$k" = "daily_dev_token" ] && continue
  if [[ "$path" == *"{$k}"* ]]; then
    enc=$(jq -rn --arg v "$v" '$v|@uri')
    path="${path//\{$k\}/$enc}"
  else
    query+=(--data-urlencode "$k=$v")
  fi
done < <(echo "$inputs" | jq -r 'to_entries[] | [.key,.value] | @tsv')

url="https://api.daily.dev/public/v1$path"
code=$(curl -sS -G "$url" "${query[@]}" \
  -H "Authorization: Bearer $DAILY_DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -o response.json -w '%{http_code}')

{
  echo "### GET \`$path\` -> HTTP $code"
  echo '```json'
  (jq . response.json 2>/dev/null || cat response.json) | head -c 60000
  echo
  echo '```'
} >> "$GITHUB_STEP_SUMMARY"

cat response.json
[ "$code" -lt 300 ]
