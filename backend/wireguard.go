package backend

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
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
	defaultListenPort = "10000"
	defaultNetworkCIDR = "10.0.0.1/24"
	defualtNetworkInterface = "eth0"

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

var (
	passingByValidator = func(i string) bool { return true }
	portValidator = func(i string) bool {
		if _, err := strconv.Atoi(i); err != nil { return false
		} else { return true }
	}
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

func (v *WireGuard) Server(config map[string]string) error {
	var newConfig = map[string]string{}

	if input, err := utils.InputAndCheck("Wireguard Server listen on port:",
		defaultListenPort, portValidator); err != nil {
			return err
	} else {
		newConfig["ListenPort"] = input
	}

	if input, err := utils.InputAndCheck("Wireguard Server CIDR:",
		defaultNetworkCIDR, passingByValidator); err != nil {
			return err
	} else {
		newConfig["Address"] = input
	}

	fmt.Println("Listing all network interfaces:")
	if err := utils.StdIOCmd("ifconfig", "-a", "-s"); err != nil {
		fmt.Println("Failed to list network interfaces")
	}
	if input, err := utils.InputAndCheck("Output Nework Interface: (eg. eth0):",
		defualtNetworkInterface, passingByValidator); err != nil {
			return err
	} else {
		newConfig["LocalEth"] = input
	}

	if err := v.newInterface(newConfig); err != nil {
		return err
	}

	if err := v.dumpConfig(path.Join("wg0.conf")); err != nil {
		return err
	}

	fmt.Println("Successfully dump WireGuard server interface")
	return nil
}

func (v *WireGuard) AddNode(config map[string]string) error {
	var newConfig = map[string]string{}

	if input, err := utils.InputAndCheck("Wireguard Peer public key:",
		"", passingByValidator); err != nil {
			return  err
	} else {
		newConfig["PublicKey"] = input
	}

	if input, err := utils.InputAndCheck("Peer allowed CIDR:",
		"", passingByValidator); err != nil {
			return err
	} else {
		newConfig["AllowedIPs"] = input
	}

	if err := v.addPeer(newConfig); err != nil {
		return err
	}

	if err := v.dumpConfig(path.Join("wg0.conf")); err != nil {
		return err
	}

	if err := v.restartIface("./wg0.conf"); err != nil {
		return err
	}
	fmt.Println("Successfully add a new peer")
	return nil
}

func (v *WireGuard) DelNode(config map[string]string) error {
	if err := v.delPeer(config["hash"]); err != nil {
		return err
	}

	if err := v.dumpConfig(path.Join("wg0.conf")); err != nil {
		return err
	}

	if err := v.restartIface("./wg0.conf"); err != nil {
		return err
	}
	fmt.Println("Successfully delete a peer")
	return nil
}

func (v *WireGuard) Client(config map[string]string) error {
	var newConfig = map[string]string{}

	if input, err := utils.InputAndCheck("Wireguard Client CIDR:",
		defaultNetworkCIDR, passingByValidator); err != nil {
		return err
	} else {
		newConfig["Address"] = input
	}

	fmt.Println("Listing all network interfaces:")
	if err := utils.StdIOCmd("ifconfig", "-a", "-s"); err != nil {
		return err
	}
	if input, err := utils.InputAndCheck("Output Nework Interface: (eg. eth0):",
		defualtNetworkInterface, passingByValidator); err != nil {
			return nil
	} else {
		newConfig["LocalEth"] = input
	}

	if err := v.newInterface(newConfig); err != nil {
		return err
	}

	if err := v.dumpConfig(path.Join("wg0.conf")); err != nil {
		return err
	}

	fmt.Println("Successfully dump WireGuard client interface")
	return nil
}

func (v *WireGuard) Connect(config map[string]string) error {
	var newConfig = map[string]string{}

	if input, err := utils.InputAndCheck("Wireguard Server serving on ip:",
		"", passingByValidator); err != nil {
			return err
	} else {
		newConfig["EndPointIp"] = input
	}

	if input, err := utils.InputAndCheck("Wireguard Server serving on port:",
		defaultListenPort, portValidator); err != nil {
		return err
	} else {
		newConfig["EndPointPort"] = input
	}

	if input, err := utils.InputAndCheck("Wireguard Peer public key:",
		"", passingByValidator); err != nil {
			return err
	} else {
		newConfig["PublicKey"] = input
	}

	if input, err := utils.InputAndCheck("Peer allowed CIDR:",
		"", passingByValidator); err != nil {
			return err
	} else {
		newConfig["AllowedIPs"] = input
	}

	if err := v.addPeer(newConfig); err != nil {
		return err
	}

	if err := v.dumpConfig(path.Join("wg0.conf")); err != nil {
		return err
	}

	if err := v.upIface("./wg0.conf"); err != nil {
		return err
	}
	fmt.Println("Successfully connect to server")
	return nil

}

func (v *WireGuard) Disconnect(config map[string]string) error {
	if err := v.delPeer(config["hash"]); err != nil {
		return err
	}

	if err := v.dumpConfig(path.Join("wg0.conf")); err != nil {
		return err
	}

	if err := v.restartIface("./wg0.conf"); err != nil {
		return err
	}
	fmt.Println("Successfully disconnect from server")
	return nil
}

func (v *WireGuard) preflight() error {
	inspectCmd := exec.Command("wg show")
	inspectCmd.Stderr = os.Stderr
	if err := inspectCmd.Run(); err != nil {
		return err
	}

	inspectCmd = exec.Command("tee", "--help")
	if err := inspectCmd.Run(); err != nil {
		return err
	}

	inspectCmd = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := inspectCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (v *WireGuard) upIface(i string) error {
	return utils.StdIOCmd("wg-quick", "up", i)
}

func (v *WireGuard) downIface(i string) error {
	return utils.StdIOCmd("wg-quick", "down", i)
}

func (v *WireGuard) restartIface(i string) error {
	if err := v.downIface(i); err != nil {
		return err
	}
	return v.upIface(i)
}

func (v *WireGuard) newKeyPair() error {
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

func (v *WireGuard) newInterface(config map[string]string) error {
	if err := v.loadConfig(); err != nil {
		return err
	}

	if err := v.newKeyPair(); err != nil {
		return err
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

func (v *WireGuard) addPeer(config map[string]string) error {
	if err := v.loadConfig(); err != nil {
		return err
	}

	var newPeer = Peer{
		PublicKey:           config["PublicKey"],
		AllowedIps:          config["AllowedIPs"],
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

func (v *WireGuard) delPeer(hash string) error {
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

func (v *WireGuard) dumpConfig(path string) error {
	context := ""

	context += "[Interface]\n"
	ifaceRp := map[string]string {
		"RWTH_PORT":    v.IfaceSec.ListenPort,
		"RWTH_CIDR":    v.IfaceSec.Address,
		"RWTH_PRV_KEY": v.IfaceSec.PrivateKey,
		"RWTH_ETH":     v.IfaceSec.LocalEth,
	}
	if v.IfaceSec.ListenPort != "" {
		if ifaceTextS, err := utils.MakeText(configServerInterface, ifaceRp); err != nil {
			return err
		} else {
			context += ifaceTextS
		}
	}
	if ifaceTextC, err := utils.MakeText(configClientInterface, ifaceRp); err != nil {
		return err
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
			return err
		} else {
			context += peerTextS
		}
		if p.EndPointIp != "" {
			if peerTextC, err := utils.MakeText(configClientPeer, peerRp); err != nil {
				return err
			} else {
				context += peerTextC
			}
		}
		context += "\n"
	}

	if err := ioutil.WriteFile(path, []byte(context),0600); err != nil {
		return err
	}

	return nil
}
