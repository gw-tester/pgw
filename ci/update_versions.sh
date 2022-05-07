#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c)
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o pipefail
set -o errexit
set -o nounset

function get_version {
    local type="$1"
    local name="$2"
    local version=""
    local attempt_counter=0
    readonly max_attempts=5

    until [ "$version" ]; do
        version=$("_get_latest_$type" "$name")
        if [ "$version" ]; then
            break
        elif [ ${attempt_counter} -eq ${max_attempts} ];then
            echo "Max attempts reached"
            exit 1
        fi
        attempt_counter=$((attempt_counter+1))
        sleep $((attempt_counter*2))
    done

    echo "${version#v}"
}

function _get_latest_docker_tag {
    curl -sfL "https://registry.hub.docker.com/v1/repositories/$1/tags" | python -c 'import json,sys,re;versions=[obj["name"] for obj in json.load(sys.stdin) if bool(re.match("^[0-9.]+$", obj["name"])) and obj["name"].count(".") == 1 ];versions.sort(key = lambda x: [int(y) for y in x.split(".")]);print("\n".join(versions))' | tail -n 1
}

if command -v go > /dev/null; then
    golang_version="$(curl -sL https://golang.org/VERSION?m=text | sed 's/go//;s/\..$//')"
    alpine_version="$(get_version docker_tag alpine)"
    go mod tidy -go="$golang_version"
    sed -i "s|FROM golang:.*|FROM golang:$golang_version-alpine$alpine_version as build|g" Dockerfile
    sed -i "s|FROM alpine:.*|FROM alpine:$alpine_version|g" Dockerfile
fi
