#!/bin/bash

set -e

if [[ -z "$(ls -A "$HUGOBOT_DB_PATH")" ]];then
    echo "WARNING !! $HUGOBOT_DB_PATH is empty, creating new database !"
fi

if [[ -z "$(ls -A "$WEBSITE_PATH")" ]];then
    echo "you need to mount the website path !"
    exit 1
fi


exec "$@"
