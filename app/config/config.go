package config

import (
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// ConfigPath is the path to viper config file
var ConfigPath *string

// InitializeConfiguration loads the config file
func InitializeConfiguration() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*ConfigPath)
	if err := viper.ReadInConfig(); err != nil {
		glog.Fatalf("Error loading config file, %s", err)
	}
}
