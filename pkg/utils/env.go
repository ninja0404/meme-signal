package utils

import (
	"fmt"
	"os"
	"strings"
)

const (
	ENV string = "ENV"

	ENV_LOCAL string = "LOCAL"
	ENV_DEV   string = "DEV"
	ENV_PROD  string = "PROD"
)

const (
	CONFIG_TYPE string = "CONFIG_TYPE"
	CONFIG_FILE string = "FILE"
	CONFIG_MSE  string = "MSE"

	CONFIG_FILE_PATH string = "CONFIG_FILE_PATH"
)

var envPrefix string

func SetEnvPrefix(prefix string) {
	envPrefix = prefix
}

func GetEnv() string {
	return os.Getenv(envPrefix + ENV)
}

func IsLocalEnv() bool {
	return GetEnv() == ENV_LOCAL
}

func IsDevEnv() bool {
	return GetEnv() == ENV_DEV
}

func IsProdEnv() bool {
	return GetEnv() == ENV_PROD
}

func GetConfigType() string {
	configType := os.Getenv(envPrefix + CONFIG_TYPE)
	if configType == "" {
		return CONFIG_FILE
	}
	return configType
}

func IsFileConfig() bool {
	configType := GetConfigType()
	return configType == "" || strings.ToUpper(configType) == CONFIG_FILE
}

func GetConfigFilePath() string {
	fmt.Println("env config name,", envPrefix+CONFIG_FILE_PATH)
	return os.Getenv(envPrefix + CONFIG_FILE_PATH)
}
