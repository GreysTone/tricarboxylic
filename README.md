# Tricarboxylic


Tricarboxylic (Tricarb) is a simple tools that quickly establishes a Virtual Private Network (VPN) for many cases.

Now *Tricarb* supports:
* wireguard

as its backend.

# Build
```bash
go get -t github.com/spf13/pflag
go get -t github.com/spf13/cobra
go get -t github.com/spf13/viper
go get -t k8s.io/klog
go get -t github.com/mitchellh/mapstructure
cp config.yaml ~
make linux-amd64
```

## Usage
1 [Server] Create a server interface
  * `tricarb build server`

2 [Client] Create a client interface and connect to one server
  * `tricarb build client`
  * `tricarb connect`

3 [Server] Add a client to server interface
  * `tricarb add`

## Tested
  * Passed on AWS

## Example Configuration
* Server Configuration
```
[Interface]
ListenPort = 10000
Address = 10.0.0.1/24
PrivateKey = ${ServerPrivateKey}
PostUp   = ...
PostDown = ...

[Peer]
PublicKey = ${ClientPublicKey}
AllowedIPs = 10.0.0.2/32
```

* Client Configuration
```
[Interface]
Address = 10.0.0.2/24
PrivateKey = ${ClientPrivateKey}
PostUp   = ...
PostDown = ...

[Peer]
PublicKey = ${SeverPublicKey}
AllowedIPs = 10.0.0.0/24
Endpoint = ${ServerIP:10000}
PersistentKeepalive = 10
```

## Todo
  * [ ] Move to [native wg packages](https://github.com/WireGuard/wgctrl-go)