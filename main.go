package main

import (
	"flag"
	"os"

	"github.com/spf13/pflag"

	"github.com/GreysTone/tricarboxylic/cli"
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()
	rootCmd := cli.NewTricarb()
	cli.SetupTricarb(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}