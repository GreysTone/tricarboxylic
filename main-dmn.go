package main

import (
  "fmt"
  "log"
  "net"

  "google.golang.org/grpc"

  "github.com/GreysTone/tricarboxylic/daemon"
  pb "github.com/GreysTone/tricarboxylic/rpc"
)

const (
  port = ":50101"
)

func main() {
  lis, err := net.Listen("tcp", port)
  if err != nil {
    log.Fatalf("failed to listen: %v", err)
  }
  fmt.Printf("Server listening on: %v\n", port)
  s := grpc.NewServer()
  pb.RegisterTricarbServer(s, &daemon.Server{})
  if err := s.Serve(lis); err != nil {
    log.Fatalf("faled to serve: %v", err)
  }
}
