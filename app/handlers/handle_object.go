package handlers

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"github.com/wos-project/wos-core-go/app/models"
	"github.com/wos-project/wos-core-go/app/utils"
)

type reqObject struct {
	ApiVersion string `json:"apiVersion" binding:"required"`
	Metadata   struct {
		Name        string    `json:"name" binding:"required"`
		CreatedAt   time.Time `json:"createdAt" binding:"required"`
		Description string    `json:"description"`
		Owner       struct {
			Id       string `json:"id" binding:"required"`
			Provider string `json:"provider" binding:"required"`
			Extra    string `json:"extra"`
		}
		Privacy    string   `json:"privacy"`
		Visibility string   `json:"visibility"`
		Fidelity   []string `json:"fidelity"`
		Location   struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}
	} `json:"metadata" binding:"required"`
	Kind string      `json:"kind" binding:"required"`
	Spec interface{} `json:"spec" binding:"required"`
}

type specArc struct{} // TODO
type specPin struct{} // TODO
type specPinnedArc struct {
	ArcSelector struct {
		Cid string `json:"cid" binding:"required"`
	} `json:"arcSelector" binding:"required"`
	PinSelector struct {
		Cid string `json:"cid" binding:"required"`
	} `json:"pinSelector" binding:"required"`
}

type respObject struct {
	Cid string `json:"cid"`
}

type reqObjectSearch struct {
	MatchExpressions []struct {
		Key      string   `json:"key" binding:"required"`
		Operator string   `json:"operator" binding:"required"`
		Values   []string `json:"values" binding:"required"`
	} `json:"matchExpressions" binding:"required"`
}

type respObjectSearchItem struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt" binding:"required"`
	Owner       struct {
		Id       string `json:"id" binding:"required"`
		Provider string `json:"provider" binding:"required"`
	}
	CoverImageUri string `json:"coverImageUri"`

	PinLocation struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
}

func (si *respObjectSearchItem) MarshalFromArc(m *models.Arc) {
	si.Name = m.Name
	si.Description = m.Description
	si.CreatedAt = m.CreatedAtInner
	si.Owner.Id = m.OwnerUid
	si.Owner.Provider = m.OwnerProvider
	si.CoverImageUri = m.CoverImageUri
}

func (si *respObjectSearchItem) MarshalFromPin(m *models.Pin) {
	si.Name = m.Name
	si.Description = m.Description
	si.CreatedAt = m.CreatedAtInner
	si.Owner.Id = m.OwnerUid
	si.Owner.Provider = m.OwnerProvider
	si.CoverImageUri = m.CoverImageUri
}

func (si *respObjectSearchItem) MarshalFromPinnedArc(m *ArcAndPin) {
	si.Name = m.Arc.Name
	si.Description = m.Arc.Description
	si.CreatedAt = m.Arc.CreatedAtInner
	si.Owner.Id = m.Arc.OwnerUid
	si.Owner.Provider = m.Arc.OwnerProvider
	si.CoverImageUri = m.Arc.CoverImageUri

	si.PinLocation.Lat = m.Pin.Location.Lat
	si.PinLocation.Lon = m.Pin.Location.Lon
}

type ArcAndPin struct {
	models.Arc
	models.Pin
}

type respObjectSearch struct {
	Results []respObjectSearchItem `json:"results"`
}

type respBatchUploadBegin struct {
	SessionID string `json:"sessionId"`
}

func x(r io.Reader) {

	tarReader := tar.NewReader(r)

	if tarReader != nil {

	}
}

