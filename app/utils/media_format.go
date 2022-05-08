package utils

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/disintegration/imaging"
	"github.com/golang/glog"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"github.com/spf13/viper"
)

type ThumbnailSizeSuffix struct {
	MaxSize int
	Suffix  string
}

const (
	MediaFormatBase = 1000 // base used for comparisons

	MediaFormatImage           = 1000 // image base code
	MediaFormatDetailImageJpeg = 1001 // jpeg
	MediaFormatDetailImagePng  = 1002 // png
	MediaFormatDetailImageHeic = 1003 // heic

	MediaFormatAudio          = 2000 // audio base code
	MediaFormatDetailAudioMp3 = 2001 // mp3
	MediaFormatDetailAudioMp4 = 2002 // audio only mp4
	MediaFormatDetailAudioAac = 2003 // audio only aac
	MediaFormatDetailAudioOgg = 2004 // audio only ogg

	MediaFormatVideo          = 3000 // video base code
	MediaFormatDetailVideoMp4 = 3001 // mp4
	MediaFormatDetailVideoMov = 3002 // mov

	MediaFormatMetadata           = 4000
	MediaFormatDetailMetadataJson = 4001
	MediaFormatDetailMetadataYaml = 4002

	Thumb_p100  = "p100"
	Thumb_p1080 = "p1080"
	Size_Actual = "actual"
)

var mapMediaFormat map[string]int
var mapThumbnailName2Size map[string]int

const ThumbnailSuffixDelim = "_"

func init() {
	mapMediaFormat = map[string]int{
		"jpeg": MediaFormatDetailImageJpeg,
		"jpg":  MediaFormatDetailImageJpeg,
		"png":  MediaFormatDetailImagePng,
		"mp3":  MediaFormatDetailAudioMp3,
		"m4a":  MediaFormatDetailAudioMp4,
		"mp4":  MediaFormatDetailVideoMp4,
		"mov":  MediaFormatDetailVideoMov,
		"MOV":  MediaFormatDetailVideoMov,
		"heic": MediaFormatDetailImageHeic,
		"json": MediaFormatDetailMetadataJson,
		"yaml": MediaFormatDetailMetadataYaml,
	}

	mapThumbnailName2Size = map[string]int{
		Thumb_p100:  100,
		Thumb_p1080: 1080,
	}
}

// GetMediaSuffix returns the suffix of a file name
func GetMediaSuffix(path string) (suffix string) {
	dot := strings.LastIndex(path, ".")
	if dot >= 0 && len(path) > dot+1 {
		suffix := path[dot+1:]
		return strings.ToLower(suffix)
	}
	return ""
}

// GetMediaFormatAndMetadata returns media format (image, audio, video), detail (jpeg, png, etc.) and metadata map for a media file
func GetMediaFormatAndMetadata(path string, suffix string) (mediaFormat int, mediaFormatDetail int, metadata map[string]string, err error) {

	mediaFormat, mediaFormatDetail, err = matchSuffixToMediaFormat("placeholder." + suffix)
	if err != nil {
		return 0, 0, nil, err
	}
	switch mediaFormat {
	case MediaFormatImage:
		metadata, err = getImageWidthHeight(path)
	case MediaFormatAudio:
		metadata, err = getMpegMetadata(path)
		metadata2, err := getAudioMediaMetadata(path)
		if err == nil {
			for k, v := range metadata2 {
				metadata[k] = v
			}
		}
	case MediaFormatVideo:
		metadata, err = getMpegMetadata(path)
	}
	if err != nil {
		return 0, 0, map[string]string{}, err
	}
	return mediaFormat, mediaFormatDetail, metadata, nil
}

