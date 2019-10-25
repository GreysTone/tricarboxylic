package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/GreysTone/tricarboxylic/backend"
	"github.com/GreysTone/tricarboxylic/config"
	"github.com/spf13/cobra"
)

var (
	be backend.VpnBackend = nil

	installCmd = &cobra.Command{
		Use:   "install",
		Short: "install <platform>",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			if err := be.Install(args[0]); err != nil {
				println("failed to install")
			}
		},
	}

	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "build <sub command>",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(0)
			}
		},
	}

	buildServerCmd = &cobra.Command{
		Use:   "server",
		Short: "create a new server interface",
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			if err := be.Server(map[string]string{}); err != nil {
				panic(err)
			}
		},
	}

	buildClientCmd = &cobra.Command{
		Use:   "client",
		Short: "create a new client interface",
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			if err := be.Client(map[string]string{}); err != nil {
				panic(err)
			}
		},
	}

	addCmd = &cobra.Command{
		Use:   "add",
		Short: "add a new node to server interface",
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			if err := be.AddNode(map[string]string{}); err != nil {
				panic(err)
			}
		},
	}

	delCmd = &cobra.Command{
		Use:   "del",
		Short: "delete a node from server interface",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			filter := map[string]string{
				"hash": args[0],
			}
			if err := be.DelNode(filter); err != nil {
				panic(err)
			}
		},
	}

	connCmd = &cobra.Command{
		Use:   "connect",
		Short: "connect to server",
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			if err := be.Connect(map[string]string{}); err != nil {
				panic(err)
			}
		},
	}

	disconnCmd = &cobra.Command{
		Use:   "disconnect",
		Short: "disconnect from server",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			be = backend.NewBackend(config.Backend())
			filter := map[string]string{
				"hash": args[0],
			}
			if err := be.Disconnect(filter); err != nil {
				panic(err)
			}
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Version())
		},
	}
)

func NewTricarb() * cobra.Command {
	return &cobra.Command{
		Use:	"tricarb",
		Short:  "tricarb COMMAND",
		Run: func (cmd *cobra.Command, args []string) {
			flag.Parse()
			if err := cmd.Help(); err != nil {
				os.Exit(0)
			}
		},
	}
}

func SetupTricarb(cmd * cobra.Command) {
	cmd.AddCommand(installCmd)
	buildCmd.AddCommand(buildServerCmd)
	buildCmd.AddCommand(buildClientCmd)
	cmd.AddCommand(buildCmd)
	cmd.AddCommand(addCmd)
	cmd.AddCommand(delCmd)
	cmd.AddCommand(connCmd)
	cmd.AddCommand(disconnCmd)
	cmd.AddCommand(versionCmd)
}
