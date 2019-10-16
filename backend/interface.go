package backend

const (
	TyWireGuard = "wireguard"
)

type VpnBackend interface {
	Install(string) error
	Uninstall() error

	Server(map[string]string) error
	AddNode(map[string]string) error
	//RemoveNode(string) error

	Client(map[string]string) error
	Connect(map[string]string) error
	//Disconnect(map[string]string) error

	preflight() error
	upIface(i string) error
	downIface(i string) error
	restartIface(i string) error
	// enableIface(i string) error
	// disableIface(i string) error
}

func NewBackend(ty string) VpnBackend {
	switch ty {
	case TyWireGuard:
		return &WireGuard{}
	default:
		println("not supported backend")
		return nil
	}
}