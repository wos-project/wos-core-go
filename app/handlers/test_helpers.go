package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func PerformRequest(r http.Handler, method, relativePath, body string) *httptest.ResponseRecorder {
	return PerformRequestFull(r, method, "/"+viper.GetString("apiVersion")+relativePath, body)
}

func PerformRequestFull(r http.Handler, method, fullpath, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, fullpath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(viper.GetString("auth.apiKey.key"), viper.GetString("auth.apiKey.value"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func UploadFile(r http.Handler, method, p string, localPath string, remotePath string) (*httptest.ResponseRecorder, error) {

	webpath := path.Join("/", viper.GetString("apiVersion"), p)
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	f, err := os.Open(path.Join("..", localPath))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fw, err := w.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, f)
	if err != nil {
		return nil, err
	}

	fw, err = w.CreateFormField("path0")
	if err != nil {
		return nil, err
	}

	_, err = fw.Write([]byte(remotePath))
	if err != nil {
		return nil, err
	}
	w.Close()

	req, err := http.NewRequest(method, webpath, &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set(viper.GetString("auth.apiKey.key"), viper.GetString("auth.apiKey.value"))

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder, nil
}

func AssertFileExists(t *testing.T, p string) {
	info, err := os.Stat(p)
	assert.Nil(t, err)
	assert.False(t, info.IsDir())
}

func AssertDirExists(t *testing.T, p string) {
	info, err := os.Stat(p)
	assert.Nil(t, err)
	assert.True(t, info.IsDir())
}
