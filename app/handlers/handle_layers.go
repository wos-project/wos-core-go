package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	"github.com/wos-project/wos-core-go/app/models"
)

type respLayer struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

type respLayers struct {
	Layers []respLayer `json:"layers"`
}

// HandleLayersGet godoc
// @Summary Gets list of all layers
// @Security JWT
// @Produce json
// @Success 200 {string} success ""
// @Failure 500 {string} error "Internal error"
// @Router /layers [get]
func HandleLayersGet(c *gin.Context) {
	var layers []respLayer
	res := models.Db.Model(&models.Layer{}).Scan(&layers)
	if res.Error != nil {
		glog.Errorf("error reading layers table %v", res.Error)
		c.JSON(500, gin.H{"error": "Internal Error"})
		return
	}
	c.JSON(200, respLayers{Layers: layers})
}
