package daemon

import (
  "fmt"
  "testing"
)

func TestIpUInt32ToAddr(t *testing.T) {
  var (
    in       = uint32(3232235876)
    expected = "192.168.1.100"
  )
  actual := IpUInt32ToAddr(in)
  if actual != expected {
    t.Errorf("IpUInt32ToAddr(%v) = %v; expected %v", in, actual, expected)
  }
  fmt.Printf("[SUCC] Translate to: %v\n", actual)
}

func TestIpAddrToUInt32(t *testing.T) {
  var (
    in       = "192.168.1.100"
    expected = uint32(3232235876)
  )
  actual, err := IpAddrToUInt32(in)
  if err != nil {
    t.Errorf("IpAddrToUInt32(%v) got error %v", in, err)
  }
  if actual != expected {
    t.Errorf("IpAddrToUInt32(%v) = %v; expected %v", in, actual, expected)
  }
  fmt.Printf("[SUCC] Translate to: %v\n", actual)
}

func TestNewNetworkCIDR(t *testing.T) {
  pool := map[uint32]bool{}
  cidr, err := NewNetworkCIDR("", &pool)
  if err != nil {
    t.Errorf("In: %v, Got error: %v", "\"\"", err)
  }
  fmt.Printf("[SUCC] Got network: %v\n", cidr)
}

