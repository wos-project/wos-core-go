package handlers

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"testing"
	"os"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wos-project/wos-core-go/app/config"
	"github.com/wos-project/wos-core-go/app/models"
	"github.com/wos-project/wos-core-go/app/utils"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	config.ConfigPath = flag.String("config", "../config.yaml", "path to YAML config file")
	config.InitializeConfiguration()
	flag.Parse()
	utils.InitMediaStorage()
	models.OpenDatabase()
	models.DropAllTables()
	models.CloseDatabase()
	models.InitializeDatabase()
}

func TestObjects(t *testing.T) {

	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp", 0755)

	router := SetupRouter()
	utils.InitMediaStorage()

	// upload arc
	w := PerformRequest(router, "POST", "/object/index", `{
		"apiVersion": "v1",
		"metadata": {
		  "name": "example-mp4",
		  "createdAt": "2021-12-15T01:01:01Z",
		  "description": "example simple audio mp4",
		  "owner": {
		    "id": "y8240fnweir02shwie8ree0",
		    "provider": "eth",
		    "extra": "id:4558373"
		  },
		  "privacy" : "public",
		  "visibility": "visible",
		  "fidelity": [
		    "phone",
		    "earbuds"
		  ]
		},
		"kind": "arc",
		"spec": {
		  "coverImageUri": "/media/example-cover-image.jpg",
		  "representation": [{
		    "profile": "audio",
		    "mimeType": "audio/mp4",
		    "uri": "/media/hello.mp3"
		  }]
		}
	      }`)
	assert.Equal(t, http.StatusOK, w.Code)
	var obj respObject
	err := json.Unmarshal([]byte(w.Body.String()), &obj)
	assert.True(t, len(obj.Cid) > 0)

	time.Sleep(1)

	// get arc
	w = PerformRequest(router, "GET", "/object/"+obj.Cid+"/index", "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 439, w.Body.Len())

	// upload tar
	w, err = UploadFile(router, "POST", "/object/archive/multipart", "test/arc-mp3.tar", "")
	assert.Equal(t, http.StatusOK, w.Code)
	var objArc respObject
	assert.Nil(t, err)
	err = json.Unmarshal([]byte(w.Body.String()), &objArc)
	assert.True(t, len(objArc.Cid) > 0)

	time.Sleep(1)

	// get tar
	w = PerformRequest(router, "GET", "/object/archive/"+objArc.Cid, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, w.Body.Len() > 50000)

	// get arc index
	w = PerformRequest(router, "GET", fmt.Sprintf("/object/%s/index", objArc.Cid), "")
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &obj)
	assert.True(t, len(obj.Cid) > 0)

	// search
	w = PerformRequest(router, "POST", "/object/search", `{"matchExpressions": [ {"key": "name", "operator": "equal", "values": ["example_mp4"]} ]}`)
	assert.Equal(t, http.StatusOK, w.Code)
	var search respObjectSearch
	err = json.Unmarshal([]byte(w.Body.String()), &search)
	assert.Equal(t, 2, len(search.Results))

	// add pin
	w = PerformRequest(router, "POST", "/object/index", `{
		"apiVersion": "v1",
		"metadata": {
		  "name": "pin",
		  "createdAt": "2021-12-15T01:01:01Z",
		  "description": "example pin",
		  "owner": {
		    "id": "y8240fnweir02shwie8ree0",
		    "provider": "eth",
		    "extra": "id:4558373"
		  },
		  "privacy" : "public",
		  "visibility": "visible",
		  "location": {"lat": 41.5, "lon": -71.5}
		},
		"kind": "pin",
		"spec": {
		}
	      }`)
	assert.Equal(t, http.StatusOK, w.Code)
	var objPin respObject
	err = json.Unmarshal([]byte(w.Body.String()), &objPin)
	assert.True(t, len(obj.Cid) > 0)

	// get pin
	w = PerformRequest(router, "GET", fmt.Sprintf("/object/%s/index", obj.Cid), "")
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &objPin)
	assert.True(t, len(obj.Cid) > 0)

	// add pinned arc
	w = PerformRequest(router, "POST", "/object/index", fmt.Sprintf(`{
		"apiVersion": "v1",
		"metadata": {
		  "name": "pinned-arc",
		  "createdAt": "2021-12-15T01:01:01Z",
		  "description": "pinned arc",
		  "owner": {
		    "id": "y8240fnweir02shwie8ree0",
		    "provider": "eth",
		    "extra": "id:4558373"
		  }
		},
		"kind": "pinnedArc",
		"spec": {
		  "arcSelector": {"cid": "%s"},
		  "pinSelector": {"cid": "%s"}
		}
	      }`, objArc.Cid, objPin.Cid))
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &obj)
	assert.True(t, len(obj.Cid) > 0)

	// get pinned arc
	w = PerformRequest(router, "GET", fmt.Sprintf("/object/%s/index", obj.Cid), "")
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &obj)
	assert.True(t, len(obj.Cid) > 0)

	// search pinned arc
	w = PerformRequest(router, "POST", "/object/search", `{"matchExpressions": [ {"key": "location", "operator": "equal", "values": ["41.5", "-71.5"]} ]}`)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &search)
	assert.Equal(t, 1, len(search.Results))
	assert.Equal(t, "https://worldos.s3.amazonaws.com/QmXN8whBDCKTDcPUBb5VJ1QjbcfJv7uGenQ7jFFoM6qxSv/media/example-cover-image.jpg", search.Results[0].CoverImageUri)
	assert.Equal(t, "QmXN8whBDCKTDcPUBb5VJ1QjbcfJv7uGenQ7jFFoM6qxSv", search.Results[0].IpfsCid)

	// batch upload
	w = PerformRequest(router, "POST", "/object/batchUpload", "")
	assert.Equal(t, http.StatusOK, w.Code)
	var bu respBatchUploadBegin
	err = json.Unmarshal([]byte(w.Body.String()), &bu)
	assert.True(t, len(bu.SessionID) > 0)

	w, err = UploadFile(router, "POST", "/object/batchUpload/multipart/"+bu.SessionID, "test/object_folder/index.json", "/index.json")
	assert.Equal(t, http.StatusOK, w.Code)

	w, err = UploadFile(router, "POST", "/object/batchUpload/multipart/"+bu.SessionID, "test/charlestown1.jpg", "/media/charlestown1.jpg")
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, "PUT", "/object/batchUpload/"+bu.SessionID, "")
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &objPin)
	assert.True(t, len(obj.Cid) > 0)

	// test
	w = PerformRequest(router, "GET", fmt.Sprintf("/object/%s/index", obj.Cid), "")
	assert.Equal(t, http.StatusOK, w.Code)
}
