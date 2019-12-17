package daemon

import (
  "context"
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "math/rand"
  "net"
  "os"
  "os/exec"
  "path"
  "strconv"
  "strings"
  "time"

  "github.com/GreysTone/tricarboxylic/backend"
  "github.com/GreysTone/tricarboxylic/config"
  pb "github.com/GreysTone/tricarboxylic/rpc"
  "github.com/GreysTone/tricarboxylic/utils"
  "google.golang.org/grpc"
)

const (
  TyServerMode = "server"
  TyClientMode = "client"

  ConfAccessKey = "access"
  ConfCIDRKey   = "default.cidr"
  ConfPortKey   = "default.port"
  ConfNetICKey  = "default.nic"
)

var (
  accessCode string

  workingMode string
  tricarbCIDR string
  tricarbPort string
  tricarbNetIC string

  addrPool map[uint32]bool
  be backend.VpnBackend
  confPath string
)

func init() {
  addrPool = map[uint32]bool{}
  confPath = path.Join(os.Getenv("HOME"), "wg.conf")

  accessCode = utils.ReadString(ConfAccessKey)
  if accessCode == "" {
    accessCode = utils.GenerateAccessCode(32)
  }

  tricarbCIDR = utils.ReadString(ConfCIDRKey)
  tricarbPort = utils.ReadString(ConfPortKey)
  tricarbNetIC = utils.ReadString(ConfNetICKey)
}

type Server struct {
  pb.UnimplementedTricarbServer
}

func (s *Server) Version(ctx context.Context, in *pb.Request) (*pb.Reply, error) {
  return &pb.Reply{Code: 0, Msg: config.Version()}, nil
}

func (s *Server) Status(ctx context.Context, in *pb.Request) (*pb.Reply, error) {
  if be == nil {
    be = backend.NewBackend(config.Backend())
  }

  conf, err := be.Config()
  if err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to get config"}, err
  }
  return &pb.Reply{Code: 0, Msg: conf}, nil
}

func (s *Server) SetMode(ctx context.Context, in *pb.ConfigRequest) (*pb.Reply, error) {
  return &pb.Reply{Code: 1, Msg: "deprecated"}, nil
}

func (s *Server) SetCIDR(ctx context.Context, in *pb.ConfigRequest) (*pb.Reply, error) {
  if _, _, err := net.ParseCIDR(in.GetConfig()); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to parse the given CIDR"}, err
  }
  utils.UpdateString(ConfCIDRKey, in.GetConfig())
  tricarbCIDR = in.GetConfig()
  return &pb.Reply{Code: 0, Msg: ""}, nil
}

func (s *Server) SetPort(ctx context.Context, in *pb.ConfigRequest) (*pb.Reply, error) {
  i, err := strconv.Atoi(in.GetConfig())
  if err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to parse the given port"}, err
  }
  if i < 10000 || i > 20000 {
    return &pb.Reply{Code: 1, Msg: "invalid range of the given port, 10000-20000"}, err
  }
  utils.UpdateString(ConfPortKey, in.GetConfig())
  tricarbPort = in.GetConfig()
  return &pb.Reply{Code: 0, Msg: ""}, nil
}

func (s *Server) SetNetIC(ctx context.Context, in *pb.ConfigRequest) (*pb.Reply, error) {
  _, err := exec.Command("ifconfig", in.GetConfig()).Output()
  if err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to detect the given network interface card"}, err
  }
  utils.UpdateString(ConfNetICKey, in.GetConfig())
  tricarbNetIC = in.GetConfig()
  return &pb.Reply{Code: 0, Msg: ""}, nil
}

