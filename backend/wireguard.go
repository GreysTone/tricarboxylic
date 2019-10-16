package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/GreysTone/tricarboxylic/utils"
)

type WireGuard struct {
	privateKey []byte
	publicKey  []byte
}

const (
	defaultListenPort = "10000"
	defaultNetworkCIDR = "10.0.0.1/24"
	defualtNetworkInterface = "eth0"

	configServerInterface = `[Interface]
ListenPort = RWTH_PORT
Address = RWTH_CIDR
PrivateKey = RWTH_PRV_KEY
PostUp   = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -A FORWARD -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o RWTH_ETH -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o RWTH_ETH -j MASQUERADE
`
	configServerPeer = `[Peer]
PublicKey = RWTH_CLIENT_PUB_KEY
AllowedIPs = RWTH_CIDR
`
	configClientInterface = `[Interface]
Address = RWTH_CIDR
PrivateKey = RWTH_PRV_KEY
PostUp   = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -A FORWARD -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o RWTH_ETH -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o RWTH_ETH -j MASQUERADE
`
	configClientPeer = `[Peer]
PublicKey = RWTH_SERVER_PUB_KEY
AllowedIPs =  RWTH_CIDR
Endpoint = RWTH_SERVER_IP:RWTH_PORT
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
	if err := v.newKeyPair(); err != nil {
		return err
	}

	var replacer = map[string]string{}
	var input string

	// port
	input, err := utils.InputAndCheck(
		"Wireguard Server listen on port:",
		defaultListenPort,
		portValidator)
	replacer["RWTH_PORT"] = input
	if err != nil {
		return err
	}

	// cidr
	input, err = utils.InputAndCheck(
		"Wireguard Server CIDR:",
		defaultNetworkCIDR,
		passingByValidator)
	replacer["RWTH_CIDR"] = input
	if err != nil {
		return err
	}

	// private key
	replacer["RWTH_PRV_KEY"] = string(v.privateKey)

	// network interface
	fmt.Println("Listing all network interfaces:")
	if err := utils.StdIOCmd("ifconfig", "-a", "-s"); err != nil {
		fmt.Println("Failed to list network interfaces")
	}
	input, err = utils.InputAndCheck(
		"Output Nework Interface: (eg. eth0):",
		defualtNetworkInterface,
		passingByValidator)
	replacer["RWTH_ETH"] = input
	if err != nil {
		return err
	}

	ifaceText, err := utils.MakeText(configServerInterface, replacer)
	if err != nil {
		return err
	}

	targetFile := path.Join("wg0.conf")
	if _, err := os.Stat(targetFile); os.IsExist(err) {
		fmt.Println("Found wg0.conf exists, will overwrite")
		if err := os.Remove(targetFile); err != nil {
			panic(err)
		}
	}

	f, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY, 0600)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(ifaceText); err != nil {
		return err
	}

	fmt.Println("Successfully dump WireGuard Interface section")

	return v.upIface("./wg0.conf")
}

func (v *WireGuard) AddNode(config map[string]string) error {
	var replacer = map[string]string{}
	var input string

	// public key
	input, err := utils.InputAndCheck(
		"Wireguard Peer public key:",
		"",
		passingByValidator)
	replacer["RWTH_CLIENT_PUB_KEY"] = input
	if err != nil {
		return err
	}

	input, err = utils.InputAndCheck(
		"Peer allowed CIDR:",
		"",
		passingByValidator)
	replacer["RWTH_CIDR"] = input
	if err != nil {
		return err
	}

	targetFile := path.Join("wg0.conf")
	peerText, err := utils.MakeText(configServerPeer, replacer)

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(peerText); err != nil {
		return err
	}

	fmt.Println("Successfully add new WireGuard Peer section")

	return v.restartIface("./wg0.conf")
}

func (v *WireGuard) Client(config map[string]string) error {
	if err := v.newKeyPair(); err != nil {
		return err
	}
	var replacer = map[string]string{}
	var input string

	// cidr
	input, err := utils.InputAndCheck(
		"Wireguard Client CIDR:",
		defaultNetworkCIDR,
		passingByValidator)
	replacer["RWTH_CIDR"] = input
	if err != nil {
		return err
	}

	// private key
	replacer["RWTH_PRV_KEY"] = string(v.privateKey)

	// network interface
	fmt.Println("Listing all network interfaces:")
	if err := utils.StdIOCmd("ifconfig", "-a", "-s"); err != nil {
		return err
	}
	input, err = utils.InputAndCheck(
		"Output Nework Interface: (eg. eth0):",
		defualtNetworkInterface,
		passingByValidator)
	replacer["RWTH_ETH"] = input
	if err != nil {
		return err
	}

	ifaceText, err := utils.MakeText(configClientInterface, replacer)
	if err != nil {
		return err
	}

	targetFile := path.Join("wg0.conf")
	if _, err := os.Stat(targetFile); os.IsExist(err) {
		fmt.Println("Found wg0.conf exists, will overwrite")
		if err := os.Remove(targetFile); err != nil {
			panic(err)
		}
	}

	f, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY, 0600)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(ifaceText); err != nil {
		return err
	}

	fmt.Println("Successfully dump WireGuard Interface section")

	return nil
}

func (v *WireGuard) Connect(config map[string]string) error {
	var replacer = map[string]string{}
	var input string

	// ip address
	input, err := utils.InputAndCheck(
		"Wireguard Server serving on ip:",
		"",
		passingByValidator)
	replacer["RWTH_SERVER_IP"] = input
	if err != nil {
		return err
	}

	// port
	input, err = utils.InputAndCheck(
		"Wireguard Server serving on port:",
		defaultListenPort,
		portValidator)
	replacer["RWTH_PORT"] = input
	if err != nil {
		return err
	}

	// public key
	input, err = utils.InputAndCheck(
		"Wireguard Peer public key:",
		"",
		passingByValidator)
	replacer["RWTH_SERVER_PUB_KEY"] = input
	if err != nil {
		return err
	}

	input, err = utils.InputAndCheck(
		"Peer allowed CIDR:",
		"",
		passingByValidator)
	replacer["RWTH_CIDR"] = input
	if err != nil {
		return err
	}

	targetFile := path.Join("wg0.conf")
	peerText, err := utils.MakeText(configClientPeer, replacer)

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(peerText); err != nil {
		return err
	}

	fmt.Println("Successfully add new WireGuard Server section")

	return v.upIface("./wg0.conf")

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

func (v *WireGuard) newKeyPair() error {
	privateKey, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return err
	}
	v.privateKey = privateKey

	pubCmd := exec.Command("wg", "pubkey")
	pubCmd.Stdin = strings.NewReader(string(v.privateKey))
	v.publicKey, err = pubCmd.Output()
	if err != nil {
		return err
	}

	fmt.Println("PrivateKey:", string(v.privateKey))
	fmt.Println("PublicKey:", string(v.publicKey))
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
