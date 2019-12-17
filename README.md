# Tricarboxylic


Tricarboxylic (Tricarb) is a simple tools that quickly establishes a Virtual Private Network (VPN) for many cases.

Now *Tricarb* supports:
* wireguard

as its backend.

## Build
```bash
go get -t github.com/spf13/pflag
go get -t github.com/spf13/cobra
go get -t github.com/spf13/viper
go get -t k8s.io/klog
go get -t github.com/mitchellh/mapstructure
cp config.yaml ~
make golang-proto
make dmn-nix-amd64
make cli-nix-amd64
```

## Components
Tricarb currently become two parts: `tricarbd` and `trictl`, a daemon process and a command line tool.

## Usage
0 No matter [Server] or [Client] side, run `tricarbd` as daemon process

1.1 [Server] Select a physical network interface card (nic)
  * `trictl set nic eth0` (a typical nic is `eth0`)

1.2 [Server] Start a tricarb server
  * `trictl server start`

2.1 [Client] Select a physical nic for client
  * `trictl set nic eth0` (a typical nic is `eth0`)

2.2 [Client] Attach to a tricarb server
  * `tricarb client attach -n [ip_of_server] -a [access_code]`

## Tested
  * Passed on AWS

## Example
* Server
```bash
$ tricli set nic eth0
> set physical network interface to: eth0

$ tricli server start
> start a tricarb server on: WVYvOCxduSMSURsjYllYbFYZLKbGgidf
```

* Client
```bash
$ tricli set nic eth0
> set physical network interface to: eth0

$ tricli client attach -n 172.31.25.37 -a WVYvOCxduSMSURsjYllYbFYZLKbGgidf
> attached to server: 172.31.25.37
```
