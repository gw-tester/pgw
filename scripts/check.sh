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

function info {
    _print_msg "INFO" "$1"
}

function error {
    _print_msg "ERROR" "$1"
    exit 1
}

function _print_msg {
    echo "$(date +%H:%M:%S) - $1: $2"
}

function assert_non_empty {
    local docker_ps=$1
    input=$(docker logs "$docker_ps" 2>&1)

    if [ -z "$input" ]; then
        error "Empty input value"
    fi
}

function assert_equals {
    local input=$1
    local expected=$2

    if [[ "$input" != "$expected" ]]; then
        error "Got $input expected $expected"
    fi
}

function assert_contains {
    local docker_ps=$1
    local expected=$2
    input=$(docker logs "$docker_ps" 2>&1)

    if ! echo "$input" | grep -q "$expected"; then
        error "Got $input expected $expected"
    fi
}

info "Getting Docker process names"
external_client=$(docker ps --filter "name=deployments_external_client_1*" --format "{{.Names}}")
http_server=$(docker ps --filter "name=deployments_http_server_1*" --format "{{.Names}}")
pgw=$(docker ps --filter "name=deployments_pgw_1*" --format "{{.Names}}")
pgw_ipv4_address="$(docker exec "$pgw" ip a s eth4 | grep -oE 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2)"
info "P-GW IPv4 Management Address (bridge network): $pgw_ipv4_address"

info "Validating non-empty logs"
assert_non_empty "$http_server"
assert_non_empty "$pgw"

info "Validating P-GW readiness"
assert_equals "$(curl -s "$pgw_ipv4_address:8080/healthcheck" | jq -r '.status')" "ok"
assert_equals "$(curl -s "$pgw_ipv4_address:8080/healthcheck" | jq -r '.details["datastore-check"].status')" "ok"
assert_equals "$(curl -s "$pgw_ipv4_address:8080/healthcheck" | jq -r '.details["main-check"].status')" "ok"

info "Validating P-GW metrics"
assert_contains "$(curl -s "$pgw_ipv4_address:8080/metrics")" "sessions_created_total"
assert_contains "$(curl -s "$pgw_ipv4_address:8080/metrics") | awk '/^sessions_created_total/{ print $2}'" "1"

info "Validating that services have started"
assert_contains "$http_server" "resuming normal operations"
assert_contains "$pgw" "Started serving S5-C"
assert_contains "$pgw" "Started serving S5-U"
assert_contains "$pgw" "P-GW server has started"

info "Validating session responses"
assert_contains "$pgw" "Create Session Request"
assert_contains "$http_server" '"GET / HTTP/1.1" 200 45'
assert_non_empty "$external_client"
assert_contains "$external_client" "It works!"
