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
if [[ "${DEBUG:-true}" == "true" ]]; then
    set -o xtrace
    export PKG_DEBUG=true
fi

function install_deps {
    pkgs=""
    for pkg in "$@"; do
        if ! command -v "$pkg"; then
            pkgs+=" $pkg"
        fi
    done
    if [ -n "$pkgs" ]; then
        # NOTE: Shorten link -> https://github.com/electrocucaracha/pkg-mgr_scripts
        curl -fsSL http://bit.ly/install_pkg | PKG=$pkgs bash
    fi
}

function exit_trap {
    if [[ "${DEBUG:-true}" == "true" ]]; then
        set +o xtrace
    fi
    printf "CPU usage: "
    grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage " %"}'
    printf "Memory free(Kb): "
    awk -v low="$(grep low /proc/zoneinfo | awk '{k+=$2}END{print k}')" '{a[$1]=$2}  END{ print a["MemFree:"]+a["Active(file):"]+a["Inactive(file):"]+a["SReclaimable:"]-(12*low);}' /proc/meminfo
    if command -v docker; then
        sudo docker ps
    fi
}

trap exit_trap ERR

echo "Running installation process..."
install_deps docker-compose docker make go-lang

sudo docker network create --subnet 10.244.0.0/16 --opt com.docker.network.bridge.name=docker_gwbridge docker_gwbridge
sudo docker swarm init --advertise-addr "${HOST_IP:-$(ip route get 8.8.8.8 | grep "^8." | awk '{ print $7 }')}"
