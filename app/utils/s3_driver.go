package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

var S3 S3_1Driver

// S3_1Driver struct manages S3 uploads and URL generation
type S3_1Driver struct {
	session    *session.Session
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	service    s3iface.S3API
}

// Init initializes AWS S3 stuff
func (s *S3_1Driver) Init() error {
	conf := aws.Config{Region: aws.String(viper.GetString("media.schemes.s3_1.region"))}
	s.session = session.New(&conf)
	s.uploader = s3manager.NewUploader(s.session)
	s.downloader = s3manager.NewDownloader(s.session)
	s.service = s3.New(s.session, &aws.Config{
		Region: aws.String(viper.GetString("media.schemes.s3_1.region")),
	})

	// TODO: handle errors
	return nil
}

// GetFileContentType detects the content type of a file
func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// UploadDirectory uploads all the files in a folder
func (s *S3_1Driver) UploadDirectory(localDirPath string, cid string) error {

	if s.service == nil {
		return fmt.Errorf("S3_1Drive not initialized")
	}

	filepath.Walk(localDirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// remove root folder, separate dir from filename, join with cid
		relativePath := p[len(localDirPath):]
		key := path.Join(cid, relativePath)
		err = s.Upload(p, key)
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

// DownloadDirectory downloads all the files in a folder
func (s *S3_1Driver) DownloadDirectory(localDirPath string, cid string) error {

	if s.service == nil {
		return fmt.Errorf("S3_1Drive not initialized")
	}

	bucket := viper.GetString("media.schemes.s3_1.bucket")

	resp, err := s.service.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String(cid)})
	if err != nil {
		return fmt.Errorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	f := func(item *s3.Object) error {
		//strip cid
		key := *item.Key
		keyWithoutCid := key[len(cid):]
		localFilePath := path.Join(localDirPath, keyWithoutCid)

		//mkdir
		folder, _ := filepath.Split(keyWithoutCid)
		err := os.MkdirAll(path.Join(localDirPath, folder), 0755)
		if err != nil {
			e := fmt.Errorf("Unable to create folder %s, %v", folder, err)
			glog.Error(e)
			return e
		}

		file, err := os.Create(localFilePath)
		defer file.Close()
		if err != nil {
			e := fmt.Errorf("Unable to create file %q, %v", item, err)
			glog.Error(e)
			return e
		}

		_, err = s.downloader.Download(
			file,
			&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    item.Key,
			})
		if err != nil {
			e := fmt.Errorf("Unable to download item %q, %v", item, err)
			glog.Error(e)
			return e
		}

		return nil
	}

	for _, item := range resp.Contents {
		if f(item) != nil {
			break
		}
	}

	return nil
}

// Upload loads a string TO S3, must specify mimeType (contentType)
func (s *S3_1Driver) UploadString(body string, contentType string, folder string, key string) error {

	if s.service == nil {
		return fmt.Errorf("S3_1Drive not initialized")
	}

	if len(body) == 0 {
		err := fmt.Errorf("attempt to upload zero length string to S3")
		glog.Error(err)
		return err

	}
	r := strings.NewReader(body)

	bucket := viper.GetString("media.schemes.s3_1.bucket")
	key2 := path.Join(folder, key)

	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key2),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file bucket %s key %s, %v", bucket, key2, err)
	}
	glog.Infof("uploaded string %s\n", aws.StringValue(&result.Location))

	return nil
}

// Upload loads a file TO S3
func (s *S3_1Driver) Upload(localPath string, key string) error {

	if s.service == nil {
		return fmt.Errorf("S3_1Drive not initialized")
	}

	f, err := os.Open(localPath)
	if err != nil {
		glog.Errorf("failed to open file for upload %s to S3 %v", localPath, err)
		return err
	}
	defer f.Close()

	// get file size, if zero, log it
	fi, err := f.Stat()
	if fi.Size() == 0 {
		err = fmt.Errorf("attempt to upload zero length file to S3 %s", localPath)
		glog.Error(err)
		return err

	}

	contentType, err := GetFileContentType(f)
	if err != nil {
		glog.Errorf("failed to detect content type of file %s to S3 %v", localPath, err)
		return err
	}
	f.Seek(0, 0)

	bucket := viper.GetString("media.schemes.s3_1.bucket")

	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file bucket %s key %s file %s, %v", bucket, key, localPath, err)
	}
	glog.Infof("s3 uploaded file %s to %s\n", localPath, aws.StringValue(&result.Location))

	return nil
}

// Download loads a file FROM S3
func (s *S3_1Driver) Download(localPath string, key string) error {

	if s.service == nil {
		return fmt.Errorf("S3_1Drive not initialized")
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("Unable to open file %q, %v", localPath, err)
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(s.session)

	bucket := viper.GetString("media.schemes.s3_1.bucket")

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	glog.Infof("downloaded bucket %s key %s file %s bytes %d", bucket, key, localPath, numBytes)

	if err != nil {
		return fmt.Errorf("failed to download file %q, %v", localPath, err)
	}

	return nil
}

// CreateFolderNameBase64 creates folder name from a string.  Creates an base64 HMAC given the string.
// Sanitizes by removing all non-alphanumerics
func (s *S3_1Driver) CreateFolderNameBase64(text string) string {

	secret := viper.GetString("media.schemes.s3_1.bucketEncoding.field1.secret")

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(text))

	// Hash and encode to base64
	sha := base64.URLEncoding.EncodeToString(h.Sum(nil))

	reg, _ := regexp.Compile("[^0-9a-zA-Z]+")
	sha2 := reg.ReplaceAllString(sha, "")

	return sha2
}

// GetExpiringURL gets a URL to s3 asset that expires, given folder and filename for config set bucket name
func (s *S3_1Driver) GetExpiringURL(key string) (string, http.Header, error) {

	if s.service == nil {
		return "", nil, fmt.Errorf("S3_1Drive not initialized")
	}

	secs := viper.GetInt("media.schemes.s3_1.expireSecs")

	bucket := viper.GetString("media.schemes.s3_1.bucket")

	sdkReq, _ := s.service.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	url, headers, err := sdkReq.PresignRequest(time.Duration(secs) * time.Second)

	if err != nil {
		return "", nil, fmt.Errorf("failed to get expiring S3 URL for bucket %s key %s, %v", bucket, key, err)
	}

	glog.V(2).Infof("created URL to S3 media %s", url)

	return url, headers, nil
}

func (s *S3_1Driver)GetFileUrl(key string) (string, error) {

	region := viper.GetString("media.schemes.s3_1.region")
	if region == "" {
		return "", errors.New("region not set")
	}
	bucket := viper.GetString("media.schemes.s3_1.bucket")
	if bucket == "" {
		return "", errors.New("bucket not set")
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s/%s", bucket, key), nil
}

