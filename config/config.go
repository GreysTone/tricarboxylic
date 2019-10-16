package config

import (
"fmt"

"github.com/GreysTone/tricarboxylic/utils"
)

var (
	buildVersion string
	buildTime    string
	buildHash    string
	goVersion    string
)

const (
	backendKey      = "backend"
	defaultBackend  = "wireguard"
)

func Version() string {
	ver := fmt.Sprintf(
		"Version: %s\nBuild Time: %s\nBuild Hash: %s\nGo Version: %s\n",
		buildVersion,
		buildTime,
		buildHash,
		goVersion)
	return ver
}

func Backend() string {
	if ret := utils.ReadString(backendKey); ret != "" {
		return ret
	} else {
		return defaultBackend
	}
}
