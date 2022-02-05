package utils

import (

	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		glog.Fatalf("Error initializing config, %s", err)
	}
}

func TestMediaConvert(t *testing.T) {

	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp/arc", 0755)

	CopyFile("../test/test.m4a", "/var/tmp/test.m4a")
	err := ConvertM4aMp4("/var/tmp/test.m4a")
	assert.Nil(t, err)

	CopyFile("../test/test.heic", "/var/tmp/test.heic")
	err = ConvertHeicJpg("/var/tmp/test.heic")
	assert.Nil(t, err)
}
