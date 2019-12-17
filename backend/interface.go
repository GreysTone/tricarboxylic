package backend

const (
  TyWireGuard = "wireguard"
)

type VpnBackend interface {
  Install(string) error
  Uninstall() error

  //Server(map[string]string) error
  //AddNode(map[string]string) error
  //DelNode(map[string]string) error

  //Client(map[string]string) error
  //Connect(map[string]string) error
  //Disconnect(map[string]string) error

  NewKeyPair() error
  NewInterface(map[string]string) error
  AddPeer(map[string]string) error
  DelPeer(string) error
  //preflight() error
  UpInterface(i string) error
  DownInterface(i string) error

  CIDR() string
  Port() string
  Peer() interface{}
  PublicKey() string
  Config() (string, error)
  //restartIface(i string) error
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
