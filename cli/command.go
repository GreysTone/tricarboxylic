package cli

import (
  "context"
  "flag"
  "fmt"
  "log"
  "os"
  "time"

  "github.com/GreysTone/tricarboxylic/config"
  pb "github.com/GreysTone/tricarboxylic/rpc"
  "github.com/spf13/cobra"
  "google.golang.org/grpc"
)

const (
  TricarbdAddr = "localhost:50101"
  TricarbdPort = "50101"
)

var (
  statusCmd = &cobra.Command{
    Use:		"list",
    Short:	"list current status of tricarb",
    Run: func(cmd *cobra.Command, args[]string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.Status(ctx, &pb.Request{Client: "trictl"})
      if err != nil {
        log.Fatalf("failed to list status: %v\n", err)
      }
      fmt.Print(r.GetMsg())
    },
  }

  versionCmd = &cobra.Command{
    Use:   "version",
    Short: "show version",
    Run: func(cmd *cobra.Command, args []string) {
      fmt.Printf("trictl:\n%v\n\n", config.Version())

      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.Version(ctx, &pb.Request{Client: "trictl"})
      if err != nil {
        log.Fatalf("failed to get version: %v\n", err)
      }
      fmt.Printf("tricarbd:\n%v\n", r.GetMsg())
    },
  }

  setCmd = &cobra.Command{
    Use:   "set",
    Short: "set <sub command>",
    Run: func(cmd *cobra.Command, args []string) {
      if err := cmd.Help(); err != nil {
        os.Exit(0)
      }
    },
  }

  setCIDRCmd = &cobra.Command{
    Use:		"cidr",
    Short:	"set a static CIDR for network",
    Args: 	cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.SetCIDR(ctx, &pb.ConfigRequest{Config: args[0]})
      if err != nil {
        log.Fatalf("failed to set CIDR: %v\n", err)
      }
      if r.GetCode() != 0 {
        fmt.Printf("failed to set CIDR, %v\n", r.GetMsg())
      } else {
        fmt.Printf("set CIDR to: %v\n", args[0])
      }
    },
  }

  setPortCmd = &cobra.Command{
    Use:		"port",
    Short:	"set a static port for network",
    Args: 	cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.SetPort(ctx, &pb.ConfigRequest{Config: args[0]})
      if err != nil {
        log.Fatalf("failed to set port: %v\n", err)
      }
      if r.GetCode() != 0 {
        fmt.Printf("failed to set port, %v\n", r.GetMsg())
      } else {
        fmt.Printf("set port to: %v\n", args[0])
      }
    },
  }

  setNetICCmd = &cobra.Command{
    Use:   "nic",
    Short: "select a physical network interface for network",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.SetNetIC(ctx, &pb.ConfigRequest{Config: args[0]})
      if err != nil {
        log.Fatalf("failed to set physical network interface: %v\n", err)
      }
      if r.GetCode() != 0 {
        fmt.Printf("failed to set physical network interface, %v\n", r.GetMsg())
      } else {
        fmt.Printf("set physical network interface to: %v\n", args[0])
      }
    },
  }

  serverCmd = &cobra.Command{
    Use:		"server",
    Short:	"server start/stop",
    Run: func(cmd *cobra.Command, args []string) {
      if err := cmd.Help(); err != nil {
        os.Exit(0)
      }
    },
  }

  serverStartCmd = &cobra.Command{
    Use:		"start",
    Short:	"start a tricarb server",
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.ServerStart(ctx, &pb.Request{Client: "trictl"})
      if err != nil {
        log.Fatalf("failed to start a tricarb server: %v\n", err)
      }
      if r.GetCode() != 0 {
        log.Fatalf("failed to start a tricarb server: %v\n", r.GetMsg())
      } else {
        fmt.Printf("start a tricarb server on: %v\n", r.GetMsg())
      }
    },
  }

  serverStopCmd = &cobra.Command{
    Use:		"stop",
    Short:	"stop a tricarb server",
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.ServerStop(ctx, &pb.Request{Client: "trictl"})
      if err != nil {
        log.Fatalf("failed to stop the tricarb server: %v\n", err)
      }
      if r.GetCode() != 0 {
        log.Fatalf("failed to stop the tricarb server: %v\n", r.GetMsg())
      } else {
        fmt.Printf("stop the tricarb server\n")
      }
    },
  }

  clientCmd = &cobra.Command{
    Use:		"client",
    Short:	"client attach/detach",
    Run: func(cmd *cobra.Command, args []string) {
      if err := cmd.Help(); err != nil {
        os.Exit(0)
      }
    },
  }

  hostFlag string
  accessCode string

  clientAttachCmd = &cobra.Command{
    Use:		"attach",
    Short:	"attach to a tricarb server",
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.ClientAttach(ctx, &pb.ServerInfo{
        Host:				hostFlag,
        Port:				TricarbdPort,
        AccessCode:	accessCode,
      })
      if err != nil {
        log.Fatalf("failed to attach to tricarb server: %v\n", err)
      }
      if r.GetCode() != 0 {
        log.Fatalf("failed to attach to tricarb server: %v\n", r.GetMsg())
      } else {
        fmt.Printf("attached to server: %v\n", hostFlag)
      }
    },
  }

  clientDetachCmd = &cobra.Command{
    Use:		"detach",
    Short:	"detach from a tricarb server",
    Run: func(cmd *cobra.Command, args []string) {
      conn, err := grpc.Dial(TricarbdAddr, grpc.WithInsecure(), grpc.WithBlock())
      if err != nil {
        log.Fatalf("failed to connect to server: %v\n", err)
      }
      defer conn.Close()
      c := pb.NewTricarbClient(conn)
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
      r, err := c.ClientDetach(ctx, &pb.ServerInfo{
        Host:       hostFlag,
        Port:       TricarbdPort,
        AccessCode: accessCode,
      })
      if err != nil {
        log.Fatalf("failed to detach from tricarb server: %v\n", err)
      }
      if r.GetCode() != 0 {
        log.Fatalf("failed to detach from tricarb server: %v\n", r.GetMsg())
      } else {
        fmt.Printf("detached from server: %v\n", hostFlag)
      }
    },
  }
)

func NewTricarbCtl() * cobra.Command {
  return &cobra.Command{
    Use:   "trictl",
    Short: "trictl COMMAND",
    Run: func (cmd *cobra.Command, args []string) {
      flag.Parse()
      if err := cmd.Help(); err != nil {
        os.Exit(0)
      }
    },
  }
}

func SetupTricarbCtl(cmd * cobra.Command) {
  cmd.AddCommand(statusCmd)
  cmd.AddCommand(versionCmd)

  cmd.AddCommand(setCmd)
  setCmd.AddCommand(setCIDRCmd)
  setCmd.AddCommand(setPortCmd)
  setCmd.AddCommand(setNetICCmd)

  cmd.AddCommand(serverCmd)
  serverCmd.AddCommand(serverStartCmd)
  serverCmd.AddCommand(serverStopCmd)

  cmd.AddCommand(clientCmd)
  clientCmd.AddCommand(clientAttachCmd)
  clientCmd.AddCommand(clientDetachCmd)
  clientAttachCmd.Flags().StringVarP(&hostFlag, "host", "n", "", "tricarb server's host")
  clientAttachCmd.Flags().StringVarP(&accessCode, "access", "a", "", "tricarb server's access code")
  clientDetachCmd.Flags().StringVarP(&hostFlag, "host", "n", "", "tricarb server's host")
  clientDetachCmd.Flags().StringVarP(&accessCode, "access", "a", "", "tricarb server's access code")
}
