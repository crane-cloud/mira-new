#!/bin/bash
set -e

mkdir -p /usr/local/crane/{git,zip,blobs}

if [ -n "$DOCKERHUB_USERNAME" ] && [ -n "$DOCKERHUB_TOKEN" ]; then
    echo "$DOCKERHUB_TOKEN" | docker login -u "$DOCKERHUB_USERNAME" --password-stdin
fi

exec air -- image-builder
