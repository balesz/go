package env

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Init initialize the environment with viper
func Init(environment string, configPaths ...string) error {
	_, caller, _, _ := runtime.Caller(1)
	configPath := path.Dir(caller)

	if len(configPaths) == 0 {
		viper.AddConfigPath(path.Dir(os.Args[0]))
		viper.AddConfigPath(configPath)
	} else {
		for _, val := range configPaths {
			if path.IsAbs(val) {
				viper.AddConfigPath(path.Dir(os.Args[0]) + "/" + val)
				viper.AddConfigPath(configPath + "/" + val)
			} else {
				viper.AddConfigPath(val)
			}
		}
	}

	viper.SetConfigName(environment)

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("viper.ReadInConfig: %v", err)
	}

	for key, val := range viper.AllSettings() {
		if strings.HasPrefix(key, "env_") {
			os.Setenv(strings.ToUpper(strings.TrimPrefix(key, "env_")), val.(string))
		}
	}

	return nil
}
