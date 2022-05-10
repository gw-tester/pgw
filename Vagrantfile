# -*- mode: ruby -*-
# vi: set ft=ruby :
##############################################################################
# Copyright (c)
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

$no_proxy = ENV['NO_PROXY'] || ENV['no_proxy'] || "127.0.0.1,localhost"
(1..254).each do |i|
  $no_proxy += ",10.0.2.#{i}"
end

Vagrant.configure("2") do |config|
  config.vm.provider :libvirt
  config.vm.provider :virtualbox

  config.vm.box = "generic/ubuntu2004"
  config.vm.box_check_update = false
  config.vm.synced_folder './', '/vagrant'

  config.vm.provision 'shell', privileged: false, inline: <<-SHELL
    set -o errexit

    function exit_trap {
        sudo docker info
        sudo docker ps -a
        make logs
    }

    cd /vagrant/

    # Install dependencies
    ./scripts/install.sh | tee ~/install.log
    source /etc/profile.d/path.sh

    # Deploy GW-Tester services
    trap exit_trap ERR
    IMAGE_VERSION=dev make build | tee ~/build.log
    make deploy | tee ~/deploy.log

    # Wait for services
    attempt_counter=0
    max_attempts=30
    until [ "$(sudo docker ps --filter "name=deployments_*_1*" --format "{{.Names}}" | wc -l)" -gt "6" ]; do
        if [ ${attempt_counter} -eq ${max_attempts} ];then
            echo "Max attempts reached"
            exit_trap
            exit 1
        fi
        attempt_counter=$((attempt_counter+1))
        sleep 10
    done
    sleep 10 # Wait for few client requests

    # Validate GW-Tester communication
    sudo ./scripts/check.sh | tee ~/check.log
  SHELL

  [:virtualbox, :libvirt].each do |provider|
  config.vm.provider provider do |p|
      p.cpus = ENV["CPUS"] || 2
      p.memory = ENV['MEMORY'] || 6144
    end
  end

  config.vm.provider "virtualbox" do |v|
    v.gui = false
  end

  config.vm.provider :libvirt do |v|
    v.random_hostname = true
    v.management_network_address = "10.0.2.0/24"
    v.management_network_name = "administration"
    v.cpu_mode = 'host-passthrough'
  end

  if ENV['http_proxy'] != nil and ENV['https_proxy'] != nil
    if Vagrant.has_plugin?('vagrant-proxyconf')
      config.proxy.http     = ENV['http_proxy'] || ENV['HTTP_PROXY'] || ""
      config.proxy.https    = ENV['https_proxy'] || ENV['HTTPS_PROXY'] || ""
      config.proxy.no_proxy = $no_proxy
      config.proxy.enabled = { docker: false }
    end
  end
end