func (s *Server) ServerStart(ctx context.Context, in *pb.Request) (*pb.Reply, error) {
  if accessCode == "" {
    accessCode = utils.GenerateAccessCode(32)
    utils.UpdateString(ConfAccessKey, accessCode)
  }

  var newServerIface = map[string]string{}
  if tricarbPort != "" {
    newServerIface["ListenPort"] = tricarbPort
  } else {
    newServerIface["ListenPort"] = strconv.Itoa(10000 + rand.Intn(9999))
  }
  assignedCIDR, err := NewNetworkCIDR(tricarbCIDR, &addrPool)
  if err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to create network"}, nil
  }
  newServerIface["Address"] = assignedCIDR
  println("check nic", tricarbNetIC)
  newServerIface["LocalEth"] = tricarbNetIC

  if be == nil {
    be = backend.NewBackend(config.Backend())
  }
  if err := be.NewKeyPair(); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to generate key pair"}, err
  }
  if err := be.NewInterface(newServerIface); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to create interface"}, err
  }

  fmt.Printf("Server starting on %v\n", newServerIface["ListenPort"])
  return &pb.Reply{Code: 0, Msg: accessCode}, nil
}

func (s *Server) ServerStop(ctx context.Context, in *pb.Request) (*pb.Reply, error) {
  if err := be.DownInterface(confPath); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to down interface"}, nil
  }
  return &pb.Reply{Code: 0, Msg: ""}, nil
}

func (s *Server) ServerAttach(ctx context.Context, in *pb.PeerInfo) (*pb.AttachReply, error) {
  if accessCode != in.GetAccessCode() {
    return &pb.AttachReply{Status: &pb.Reply{Code: 1, Msg: "invalid access code"}}, nil
  }

  if be == nil {
    return &pb.AttachReply{Status: &pb.Reply{Code: 1, Msg: "no server was started"}}, nil
  }

  var newPeer = map[string]string{}
  newPeer["PublicKey"] = in.GetPeerPublicKey()
  dynamicIp, err := NewDynamicIpUnderCIDR(be, &addrPool)
  if err != nil {
    return &pb.AttachReply{Status: &pb.Reply{Code: 1, Msg: "failed to get dynamic ip address"}}, nil
  }
  newPeer["AllowedIPs"] = dynamicIp+"/32"

  if err := be.AddPeer(newPeer); err != nil {
    return &pb.AttachReply{Status: &pb.Reply{Code: 1, Msg: "failed to attach to client node"}}, err
  }

  if err := dumpConfigAndRestartVirtualTap(be); err != nil {
    return &pb.AttachReply{Status: &pb.Reply{Code: 1, Msg: err.Error()}}, nil
  }
  return &pb.AttachReply{
    Status:        &pb.Reply{Code: 0, Msg: ""},
    AssignedCIDR:  dynamicIp+"/24",
    SrvPublicKey:  be.PublicKey(),
    SrvListenPort: be.Port(),
  }, nil
}

func (s *Server) ClientAttach(ctx context.Context, in *pb.ServerInfo) (*pb.Reply, error) {
  if be == nil {
    be = backend.NewBackend(config.Backend())
  }

  if err := be.NewKeyPair(); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to generate key pair"}, err
  }

  // dynamic ip requesting from server
  conn, err := grpc.Dial(in.GetHost()+":"+in.GetPort(), grpc.WithInsecure(), grpc.WithBlock())
  if err != nil {
    log.Fatalf("failed to connect to server: %v", err)
  }
  defer conn.Close()
  c := pb.NewTricarbClient(conn)

  remoteCtx, cancel := context.WithTimeout(context.Background(), time.Second)
  defer cancel()
  r, err := c.ServerAttach(remoteCtx, &pb.PeerInfo{
    AccessCode: in.GetAccessCode(),
    PeerPublicKey: be.PublicKey(),
  })
  if err != nil {
    log.Fatalf("failed to request to server: %v", err)
  }
  if r.GetStatus().GetCode() != 0 {
    return &pb.Reply{Code: 1, Msg: r.GetStatus().GetMsg()}, nil
  }

  var newClientIface = map[string]string{}
  newClientIface["Address"] = r.GetAssignedCIDR()
  newClientIface["LocalEth"] = tricarbNetIC
  if err := be.NewInterface(newClientIface); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to create interface"}, nil
  }

  var newPeer = map[string]string{}
  newPeer["EndPointIp"] = in.GetHost()
  newPeer["EndPointPort"] = r.GetSrvListenPort()
  newPeer["PublicKey"] = r.GetSrvPublicKey()
  _, ipNet, _ := net.ParseCIDR(r.GetAssignedCIDR())
  newPeer["AllowedIPs"] = ipNet.String()

  if err := be.AddPeer(newPeer); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to attach to server node"}, nil
  }

  if err := dumpConfigAndRestartVirtualTap(be); err != nil {
    return &pb.Reply{Code: 1, Msg: err.Error()}, nil
  }
  return &pb.Reply{Code: 0, Msg: ""}, nil
}

