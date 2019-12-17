package backend

import (
  "errors"
  "fmt"
  "os/exec"
  "strings"

  "github.com/mitchellh/mapstructure"

  "github.com/GreysTone/tricarboxylic/config"
  "github.com/GreysTone/tricarboxylic/utils"
)

type KeyPair struct {
  privateKey []byte
  publicKey  []byte
}

type Interface struct {
  ListenPort string
  Address    string
  PrivateKey string
  LocalEth   string
}

type Peer struct {
  PublicKey  string
  AllowedIps string
  EndPointIp   string
  EndPointPort string
}

type WireGuard struct {
  kp       KeyPair
  IfaceSec Interface
  PeersSec []Peer
}

const (
  configServerInterface = `ListenPort = RWTH_PORT
`
  configServerPeer = `PublicKey = RWTH_PUB_KEY
AllowedIPs = RWTH_CIDR
`
  configClientInterface = `Address = RWTH_CIDR
PrivateKey = RWTH_PRV_KEY
PostUp   = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -A FORWARD -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o RWTH_ETH -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o RWTH_ETH -j MASQUERADE

`
  configClientPeer = `Endpoint = RWTH_SERVER_IP:RWTH_PORT
PersistentKeepalive = 10
`
)

// please check https://www.wireguard.com/install/
func (v *WireGuard) Install(platform string) error {
  switch platform {
  case "ubuntu.x86":
  if err := utils.StdIOCmd("sudo", "add-apt-repository", "ppa:wireguard/wireguard"); err != nil {
    return err
  }
  if err := utils.StdIOCmd("sudo", "apt-get", "update"); err != nil {
    return err
  }
  if err := utils.StdIOCmd("sudo", "apt-get", "install", "wireguard"); err != nil {
    return err
  }
  default:
    println("not implemented")
  }
  return nil
}

func (v *WireGuard) Uninstall() error {
  fmt.Println("Not implemented")
  return nil
}

func (v *WireGuard) NewKeyPair() error {
  privateKey, err := exec.Command("wg", "genkey").Output()
  if err != nil {
    return err
  }
  v.kp.privateKey = privateKey

  pubCmd := exec.Command("wg", "pubkey")
  pubCmd.Stdin = strings.NewReader(string(v.kp.privateKey))
  v.kp.publicKey, err = pubCmd.Output()
  if err != nil {
    return err
  }

  fmt.Println("PublicKey:", string(v.kp.publicKey))
  return nil
}

func (v *WireGuard) NewInterface(config map[string]string) error {
  if err := v.loadConfig(); err != nil {
    return err
  }

  if string(v.kp.privateKey) == "" {
    return errors.New("empty keypair")
  }

  var newInterface = Interface{
    ListenPort: config["ListenPort"],
    Address:    config["Address"],
    PrivateKey: strings.TrimSpace(string(v.kp.privateKey)),
    LocalEth:   config["LocalEth"],
  }
  v.IfaceSec = newInterface

  if err := v.saveConfig(); err != nil {
    return err
  }
  return nil
}

func (v *WireGuard) AddPeer(config map[string]string) error {
  if err := v.loadConfig(); err != nil {
    return err
  }

  var newPeer = Peer{
    PublicKey:  config["PublicKey"],
    AllowedIps: config["AllowedIPs"],
  }
  if config["EndPointIp"] != "" {
    newPeer.EndPointIp = config["EndPointIp"]
    newPeer.EndPointPort = config["EndPointPort"]
  }
  v.PeersSec = append(v.PeersSec, newPeer)

  if err := v.saveConfig(); err != nil {
    return err
  }
  return nil
}

func (v *WireGuard) DelPeer(hash string) error {
  if err := v.loadConfig(); err != nil {
    return err
  }

  for i, p := range v.PeersSec {
    if strings.Contains(p.PublicKey, hash) {
      fmt.Println("Delete the following node:")
      fmt.Printf("%v\n", p)
      v.PeersSec = append(v.PeersSec[:i], v.PeersSec[i+1:]...)
      break
    }
  }

  if err := v.saveConfig(); err != nil {
    return err
  }
  return nil
}

func (v *WireGuard) UpInterface(i string) error {
  return utils.StdIOCmd("wg-quick", "up", i)
}

func (v *WireGuard) DownInterface(i string) error {
  return utils.StdIOCmd("wg-quick", "down", i)
}

func (v *WireGuard) Config() (string, error) {
  context := ""

  context += "[Interface]\n"
  ifaceRp := map[string]string{
    "RWTH_PORT":    v.IfaceSec.ListenPort,
    "RWTH_CIDR":    v.IfaceSec.Address,
    "RWTH_PRV_KEY": v.IfaceSec.PrivateKey,
    "RWTH_ETH":     v.IfaceSec.LocalEth,
  }
  if v.IfaceSec.ListenPort != "" {
    if ifaceTextS, err := utils.MakeText(configServerInterface, ifaceRp); err != nil {
      return "", err
    } else {
      context += ifaceTextS
    }
  }
  if ifaceTextC, err := utils.MakeText(configClientInterface, ifaceRp); err != nil {
    return "", err
  } else {
    context += ifaceTextC
  }

  // dump Peer Section
  for _, p := range v.PeersSec {
    context += "[Peer]\n"
    peerRp := map[string]string {
      "RWTH_PUB_KEY":   p.PublicKey,
      "RWTH_CIDR":      p.AllowedIps,
      "RWTH_SERVER_IP": p.EndPointIp,
      "RWTH_PORT":      p.EndPointPort,
    }
    if peerTextS, err := utils.MakeText(configServerPeer, peerRp); err != nil {
      return "", err
    } else {
      context += peerTextS
    }
    if p.EndPointIp != "" {
      if peerTextC, err := utils.MakeText(configClientPeer, peerRp); err != nil {
        return "", err
      } else {
        context += peerTextC
      }
    }
    context += "\n"
  }

  return context, nil
}

func (v *WireGuard) CIDR() string {
  return v.IfaceSec.Address
}

func (v *WireGuard) Port() string {
  return v.IfaceSec.ListenPort
}

func (v *WireGuard) Peer() interface{} {
  return v.PeersSec
}

func (v *WireGuard) PublicKey() string {
  return string(v.kp.publicKey)
}

func (v *WireGuard) loadConfig() error {
  rawIface := config.Iface("wg.iface")
  if err := mapstructure.Decode(rawIface, &v.IfaceSec); err != nil {
    return err
  }
  rawPeers := config.Peers("wg.peers")
  if err := mapstructure.Decode(rawPeers, &v.PeersSec); err != nil {
    return err
  }
  return nil
}

func (v *WireGuard) saveConfig() error {
	iface := map[string]interface{} {
		"ListenPort": v.IfaceSec.ListenPort,
		"Address":    v.IfaceSec.Address,
		"PrivateKey": v.IfaceSec.PrivateKey,
		"LocalEth":   v.IfaceSec.LocalEth,
	}
	config.SubmitIface("wg.iface", iface)
	peers := []interface{}{}
	for _, p := range v.PeersSec {
		peer := map[string]string {
			"PublicKey":  p.PublicKey,
			"AllowedIPs": p.AllowedIps,
		}
		if p.EndPointIp != "" {
			peer["EndPointIp"] = p.EndPointIp
			peer["EndPointPort"] = p.EndPointPort
		}
		peers = append(peers, peer)
	}
	config.SubmitPeers("wg.peers", peers)
	return nil
}
