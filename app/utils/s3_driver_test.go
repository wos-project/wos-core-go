package utils

import (
	"os"
	"path"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3(t *testing.T) {

	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp/arc", 0755)

	s3 := S3_1Driver{}
	s3.Init()

	// test create folder name
	cid := s3.CreateFolderNameBase64("testing123")
	assert.Equal(t, "6O0b5Yl4JuH0oN2I1tFarxiqsq5qnGwkN53KXRA61K4", cid)

	// test upload single file
	err := s3.Upload("../test/NewportAV.jpg", path.Join(cid, "image.jpg"))
	assert.Nil(t, err)

	// test download single file
	err = s3.Download("/var/tmp/blah.jpg", path.Join(cid, "image.jpg"))
	assert.Nil(t, err)

	// test expiring URL
	u, h, err := s3.GetExpiringURL(path.Join(cid, "image.jpg"))
	assert.Equal(t, "https://worldos.s3.amazonaws.com", u[:32])
	assert.Nil(t, h)
	assert.Nil(t, err)

	resp, err := http.Get(u)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 104945, len(bodyBytes))
	defer resp.Body.Close()

	// test upload string
	err = s3.UploadString("test", "text/plain", "test", "test")
	assert.Nil(t, err)

	// test upload directory
	err = s3.UploadDirectory("../test/object_folder", "testtesthhhjj778877")
	assert.Nil(t, err)

	// test download directory
	err = s3.DownloadDirectory("/var/tmp/mediatmp/arc", "testtesthhhjj778877")
	assert.Nil(t, err)

	// TODO: test there!!
}