func (s *Server) ClientDetach(ctx context.Context, in *pb.ServerInfo) (*pb.Reply, error) {
  if be == nil {
    be = backend.NewBackend(config.Backend())
  }

  if be.PublicKey() == "" {
    return &pb.Reply{Code: 1, Msg: "no client detected"}, nil
  }

  // dynamic ip requesting from server
  conn, err := grpc.Dial(in.GetHost()+":"+in.GetPort(), grpc.WithInsecure(), grpc.WithBlock())
  if err != nil {
    log.Fatalf("failed to connect to server: %v", err)
  }
  defer conn.Close()
  c := pb.NewTricarbClient(conn)

  remoteCtx, cancel := context.WithTimeout(context.Background(), time.Second)
  defer cancel()
  r, err := c.ServerDetach(remoteCtx, &pb.PeerInfo{
    AccessCode: in.GetAccessCode(),
    PeerPublicKey: be.PublicKey(),
  })
  if err != nil {
    log.Fatalf("failed to request to server: %v", err)
  }
  if r.GetStatus().GetCode() != 0 {
    return &pb.Reply{Code: 1, Msg: r.GetStatus().GetMsg()}, nil
  }

  if err := be.DelPeer(r.GetPeerPublicKey()); err != nil {
    return &pb.Reply{Code: 1, Msg: "failed to detach server node"}, nil
  }

  if err := dumpConfigAndRestartVirtualTap(be); err != nil {
    return &pb.Reply{Code: 1, Msg: err.Error()}, nil
  }
  return &pb.Reply{Code: 0, Msg: ""}, nil

}

func (s *Server) ServerDetach(ctx context.Context, in *pb.PeerInfo) (*pb.DetachReply, error) {
  if accessCode != in.GetAccessCode() {
    return &pb.DetachReply{Status: &pb.Reply{Code: 1, Msg: "invalid access code"}}, nil
  }

  if be == nil {
    return &pb.DetachReply{Status: &pb.Reply{Code: 1, Msg: "no server was started"}}, nil
  }

  if err := be.DelPeer(in.GetPeerPublicKey()); err != nil {
    return &pb.DetachReply{Status: &pb.Reply{Code: 1, Msg: "failed to detach client node"}}, err
  }

  if err := dumpConfigAndRestartVirtualTap(be); err != nil {
    return &pb.DetachReply{Status: &pb.Reply{Code: 1, Msg: err.Error()}}, nil
  }
  return &pb.DetachReply{
    Status:        &pb.Reply{Code: 0, Msg: ""},
    PeerPublicKey:  be.PublicKey(),
  }, nil
}

