package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaFormat(t *testing.T) {
	format, formatDetail, metadata, err := GetMediaFormatAndMetadata("../test/hello.mp3", "mp3")
	assert.Nil(t, err)
	assert.Equal(t, 2000, format)
	assert.Equal(t, 2001, formatDetail)
	assert.Equal(t, "Test Title", metadata["title"])

	format, formatDetail, metadata, err = GetMediaFormatAndMetadata("../test/image.jpg", "jpg")
	assert.Nil(t, err)
	assert.Equal(t, 1000, format)
	assert.Equal(t, 1001, formatDetail)
	assert.Equal(t, "200", metadata["width"])
}

func TestChangeSuffix(t *testing.T) {
	assert.Equal(t, "t", ChangeSuffix("t", "mp4"))
	assert.Equal(t, "t.", ChangeSuffix("t.", "mp4"))
	assert.Equal(t, "t.mp4", ChangeSuffix("t.m4a", "mp4"))
	assert.Equal(t, ".mp4", ChangeSuffix(".m4a", "mp4"))
}

func TestImageThumbs(t *testing.T) {
	assert.Equal(t, "/adf/adf/mmm_1.jpg", AddSuffixToBasename("/adf/adf/mmm.jpg", "1"))
	CopyFile("../test/3024x4032.jpg", "/var/tmp/image.jpg")
	storedPaths, uploadedNames, err := CreateAltSizes("/var/tmp/image.jpg", "john.jpg")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(storedPaths))
	assert.Contains(t, storedPaths, "/var/tmp/image_p100.jpg")
	assert.Contains(t, storedPaths, "/var/tmp/image_p1080.jpg")
	_, err = os.Stat("/var/tmp/image_p100.jpg")
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat("/var/tmp/image_p1080.jpg")
	assert.False(t, os.IsNotExist(err))

	assert.Equal(t, 2, len(uploadedNames))
	assert.Contains(t, uploadedNames, "john_p100.jpg")
	assert.Contains(t, uploadedNames, "john_p1080.jpg")

	m, err := getImageWidthHeight("/var/tmp/image_p100.jpg")
	assert.Equal(t, "75", m["width"])
	assert.Equal(t, "100", m["height"])
	m, err = getImageWidthHeight("/var/tmp/image_p1080.jpg")
	assert.Equal(t, "810", m["width"])
	assert.Equal(t, "1080", m["height"])

	os.Remove("/var/tmp/image.jpg")
	os.Remove("/var/tmp/image_a10.jpg")
	os.Remove("/var/tmp/image_a20.jpg")
}

func TestImageMetadata(t *testing.T) {
	metadata, err := getImageWidthHeight("../test/android_portrait.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "1080", metadata["width"])
	assert.Equal(t, "1920", metadata["height"])

	metadata, err = getImageWidthHeight("../test/NewportAV.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "800", metadata["width"])
	assert.Equal(t, "600", metadata["height"])
}
