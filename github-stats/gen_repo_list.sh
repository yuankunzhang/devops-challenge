#!/bin/bash

languages=( "golang" "python" "java" "c" "c++" )

for l in "${languages[@]}"; do
    curl -s -G https://api.github.com/search/repositories --data-urlencode "sort=stars" --data-urlencode "order=desc" --data-urlencode "per_page=100" --data-urlencode "q=language:${l}" | jq '.items | .[] | .full_name' | tr -d '"'
done