func NewNetworkCIDR(baseCIDR string, pool *map[uint32]bool) (string, error) {
  networkBits := uint32(0)

  if _, ipNet, err := net.ParseCIDR(baseCIDR); err != nil {
    networkBits = uint32(24)
  } else {
    spNet := strings.Split(ipNet.String(), "/")
    spNetBits, err := strconv.Atoi(spNet[1])
    if err != nil {
      return "", err
    }
    networkBits = uint32(spNetBits)
  }

  mask := uint32(1)
  leading := uint32(1)
  for i := uint32(1); i < networkBits; i++ {
    mask = mask << 1 | 0x1
    leading = leading << 1
  }
  networkIp := rand.Uint32() & mask | leading
  fullIp := networkIp << (32 - networkBits)
  fullIp = fullIp | 0x1       // assigned 0x1 as server IP
  ipCIDR := IpUInt32ToAddr(fullIp)+"/"+strconv.Itoa(int(networkBits))

  (*pool)[1] = false
  for i := uint32(2); i <= (uint32(2) << (32 - networkBits)); i++ {
    (*pool)[i] = true
  }

  return ipCIDR, nil
}

func NewDynamicIpUnderCIDR(be backend.VpnBackend, pool *map[uint32]bool) (string, error) {
  spNet := strings.Split(be.CIDR(), "/")

  bits, _ := strconv.Atoi(spNet[1])
  networkBits := uint32(bits)
  rstHostBits := 32 - networkBits

  networkMask := uint32(1)
  for i := uint32(1); i < networkBits; i++ {
    networkMask = networkMask << 1 | 0x1
  }
  networkMask = networkMask << (32 - networkBits)
  rstHostMask := ^networkMask

  network, err := IpAddrToUInt32(spNet[0])
  if err != nil {
    return "", err
  }
  networkNums := network & networkMask
  peers := be.Peer().([]backend.Peer)
  for _, p := range peers {
    ipAddr := strings.Split(p.AllowedIps, "/")[0]
    ip, err := IpAddrToUInt32(ipAddr)
    if err != nil {
      return "", err
    }
    if (ip & networkMask) != networkNums {
      continue
    }
    (*pool)[ip & rstHostMask] = false
  }

  availableIp := uint32(1)
  for i := 2; i < (2 << rstHostBits); i++ {
    if (*pool)[uint32(i)] == true {
      availableIp = uint32(i)
      break
    }
  }
  if availableIp == 1 {
    return "", errors.New("failed to generate config")
  }

  fullNet, err := IpAddrToUInt32(spNet[0])
  if err != nil {
    return "", err
  }
  assignedIp := fullNet & networkMask | availableIp
  ipAddr := IpUInt32ToAddr(assignedIp)

  return ipAddr, nil
}

func IpUInt32ToAddr(ip uint32) string {
  ipSection := [4]uint8{}
  for i := uint32(4); i > 0; i-- {
    ipSection[i-1] = uint8(ip & 0xFF)
    ip = ip >> 8
  }
  ipAddr := ""
  for i := 0; i < 4; i++ {
    ipAddr += strconv.Itoa(int(ipSection[i]))
    if i != 3 {
      ipAddr += "."
    }
  }
  return ipAddr
}

func IpAddrToUInt32(addr string) (uint32, error) {
  ipSection := strings.Split(addr, ".")
  ip := uint32(0)
  for i := 0; i < 4; i++ {
    sec, err := strconv.Atoi(ipSection[i])
    if err != nil {
      return uint32(0), err
    }
    ip = (ip | uint32(sec))
    if i != 3 {
      ip = ip << 8
    }
 }
 return ip, nil
}

//
func dumpConfigAndRestartVirtualTap(be backend.VpnBackend) error {
  conf, err := be.Config()
  if err != nil {
    return errors.New("failed to generate config")
  }
  if err := ioutil.WriteFile(confPath, []byte(conf), 0600); err != nil {
    fmt.Printf("%v", err)
    return errors.New("failed to generate conf file")
  }
  if err := utils.StdIOCmd("ifconfig", "wg"); err == nil {
    if err := be.DownInterface(confPath); err != nil {
      return errors.New("failed to down interface")
    }
  }
  if err := be.UpInterface(confPath); err != nil {
    return errors.New("failed to up interface")
  }
  return nil
}