// HandleObjectArchiveUploadMultipart godoc
// @Summary HandleObjectArchiveUploadMultipart uploads a tar archive.  It saves to IPFS, creates thumbs, then saves to S3
// @Accept mpfd
// @Produce json
// @Param App-Key header string true "Application key header"
// @Success 200 object respObject success "CID of object"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 451 {string} error "Cannot expand tar"
// @Failure 500 {string} error "Internal error"
// @Router /object/archive/multipart [post]
func HandleObjectArchiveUploadMultipart(c *gin.Context) {

	// get file from http
	file, err := c.FormFile("file")
	if err != nil {
		glog.Errorf("cannot get file from http %v", err)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// create temp dir
	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		glog.Errorf("cannot create temp dir %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}
	defer os.RemoveAll(tempDirPath)

	// expand tar
	mf, err := file.Open()
	err = utils.ExpandTarReader(mf, tempDirPath)
	if err != nil {
		glog.Errorf("cannot expand tar %v", err)
		c.JSON(451, gin.H{"error": ""})
		return
	}

	// save to ipfs
	cid, err := utils.Ipfs.UploadDirectory(tempDirPath)
	if err != nil {
		glog.Errorf("cannot save to IPFS %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// index
	err = indexObjectFile(cid, path.Join(tempDirPath, viper.GetString("media.indexFilename")))
	if err != nil {
		glog.Errorf("cannot index %v", err)
		c.JSON(500, gin.H{"error": ""})
		//TODO: delete from IPFS
		return
	}

	// create thumbs
	err = utils.CreateThumbnailsInFolder(tempDirPath)
	if err != nil {
		glog.Errorf("cannot create thumbnails %s %v", tempDirPath, err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// save to S3
	err = utils.S3.UploadDirectory(tempDirPath, cid)
	if err != nil {
		glog.Errorf("cannot upload to S3 %s %v", cid, err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	c.JSON(200, respObject{Cid: cid})
}

// HandleObjectArchiveGet godoc
// @Summary HandleObjectArchiveGet gets an Object tar archive
// @Produce json
// @Param cid path string true "content identifier"
// @Success 200 {string} success ""
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 500 {string} error "Internal error"
// @Router /object/archive/{cid} [get]
func HandleObjectArchiveGet(c *gin.Context) {

	cid := c.Param("cid")
	if cid == "" {
		glog.Errorf("cid parameter missing")
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// TODO: cid exists in database?

	// create temp folder
	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		glog.Errorf("cannot create temp dir %s %v", cid, err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// get files from IPFS and save to tmp
	err = utils.Ipfs.DownloadDirectory(tempDirPath, cid)
	if err != nil {
		glog.Errorf("cannot get S3 files %s %v", cid, err)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// create tmp tar file, build it, send back
	tarballFilePath := path.Join(viper.GetString("media.uploadTemp.path"), cid+".tar")
	tarballFile, err := os.Create(tarballFilePath)
	if err != nil {
		glog.Errorf("Could not create tarball file '%s', got error '%s'", tarballFilePath, err.Error())
	}
	defer tarballFile.Close()

	tarWriter := tar.NewWriter(tarballFile)
	defer tarWriter.Close()

	filepath.Walk(tempDirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		pSansTempDirPath := p[len(tempDirPath)+1:]
		err = utils.AddFileToTarWriter(p, pSansTempDirPath, tarWriter)
		if err != nil {
			return fmt.Errorf("Could not add file '%s', to tarball, got error '%v'", pSansTempDirPath, err)
		}

		return nil
	})

	tarWriter.Close()
	tarballFile.Close()

	http.ServeFile(c.Writer, c.Request, tarballFilePath)
}

// HandleObjectIndexPost godoc
// @Summary HandleObjectIndexPost adds an object index to IPFS and S3
// @Accept json
// @Produce json
// @Param App-Key header string true "Application key header"
// @Param json body reqObject required "index only object to create as JSON"
// @Success 200 object respObject success "CID of object"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 451 {string} error "Cannot match arc selector with an arc"
// @Failure 452 {string} error "Cannot match pin selector with a pin"
// @Failure 500 {string} error "Internal error"
// @Router /object/index [post]
func HandleObjectIndexPost(c *gin.Context) {

	jsonData, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// create temp dir
	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		glog.Errorf("cannot create temp dir %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}
	defer os.RemoveAll(tempDirPath)

	// save index file
	err = os.WriteFile(path.Join(tempDirPath, viper.GetString("media.indexFilename")), jsonData, 0644)
	if err != nil {
		glog.Errorf("cannot create index file %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// save to ipfs
	cid, err := utils.Ipfs.UploadDirectory(tempDirPath)
	if err != nil {
		glog.Errorf("cannot save to IPFS %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// index
	err = indexObjectFile(cid, path.Join(tempDirPath, viper.GetString("media.indexFilename")))
	if err != nil {
		glog.Errorf("cannot index %v", err)
		c.JSON(500, gin.H{"error": ""})
		//TODO: delete from IPFS
		return
	}

	// save to S3
	err = utils.S3.UploadDirectory(tempDirPath, cid)
	if err != nil {
		glog.Errorf("cannot upload to S3 %s %v", cid, err)
		c.JSON(500, gin.H{"error": ""})
		//TODO: delete from IPFS
		return
	}

	c.JSON(200, respObject{Cid: cid})
}

// HandleObjectIndexGet godoc
// @Summary HandleObjectIndexGet gets an index at Object cid
// @Accept json
// @Produce json
// @Param cid path string true "object address"
// @Success 200 object respObject success "Object body"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 451 {string} error "Cannot find object"
// @Failure 500 {string} error "Internal error"
// @Router /object/{cid}/index [get]
func HandleObjectIndexGet(c *gin.Context) {

	cid := c.Param("cid")
	if cid == "" {
		glog.Errorf("cid parameter missing")
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// find object, whether is arc, pin, pinned_arc
	var arc models.Arc
	res := models.Db.Where("cid = ?", cid).First(&arc)
	if res.RowsAffected == 1 {
		c.JSON(200, arc.Body)
		return
	}
	var pin models.Pin
	res = models.Db.Where("cid = ?", cid).First(&pin)
	if res.RowsAffected == 1 {
		c.JSON(200, pin.Body)
		return
	}
	var pa models.PinnedArc
	res = models.Db.Where("cid = ?", cid).First(&pa)
	if res.RowsAffected == 1 {
		c.JSON(200, pa.Body)
		return
	}

	c.JSON(451, gin.H{"error": ""})
}

// HandleObjectSearch godoc
// @Summary HandleObjectSearch searches for Objects
// @Accept json
// @Produce json
// @Param json body reqObjectSearch required "search criteria JSON"
// @Success 200 object respObjectSearch success "Array of Objects matching search criteria"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 500 {string} error "Internal error"
// @Router /object/search [get]
func HandleObjectSearch(c *gin.Context) {

	// marshall to struct
	var request reqObjectSearch
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		glog.Errorf("cannot unmarshall object search %v", err)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	if len(request.MatchExpressions) == 0 {
		glog.Errorf("missing search expression")
		c.JSON(400, gin.H{"error": ""})
		return
	}
	if len(request.MatchExpressions[0].Values) == 0 {
		glog.Errorf("missing search expression values")
		c.JSON(400, gin.H{"error": ""})
		return
	}

	var resp respObjectSearch

	// TODO: validate more

	if request.MatchExpressions[0].Key == "name" {

		// TODO: validate key/values

		var arcs []models.Arc
		res := models.Db.Where("name like ?", "%"+request.MatchExpressions[0].Values[0]+"%").Find(&arcs)
		if res.Error != nil {
			glog.Errorf("search query error %v", res.Error)
			c.JSON(500, gin.H{"error": ""})
			return
		}

		resp.Results = make([]respObjectSearchItem, len(arcs))
		for i, a := range arcs {
			resp.Results[i].MarshalFromArc(&a)
		}

	} else if request.MatchExpressions[0].Key == "location" {

		lat, err := strconv.ParseFloat(request.MatchExpressions[0].Values[0], 64)
		if err != nil {
			glog.Errorf("location search, lat conversion error %v", err)
			c.JSON(400, gin.H{"error": ""})
			return
		}

		lon, err := strconv.ParseFloat(request.MatchExpressions[0].Values[1], 64)
		if err != nil {
			glog.Errorf("location search, lon, conversion error %v", err)
			c.JSON(400, gin.H{"error": ""})
			return
		}

		var pas []ArcAndPin
		res := models.Db.
			Joins("JOIN pins p on p.id = pinned_arcs.pin_id").
			Joins("JOIN arcs a on a.id = pinned_arcs.arc_id").
			Select("p.*, a.*").
			Where(fmt.Sprintf("ST_DWithin(p.location, 'SRID=4326;POINT(%f %f)'::geography, %f)", lon, lat, 10.0)).
			Table("pinned_arcs").
			Find(&pas)
		if res.Error != nil {
			glog.Errorf("search query error %v", res.Error)
			c.JSON(500, gin.H{"error": ""})
			return
		}

		resp.Results = make([]respObjectSearchItem, len(pas))
		for i, a := range pas {
			resp.Results[i].MarshalFromPinnedArc(&a)
		}
	}

	c.JSON(200, resp)
}

// indexObjectFile indexes index.json file
func indexObjectFile(cid string, indexPath string) error {

	body, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("cannot read index file %v", err)
	}
	return indexObjectString(cid, string(body))
}

// indexObjectString adds an object index given the string body of the object index
func indexObjectString(cid string, body string) error {

	var request reqObject
	err := json.Unmarshal([]byte(body), &request)
	if err != nil {
		return fmt.Errorf("cannot unmarshall object index post %v", err)
	}

	// determine type and insert into arc, pin, pinned_arc
	switch request.Kind {
	case "arc":
		// add arc
		var arc models.Arc
		arc.Cid = cid
		arc.OwnerUid = request.Metadata.Owner.Id
		arc.OwnerProvider = request.Metadata.Owner.Provider
		arc.Name = request.Metadata.Name
		arc.Description = request.Metadata.Description
		arc.CreatedAtInner = request.Metadata.CreatedAt
		arc.Body.UnmarshalJSON([]byte(body))

		res := models.Db.Save(&arc)
		if res.Error != nil {
			return fmt.Errorf("cannot add arc %v", res.Error)
		}

	case "pin":
		// add pin
		var pin models.Pin
		pin.Cid = cid
		pin.OwnerUid = request.Metadata.Owner.Id
		pin.OwnerProvider = request.Metadata.Owner.Provider
		pin.Name = request.Metadata.Name
		pin.Description = request.Metadata.Description
		pin.CreatedAtInner = request.Metadata.CreatedAt
		pin.Location = models.PointGeo{Lat: request.Metadata.Location.Lat, Lon: request.Metadata.Location.Lon}
		pin.Body.UnmarshalJSON([]byte(body))

		res := models.Db.Save(&pin)
		if res.Error != nil {
			return fmt.Errorf("cannot add pin %v", res.Error)
		}

	case "pinnedArc":

		b, err := json.Marshal(request.Spec)
		if err != nil {
			return fmt.Errorf("cannot marshal spec %v", err)
		}
		var spec specPinnedArc
		json.Unmarshal(b, &spec)

		// find arc
		var arc models.Arc
		res := models.Db.Where("cid = ?", spec.ArcSelector.Cid).First(&arc)
		if res.Error != nil {
			return fmt.Errorf("cannot find arc %s %v", spec.ArcSelector.Cid, res.Error)
		}

		// find pin
		var pin models.Pin
		res = models.Db.Where("cid = ?", spec.PinSelector.Cid).First(&pin)
		if res.Error != nil {
			return fmt.Errorf("cannot find pin %s %v", spec.PinSelector.Cid, res.Error)
		}

		// add pinned arc
		var pa models.PinnedArc
		pa.Cid = cid
		pa.OwnerUid = request.Metadata.Owner.Id
		pa.OwnerProvider = request.Metadata.Owner.Provider
		pa.Name = request.Metadata.Name
		pa.Description = request.Metadata.Description
		pa.CreatedAtInner = request.Metadata.CreatedAt
		pa.Body.UnmarshalJSON([]byte(body))

		pa.ArcId = arc.ID
		pa.PinId = pin.ID

		res = models.Db.Save(&pa)
		if res.Error != nil {
			return fmt.Errorf("cannot add pinned arc %v", res.Error)
		}
	}
	return nil
}

// HandleObjectBatchUploadBegin godoc
// @Summary HandleObjectBatchUploadBegin starts a batch upload for an object
// @Accept json
// @Produce json
// @Success 200 object respBatchUploadBegin success "Batch upload session ID"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 500 {string} error "Internal error"
// @Router /object/batchUpload [post]
func HandleObjectBatchUploadBegin(c *gin.Context) {

	// create temp folder
	tempDirPath, err := os.MkdirTemp(viper.GetString("media.uploadTemp.path"), "")
	if err != nil {
		glog.Errorf("cannot create temp dir %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// create session
	var mu models.MediaUpload
	sid := utils.GenerateBase64Rand()
	mu.SessionID = sid
	mu.Status = models.MediaUploadEnabled
	mu.Path = tempDirPath
	res := models.Db.Save(&mu)
	if res.Error != nil {
		os.RemoveAll(tempDirPath)
		glog.Errorf("cannot add MediaUpload %v", res.Error)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	c.JSON(200, respBatchUploadBegin{SessionID: sid})
}

// HandleObjectBatchUploadMultipart
// @Summary HandleObjectBatchUploadMultipart Uploads multipart files to an Object.  Optional path<n> form field specifies the full path for stored full path.
// @Accept json
// @Produce json
// @Success 200 {string} success ""
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 451 {string} error "Cannot find sessionId"
// @Failure 500 {string} error "Internal error"
// @Router /object/batchUpload/multipart/{sessionId} [post]
func HandleObjectBatchUploadMultipart(c *gin.Context) {

	sessionId := c.Param("sessionId")
	if sessionId == "" {
		glog.Errorf("sessionId parameter missing")
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// verify session
	var mu models.MediaUpload
	res := models.Db.Where("session_id = ?", sessionId).First(&mu)
	if res.Error != nil {
		glog.Errorf("cannot find sessionId for upload mp %s", sessionId)
		c.JSON(451, gin.H{"error": ""})
		return
	}

	// upload files to temp folder using path information
	form, _ := c.MultipartForm()
	files := form.File["file"]
	paths := form.Value

	for i, file := range files {

		tp := path.Join(mu.Path, file.Filename)
		if paths != nil {
			if val, ok := paths[fmt.Sprintf("path%d", i)]; ok {
				if len(val) > 0 {
					p := val[0]
					if strings.Contains(p, "..") {
						glog.Errorf("upload path contains illegal string %s", p)
						continue
					}
					p2 := filepath.Dir(p)
					err := os.MkdirAll(path.Join(mu.Path, p2), 0755)
					if err != nil {
						glog.Errorf("error creating upload directory %s %v", p2, err)
						continue
					}
					tp = path.Join(mu.Path, p)
				}
			}
		}
		err := c.SaveUploadedFile(file, tp)
		if err != nil {
			glog.Errorf("cannot save file uploaded file sessionId=%s %v", sessionId, err)
			c.JSON(500, gin.H{"error": ""})
			return
		}
		glog.Info("stored uploaded file to temp path " + tp)
	}
}

// HandleObjectBatchUploadEnd godoc
// @Summary HandleObjectBatchUploadEnd ends a batch upload and loads files to IPFS and S3
// @Accept mpfd
// @Produce json
// @Success 200 object respObject success "CID of uploaded Object"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 451 {string} error "Cannot file session ID"
// @Failure 500 {string} error "Internal error"
// @Router /object/batchUpload/end/{sessionId} [put]
func HandleObjectBatchUploadEnd(c *gin.Context) {

	sessionId := c.Param("sessionId")
	if sessionId == "" {
		glog.Errorf("sessionId parameter missing")
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// verify session
	var mu models.MediaUpload
	res := models.Db.Where("session_id = ?", sessionId).First(&mu)
	if res.Error != nil {
		glog.Errorf("cannot find sessionId for upload mp %s", sessionId)
		c.JSON(451, gin.H{"error": ""})
		return
	}
	defer os.RemoveAll(mu.Path)

	// save to ipfs
	cid, err := utils.Ipfs.UploadDirectory(mu.Path)
	if err != nil {
		glog.Errorf("cannot save to IPFS %v", err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// index
	err = indexObjectFile(cid, path.Join(mu.Path, viper.GetString("media.indexFilename")))
	if err != nil {
		glog.Errorf("cannot index %v", err)
		c.JSON(500, gin.H{"error": ""})
		//TODO: delete from IPFS
		return
	}

	// create thumbs
	err = utils.CreateThumbnailsInFolder(mu.Path)
	if err != nil {
		glog.Errorf("cannot create thumbnails %s %v", mu.Path, err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	// save to S3
	err = utils.S3.UploadDirectory(mu.Path, cid)
	if err != nil {
		glog.Errorf("cannot upload to S3 %s %v", cid, err)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	c.JSON(200, respObject{Cid: cid})
}
