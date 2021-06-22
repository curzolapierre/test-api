#!/bin/bash

FILE='list_excuses.json'

for k in $(jq ' keys | .[]' $FILE); do
    row=$(jq -r ".[$k]" $FILE);

    echo $row | cat - list_excuses_new.json > temp && mv temp list_excuses_new.json
done | column -t -s$'\t'
