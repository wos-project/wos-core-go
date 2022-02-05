package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpfs(t *testing.T) {

	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp", 0755)

	ipfs := IPFS_Driver{}
	ipfs.Init()

	// upload file
	cid, err := ipfs.UploadFile("../test/NewportAV.jpg", "/")
	assert.Nil(t, err)

	// download
	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp", 0755)
	err = ipfs.DownloadDirectory("/var/tmp/mediatmp", cid)
	assert.Nil(t, err)
	// TODO: test file exists

	// upload directory
	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp/a/media", 0755)
	CopyFile("../test/NewportAV.jpg", "/var/tmp/mediatmp/a/NewportAV.jpg")
	CopyFile("../test/NewportAV.jpg", "/var/tmp/mediatmp/a/media/NewportAV.jpg")
	cid, err = ipfs.UploadDirectory("/var/tmp/mediatmp/a")
	assert.Nil(t, err)

	// download directory
	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp/b", 0755)
	err = ipfs.DownloadDirectory("/var/tmp/mediatmp/b", cid)
	assert.Nil(t, err)
	info, _ := os.Stat("/var/tmp/mediatmp/b")
	assert.True(t, info.IsDir())
	info, _ = os.Stat("/var/tmp/mediatmp/b/NewportAV.jpg")
	assert.Equal(t, "NewportAV.jpg", info.Name())
	info, _ = os.Stat("/var/tmp/mediatmp/b/media")
	assert.True(t, info.IsDir())
	info, _ = os.Stat("/var/tmp/mediatmp/b/media/NewportAV.jpg")
	assert.Equal(t, "NewportAV.jpg", info.Name())
}
