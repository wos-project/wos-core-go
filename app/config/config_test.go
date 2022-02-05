package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {

	viper.SetConfigFile("../config.yaml")
	assert.Nil(t, viper.ReadInConfig())
}