// matchSuffixToMediaFormat returns mediaFormat (image, audio, video) and mediaFormatDetail (kind of image, etc.) for a file suffix
func matchSuffixToMediaFormat(path string) (mediaFormat int, mediaFormatDetail int, err error) {

	dot := strings.LastIndex(path, ".")
	if dot >= 0 && len(path) > dot+1 {
		suffix := path[dot+1:]

		mediaFormatDetail, ok := mapMediaFormat[suffix]
		if !ok {
			goto nomatch
		}
		mediaFormat := mediaFormatDetail / MediaFormatBase * MediaFormatBase

		return mediaFormat, mediaFormatDetail, nil
	}

nomatch:
	return 0, 0, fmt.Errorf("cannot determine suffix type of %s", path)
}

// getMpegMetadata returns title metadata for mpeg file
func getMpegMetadata(path string) (metadata map[string]string, err error) {

	var m tag.Metadata

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		goto nomatch
	}
	m, err = tag.ReadFrom(f)
	if err != nil {
		goto nomatch
	}
	metadata = make(map[string]string)
	metadata["title"] = m.Title()
	return metadata, nil

nomatch:
	return nil, fmt.Errorf("no metadata tags %s %v", path, err)
}

// getImageMetadata returns EXIF width, height, orientation metadata tags for JPEG
func getImageMetadata(path string) (metadata map[string]string, err error) {

	var x *exif.Exif
	var tag *tiff.Tag

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		goto nomatch
	}
	x, err = exif.Decode(f)
	if err != nil {
		goto nomatch
	}
	metadata = make(map[string]string)
	tag, err = x.Get(exif.ImageWidth)
	if err == nil {
		if tag.String() != "" {
			metadata["width"] = tag.String()
		}
	}
	tag, err = x.Get(exif.ImageLength)
	if err == nil {
		if tag.String() != "" {
			metadata["height"] = tag.String()
		}
	}
	tag, err = x.Get(exif.Orientation)
	if err == nil {
		if tag.String() != "" {
			metadata["orientation"] = tag.String()
		}
	}

	return metadata, nil

nomatch:
	return nil, fmt.Errorf("cannot determine image metadata %s %v", path, err)
}

// getImageWidthHeight returns image width and height
func getImageWidthHeight(path string) (metadata map[string]string, err error) {

	// first try to read exif data
	var img image.Config
	metadata = make(map[string]string)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		goto nomatch
	}
	img, _, err = image.DecodeConfig(f)
	if err == nil {
		w := strconv.Itoa(img.Width)
		h := strconv.Itoa(img.Height)

		m, err := getImageMetadata(path)
		if err == nil {
			o, ok := m["orientation"]
			if ok {
				if len(o) == 1 && (o[0] == '5' || o[0] == '6' || o[0] == '7' || o[0] == '8') {
					w, h = h, w
				}
			}
		}

		metadata["width"] = w
		metadata["height"] = h
		return metadata, nil
	}

nomatch:
	return nil, fmt.Errorf("cannot get image width/height %s %v", path, err)
}

