package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	log "k8s.io/klog"
)

var (
	viper_ = viper.New()
)

func init() {
	viper_.SetConfigName("config")
	configDir := os.Getenv("TRICARB_CONFIG")
	if configDir == "" {
		curUser, err := user.Current()
		if err != nil {
			panic("unable to access current user's home directory")
		}
		configDir = curUser.HomeDir
	}
	viper_.AddConfigPath(configDir)
	viper_.SetConfigType("yaml")

	viper_.AutomaticEnv()
	if err := viper_.ReadInConfig(); err != nil {
		panic(err)
	}
}

func ReadString(key string) string {
	if viper_.IsSet(key) {
		log.Info("load config :: " + key)
		if ret := viper_.Get(key); ret != nil {
			return cast.ToString(ret)
		} else {
			return ""
		}
	}
	return ""
}

func ReadMap(key string) map[string]interface{} {
	if viper_.IsSet(key) {
		log.Info("load config :: " + key)
		if ret := viper_.Get(key); ret != nil {
			return cast.ToStringMap(ret)
		} else {
			return map[string]interface{}{}
		}
	}
	return map[string]interface{}{}
}

func UpdateMap(key string, context map[string]interface{}) {
	viper_.Set(key, context)
	if err := viper_.WriteConfig(); err != nil {
		panic(err)
	}
}

func ReadArray(key string) []interface{} {
	if viper_.IsSet(key) {
		log.Info("load config :: " + key)
		if ret := viper_.Get(key); ret != nil {
			return cast.ToSlice(ret)
		} else {
			return []interface{}{}
		}
	}
	return []interface{}{}
}

func UpdateArray(key string, context []interface{}) {
	viper_.Set(key, context)
	if err := viper_.WriteConfig(); err != nil {
		panic(err)
	}
}

func InputAndCheck(prompt string, defaultValue string, validator func(string) bool) (string, error) {
	var input = ""

	if prompt != "" {
		fmt.Print(prompt + " ")
		if _, err := fmt.Scanln(&input); err != nil {
			return "", err
		}
	}

	if input == "" && defaultValue != "" {
		input = defaultValue
	}
	if validator(input) != true {
		return "", errors.New("validation failed")
	}
	return input, nil
}

func StdIOCmd(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func MakeText(script string, replacer map[string]string) (string, error) {
	content := script

	// replace content
	for key, value := range replacer {
		content = strings.Replace(content, key, strings.TrimSpace(value), -1)
	}

	return content, nil
}
