package handlers

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// CleanupOldTempFiles cleans up old temp files in temp upload folder
func CleanupOldTempFiles() {

	files, err := ioutil.ReadDir(viper.GetString("media.uploadTemp.path"))
	if err != nil {
		glog.Errorf("cleanup old temp files error %v", err)
	}
	for _, f := range files {
		secs := viper.GetInt32("media.uploadTemp.maxAgeSecs")
		if time.Now().After(f.ModTime().Add(time.Duration(secs) * time.Second)) {
			if f.IsDir() {
				err = os.RemoveAll(f.Name())
				if err != nil {
					glog.Errorf("cannot remove old directory %s %v", f.Name(), err)
				}
			} else {
				err = os.Remove(f.Name())
				if err != nil {
					glog.Errorf("cannot remove old file %s %v", f.Name(), err)
				}
			}
		}
	}
}
