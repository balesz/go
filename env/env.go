package env

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// Init initialize the environment with viper
func Init(environment string, configPaths ...string) error {
	filePath := path.Dir(os.Args[0])

	if len(configPaths) == 0 {
		configPaths = append(configPaths, filePath)
	}

	for _, val := range configPaths {
		if path.IsAbs(val) {
			viper.AddConfigPath(val)
		} else {
			viper.AddConfigPath(filePath + "/" + val)
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
