# PDN Gateway (P-GW)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Summary

This project improves the [P-GW CNF][1] developed by *Yoshiyuki Kurauchi*
for testing _go-gtp_ project. The major change of this implementation
is the usage of a external database for sharing IP addresses.

## Local Deployment

This project can be deployed locally using [Vagrant tool][2] which
provisions a Ubuntu Focal Virtual Machine automatically. It's highly
recommended to use the `setup.sh` script of the
[bootstrap-vagrant project][3] to install Vagrant dependencies and
plugins required for the tool. The script supports two Virtualization
providers (Libvirt and VirtualBox).

    curl -fsSL http://bit.ly/initVagrant | PROVIDER=libvirt bash

Once Vagrant is installed, it's possible to deploy the End-to-End
solution  with the following instruction:

    vagrant up

[1]: https://github.com/wmnsk/go-gtp/tree/master/examples/gw-tester/pgw
[2]: https://www.vagrantup.com/
[3]: https://github.com/electrocucaracha/bootstrap-vagrant
