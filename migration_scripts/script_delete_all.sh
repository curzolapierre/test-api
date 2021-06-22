#!/bin/bash

# Don't forget to add http or https
API_URL="${API_URL:-localhost:8080}"
GUILD_ID="${GUILD_ID:-guidID}"
FILE='list_excuses.json'
BASIC_AUTH=$(echo -n "$USER:$PASSWORD" | base64)

for k in $(jq ' keys | .[]' $FILE); do
    row=$(jq -r ".[$k]" $FILE);
    id=$(jq -r '.id' <<< "$row");
    curl --location --request DELETE "${API_URL}/api/codexcuses/${GUILD_ID}/${id}" \
        --header "Authorization: Basic ${BASIC_AUTH}"
    echo "$id deleted"
done | column -t -s$'\t'