// getAudioMediaMetadata calls ffprobe to retrieve media metadata such as duration
func getAudioMediaMetadata(path string) (map[string]string, error) {

	//TODO: sanitize filename b/c of injection possibilities

	s := viper.GetStringSlice("commands.exec.ffprobe.args")
	s = append(s, path)
	cmd := exec.Command(viper.GetString("commands.exec.ffprobe.command"), s...)
	glog.V(2).Infof("running ffprobe %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	glog.V(2).Infof("ffprobe output %s", string(out))

	meta := make(map[string]string)

	if err != nil {
		return meta, err
	}

	type ffprobe struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	f := ffprobe{}
	err = json.Unmarshal([]byte(out), &f)
	if err != nil {
		return meta, err
	}

	meta["duration"] = f.Format.Duration
	return meta, nil
}

// ChangeSuffix changes the suffix of filename from .from to .to
func ChangeSuffix(filename string, newSuffix string) string {

	dot := strings.LastIndex(filename, ".")
	if dot >= 0 && len(filename) > dot+1 {
		return filename[:dot+1] + newSuffix
	}
	return filename
}

// AddSuffixToBasename adds suffix to base of filename, for example filename.jpg to filename_t1.jpg
func AddSuffixToBasename(fullpath string, suffix string) string {

	dir := path.Dir(fullpath)
	filename := path.Base(fullpath)
	var newFilename string

	dot := strings.LastIndex(filename, ".")
	if dot >= 0 && len(filename) > dot+1 {
		newFilename = filename[:dot] + ThumbnailSuffixDelim + suffix + filename[dot:]
	} else {
		newFilename = filename + ThumbnailSuffixDelim + suffix
	}

	return path.Join(dir, newFilename)
}

// CreateAltSizes creates thumbnails of mediaPath, saves thumbnails, and returns paths and names of uploaded images.
// Thumbnail size is set to max of width or height.
// mediaPath is the path to the file, altName is an alternate name that gets thumbnail suffixes added to it
func CreateAltSizes(mediaPath string, altName string) (imagePathsThumbs []string, altNameThumbs []string, err error) {

	// get suffix and format
	suffix := GetMediaSuffix(mediaPath)
	if err != nil {
		return []string{}, []string{}, fmt.Errorf("cannot get suffix %v %s", err, mediaPath)
	}
	mediaFormat, _, err := matchSuffixToMediaFormat("placeholder." + suffix)
	if err != nil {
		return []string{}, []string{}, fmt.Errorf("cannot get media format %v %s", err, mediaPath)
	}

	// FUTURE: return if not images, we can later expand this to create audio clips
	if mediaFormat != MediaFormatImage {
		return []string{}, []string{}, nil
	}

	src, err := imaging.Open(mediaPath, imaging.AutoOrientation(true))
	if err != nil {
		return []string{}, []string{}, fmt.Errorf("cannot open file for resize " + mediaPath)
	}

	for name, size := range mapThumbnailName2Size {

		maxWidth := 0
		maxHeight := 0
		if src.Bounds().Max.X > src.Bounds().Max.Y {
			maxWidth = size
		} else {
			maxHeight = size
		}
		thumb := imaging.Resize(src, maxWidth, maxHeight, imaging.Lanczos)
		imagePathThumb := AddSuffixToBasename(mediaPath, name)
		err = imaging.Save(thumb, imagePathThumb)
		if err != nil {
			// unwind saving thumbs and don't save partials
			for _, np := range imagePathsThumbs {
				os.Remove(np)
			}
			return []string{}, []string{}, fmt.Errorf("cannot save file when resizing " + imagePathThumb)
		}

		imagePathsThumbs = append(imagePathsThumbs, imagePathThumb)

		altNameThumb := path.Join(AddSuffixToBasename(altName, name))
		altNameThumbs = append(altNameThumbs, altNameThumb)

		glog.Infof("created alternate thumb %s", imagePathThumb)
	}

	return
}

// GetThumbnailName returns thumbnail name given image name and format.  Format could be p100, p1080, etc.
func GetThumbnailName(imageName string, format string) string {
	if format == Size_Actual {
		return imageName
	} else {
		return AddSuffixToBasename(imageName, format)
	}
}

// CreateThumbnailsInFolder creates thumbnails for all images in the folder.
// The thumbnails appear next to the images with _p1080 basename suffix, like this a.jpg => a_p1080.jpg
func CreateThumbnailsInFolder(mediaPath string) error {

	// iterate over media files and create thumbnails
	err := filepath.Walk(mediaPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// find all images and create thumbnails

		suffix := GetMediaSuffix(p)

		mediaFormat, _, _, err := GetMediaFormatAndMetadata(p, suffix)

		if err != nil {
			glog.Warningf("Could not determine media file format %s %v", p, err)
		} else if mediaFormat == MediaFormatImage {
			_, _, err := CreateAltSizes(p, "")
			glog.Warningf("Could not create thumbnails for media file %s %v", p, err)
		}

		return nil
	})

	return err
}
