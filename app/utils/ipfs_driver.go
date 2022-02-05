package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	shell "github.com/ipfs/go-ipfs-api"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

var Ipfs IPFS_Driver

// IPFS_Driver struct manages IPFS
type IPFS_Driver struct {
	sh *shell.Shell
}

// Init initializes IPFS shell
func (i *IPFS_Driver) Init() error {

	connString := fmt.Sprintf("%s:%d", viper.GetString("media.schemes.ipfs.host"), viper.GetInt32("media.schemes.ipfs.port"))
	i.sh = shell.NewShell(connString)

	if i.sh == nil {
		err := fmt.Errorf("cannot initialize IPFS %s", connString)
		glog.Error(err)
		return err
	}

	return nil
}

// UploadDirectory uploads all the files in a folder.  Files are stored in IPFS in directory prefixPath.
func (i *IPFS_Driver) UploadDirectory(localDirPath string) (cid string, err error) {

	if i.sh == nil {
		return "", fmt.Errorf("IPFS not initialized")
	}

	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		return "", err
	}

	err = CopyDir(localDirPath, tempDirPath)
	if err != nil {
		e := fmt.Errorf("Cannot copy dir for ipfs %s, %v", tempDirPath, err)
		glog.Error(e)
		return "", e
	}

	cid, err = i.sh.AddDir(tempDirPath)
	if err != nil {
		e := fmt.Errorf("AddDir failed %s, %v", tempDirPath, err)
		glog.Error(e)
		return "", e
	}
	i.sh.Pin(cid)

	glog.Infof("uploaded directory %s\n", cid)

	return cid, err
}

// Upload loads a string with relPath relative path.  Files are stored in IPFS in directory prefixPath.
func (i *IPFS_Driver) UploadString(body string, relPath string) (cid string, err error) {

	if i.sh == nil {
		return "", fmt.Errorf("IPFS_Drive not initialized")
	}

	if len(body) == 0 {
		err := fmt.Errorf("attempt to upload zero length string to IPFS")
		glog.Error(err)
		return "", err

	}

	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		return "", err
	}

	prefixedPathPlusRelPath := path.Join(tempDirPath, filepath.Dir(relPath))
	err = os.MkdirAll(prefixedPathPlusRelPath, 0755)
	if err != nil {
		err := fmt.Errorf("cannot add relpath to prefix %s", relPath)
		glog.Error(err)
		return "", err
	}

	err = os.WriteFile(prefixedPathPlusRelPath, []byte(body), 0644)
	if err != nil {
		return "", err
	}

	cid, err = i.sh.AddDir(tempDirPath)
	if err != nil {
		err := fmt.Errorf("IPFS add string error %v", err)
		glog.Error(err)
		return "", err
	}
	i.sh.Pin(cid)

	glog.Infof("uploaded string %s\n", cid)

	return cid, err
}

// UploadFile loads a file to IPFS with relPath relative path.  Files are stored in IPFS in directory prefixPath.
func (i *IPFS_Driver) UploadFile(localPath string, relPath string) (cid string, err error) {

	if i.sh == nil {
		return "", fmt.Errorf("IPFS_Drive not initialized")
	}

	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		return "", err
	}

	prefixedPathPlusRelPath := path.Join(tempDirPath, filepath.Dir(relPath))
	err = os.MkdirAll(prefixedPathPlusRelPath, 0755)
	if err != nil {
		err := fmt.Errorf("cannot add relpath to prefix %s", relPath)
		glog.Error(err)
		return "", err
	}

	CopyFile(localPath, path.Join(prefixedPathPlusRelPath, filepath.Base(localPath)))

	cid, err = i.sh.AddDir(tempDirPath)
	if err != nil {
		err := fmt.Errorf("IPFS add string error %v", err)
		glog.Error(err)
		return "", err
	}
	i.sh.Pin(cid)

	glog.Infof("uploaded string %s\n", cid)

	return cid, err
}

// DownloadDirectory downloads all the files in a folder identified by CID
func (i *IPFS_Driver) DownloadDirectory(localDirPath string, cid string) error {

	if i.sh == nil {
		return fmt.Errorf("IPFS not initialized")
	}

	// TODO
	err := i.sh.Get(cid, localDirPath)
	if err != nil {
		e := fmt.Errorf("Cannot get IPFS %s, %v", cid, err)
		glog.Error(e)
		return e
	}

	return nil
}

// mkTempPrefixedPath creates temp prefixed path directory
func mkTempPrefixPath() (prefixedPath string, err error) {
	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		e := fmt.Errorf("Cannot MkdirTemp for ipfs %v", err)
		glog.Error(e)
		return "", e
	}
	return tempDirPath, err
}
