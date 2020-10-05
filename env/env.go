package env

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// Init initialize the environment with viper and return keys
func Init(environment string, configPaths ...string) ([]string, error) {
	filePath := path.Dir(os.Args[0])
	_, caller, _, _ := runtime.Caller(1)
	callerPath := path.Dir(caller)

	if len(configPaths) == 0 {
		configPaths = append(configPaths, filePath, callerPath)
	}

	for _, val := range configPaths {
		if path.IsAbs(val) {
			viper.AddConfigPath(val)
		} else {
			viper.AddConfigPath(filePath + "/" + val)
			viper.AddConfigPath(callerPath + "/" + val)
		}
	}

	viper.SetConfigName(environment)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig: %v", err)
	}

	for key, val := range viper.AllSettings() {
		if strings.HasPrefix(key, "env_") {
			os.Setenv(strings.ToUpper(strings.TrimPrefix(key, "env_")), val.(string))
		}
	}

	keys := viper.AllKeys()
	sort.Strings(keys)

	return keys, nil
}
