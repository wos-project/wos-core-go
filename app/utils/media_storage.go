package utils

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// InitMediaStorage initializes media storage
func InitMediaStorage() {
	path := viper.GetString("media.schemes.localSimple.localPath")
	if path != "" {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			glog.Fatalf("cannot create simple scheme path '%s' %v", path, err)
		}
	}

	S3 = S3_1Driver{}
	err := S3.Init()
	if err != nil {
		glog.Fatalf("cannot init S3 driver %v", err)
	}

	Ipfs = IPFS_Driver{}
	err = Ipfs.Init()
	if err != nil {
		glog.Fatalf("cannot init IPFS driver %v", err)
	}
}