#!/bin/bash

# Don't forget to add http or https
API_URL="${API_URL:-localhost:8080}"
GUILD_ID="${GUILD_ID:-guidID}"
FILE='list_excuses.json'
BASIC_AUTH=$(echo -n "$USER:$PASSWORD" | base64)

echo "Posting on ${API_URL}/api/codexcuses/${GUILD_ID}"

# for readability, factored out
args=( -d @- -H "Content-Type: application/json" -H "Authorization: Basic ${BASIC_AUTH}" )

while IFS= read -r value; do
  echo "Read  $value" >&2
  curl "${args[@]}" "${API_URL}/api/codexcuses/${GUILD_ID}" <<<"$value"
done < <(jq -c '.[]' <"$FILE")
