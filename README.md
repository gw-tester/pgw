# PDN Gateway (P-GW)
[![Go Report Card](https://goreportcard.com/badge/github.com/gw-tester/pgw)](https://goreportcard.com/report/github.com/gw-tester/pgw)
[![GoDoc](https://godoc.org/github.com/gw-tester/pgw?status.svg)](https://godoc.org/github.com/gw-tester/pgw)
[![Docker](https://images.microbadger.com/badges/image/gwtester/pgw.svg)](http://microbadger.com/images/gwtester/pgw)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Super-Linter](https://github.com/gw-tester/pgw/workflows/Lint%20Code%20Base/badge.svg)](https://github.com/marketplace/actions/super-linter)

## Summary

This project improves the [P-GW CNF][1] developed by _Yoshiyuki Kurauchi_
for testing _go-gtp_ project. The major change of this implementation
is the usage of a external database for sharing IP addresses.

### Environment Variables

| Name           | Default       | Description                                             |
|:---------------|:--------------|:--------------------------------------------------------|
| LOG_LEVEL      | info          | Specifies the application log level                     |
| REDIS_URL      |               | Specifies the Connection string for Redis Datastore     |
| REDIS_PASSWORD |               | Specifies the passdor for connecting to Redis Datastore |
| ETCD_URL       |               | Specifies the Connection string for ETCD Datastore      |
| S5U_NETWORK    | 172.25.0.0/24 | Defines the S5 User Plane Network CIDR                  |
| S5C_NETWORK    | 172.25.1.0/24 | Defines the S5 Control Plane Network CIDR               |
| SGI_NIC        | eth2          | Network interface used for SGI connection               |
| SGI_SUBNET     | 10.0.1.0/24   | SGI Subnet                                              |

### Management API

| URL          | Description              |
|:-------------|:-------------------------|
| metrics/     | Prometheus metrics       |
| healthcheck/ | Kubernetes health checks |

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
