package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	"github.com/wos-project/wos-core-go/app/models"
)

type respContract struct {
	Uid           string `json:"uid"`
	WalletAddr    string `json:"walletAddr"`
	WalletKind    string `json:"walletKind"`
	IpfsCid       string `json:"ipfsCid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	CoverImageUri string `json:"coverImageUri"`
	ContractAddr  string `json:"contractAddr"`
	Location      struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
}

// HandleGetContract godoc
// @Summary HandleGetContract gets a contract's details
// @Produce json
// @Param caddr path string required "Contract address"
// @Success 200 {string} success ""
// @Failure 400 {string} error "Request params wrong"
// @Failure 451 {string} error "Cannot find contract"
// @Failure 500 {string} error "Internal Server Error"
// @Router /contract/{caddr} [get]
func HandleGetContract(c *gin.Context) {

	caddr := c.Param("caddr")
	if caddr == "" {
		glog.Errorf("caddr empty")
		c.JSON(400, gin.H{
			"error": "",
		})
		return
	}

	// get transaction and object details
	var tx models.Transaction
	resp := models.Db.Where("contract_addr = ?", caddr).First(&tx)
	if resp.RowsAffected == 0 {
		c.JSON(451, "")
		return
	}

	// get pinned arc
	var pa models.PinnedArc
	resp = models.Db.Where("cid = ?", tx.IpfsCid).First(&pa)
	if resp.RowsAffected == 0 {
		c.JSON(451, "")
		return
	}

	// get arc
	var arc models.Arc
	resp = models.Db.Where("id = ?", pa.ArcId).First(&arc)
	if resp.RowsAffected == 0 {
		c.JSON(451, "")
		return
	}

	// get pin
	var pin models.Pin
	resp = models.Db.Where("id = ?", pa.PinId).First(&pin)
	if resp.RowsAffected == 0 {
		c.JSON(451, "")
		return
	}

	con := respContract{
		Uid:           tx.Uid,
		WalletAddr:    tx.WalletAddr,
		WalletKind:    tx.WalletKind,
		IpfsCid:       tx.IpfsCid,
		Name:          pa.Name,
		Description:   pa.Description,
		CoverImageUri: arc.CoverImageUri,
		ContractAddr:  tx.ContractAddr,
		Location: struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{
			Lat: pin.Location.Lat,
			Lon: pin.Location.Lon,
		},
	}

	c.JSON(200, con)
}
