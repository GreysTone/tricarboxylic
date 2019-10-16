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
cp config.yaml ~
make linux-amd64
```

## Usage
1 Create Server side
  * `tricarb build server`

2 Create Client side and connect to server
  * `tricarb build client`
  * `tricarb connect`

3 Add client at server
  * `tricarb add`

## Tested
  * Passed on AWS