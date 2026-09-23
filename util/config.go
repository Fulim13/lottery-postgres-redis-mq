package util

import (
	"fmt"
	"path"

	"github.com/spf13/viper"
)

// FileType
const (
	JSON = "json"
	YAML = "yaml"
	ENV  = "env"
)

// Viper parses JSON, TOML, YAML, HCL, INI and ENV files. It can even watch a config file for
// changes (WatchConfig), so new values can be picked up without restarting the process.
func InitViper(dir, file, FileType string) *viper.Viper {
	config := viper.New()
	config.AddConfigPath(dir)      // directory containing config file
	config.SetConfigName(file)     // file name, without the path and without the extension
	config.SetConfigType(FileType) // file type

	if err := config.ReadInConfig(); err != nil {
		// Anything that goes wrong during startup should end the process. The logger isn't up
		// yet at this point, so we can't report it through slog.
		panic(fmt.Errorf("failed to parse config file %s: %s", path.Join(dir, file)+"."+FileType, err))
	}

	return config
}
