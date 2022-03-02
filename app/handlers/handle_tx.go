package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"github.com/wos-project/wos-core-go/app/models"
)

type reqTransactionEnqueue struct {
	Kind string                 `json:"kind"`
	Spec map[string]interface{} `json:"spec"`
}

type reqTransactionEnqueueSpecAirdropErc721 struct {
	Uid         string `json:"uid"`
	WalletAddr  string `json:"walletAddr"`
	WalletKind  string `json:"walletKind"`
	IpfsCid     string `json:"ipfsCid"`
	CallbackUri string `json:"callbackUri"`
}

type reqTransactionEnqueueSpecAirdropErc20 struct {
	Uid         string `json:"uid"`
	WalletAddr  string `json:"walletAddr"`
	WalletKind  string `json:"walletKind"`
	Quantity    int    `json:"tokenQuantity"`
	CallbackUri string `json:"callbackUri"`
}

type respTransactionQueuedItem struct {
	Kind string      `json:"kind"`
	Spec interface{} `json:"spec"`
}

type respTransactionQueuedItemAirdropErc721 struct {
	Uid         string `json:"uid"`
	CallbackUri string `json:"callbackUri"`
	WalletAddr  string `json:"walletAddr"`
	WalletKind  string `json:"walletKind"`
	IpfsCid     string `json:"ipfsCid"`
}

type respTransactionQueuedItemAirdropErc20 struct {
	Uid         string `json:"uid"`
	CallbackUri string `json:"callbackUri"`
	WalletAddr  string `json:"walletAddr"`
	WalletKind  string `json:"walletKind"`
	Quantity    int    `json:"tokenQuantity"`
}

type reqTransactionQueuedItemCallback struct {
	Uid           string `json:"uid"`
	TxId          string `json:"txId"`
	ContractAddr  string `json:"contractAddr"`
	Status        string `json:"status"`
	IpfsCid       string `json:"ipfsCid"`
	Cost          string `json:"cost"`
	TokenQuantity int    `json:"tokenQuantity"`
}

// HandleTransactionEnqueue godoc
// @Summary HandleTransactionEnqueue enqueues a transaction
// @Accept mpfd
// @Produce json
// @Param App-Key header string true "Application key header"
// @Param json body reqTransactionEnqueue required "transaction details"
// @Success 200 {string} success ""
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 500 {string} error "Internal Server Error"
// @Router /transaction/enqueue [post]
func HandleTransactionEnqueue(c *gin.Context) {

	var request reqTransactionEnqueue
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		glog.Errorf("cannot unmarshall object tx %v", err)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	b, err := json.Marshal(request.Spec)
	if err != nil {
		glog.Errorf("cannot unmarshall object tx %v", err)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	var tx models.Transaction

	if request.Kind == "erc20" {
		var erc20 reqTransactionEnqueueSpecAirdropErc20
		err = json.Unmarshal(b, &erc20)
		if err != nil {
			glog.Errorf("cannot unmarshall object tx %v", err)
			c.JSON(400, gin.H{"error": ""})
			return
		}
		tx = models.Transaction{
			Uid:           erc20.Uid,
			WalletAddr:    erc20.WalletAddr,
			WalletKind:    erc20.WalletKind,
			TokenQuantity: erc20.Quantity,
			CallbackUri:   erc20.CallbackUri,
		}
	} else if request.Kind == "erc721" {
		var erc721 reqTransactionEnqueueSpecAirdropErc721
		err = json.Unmarshal(b, &erc721)
		if err != nil {
			glog.Errorf("cannot unmarshall object tx %v", err)
			c.JSON(400, gin.H{"error": ""})
			return
		}
		tx = models.Transaction{
			Uid:         erc721.Uid,
			WalletAddr:  erc721.WalletAddr,
			WalletKind:  erc721.WalletKind,
			IpfsCid:     erc721.IpfsCid,
			CallbackUri: erc721.CallbackUri,
		}
	} else {
		glog.Errorf("tx kind not supported %s", request.Kind)
		c.JSON(400, gin.H{"error": ""})
		return
	}
	resp := models.Db.Save(&tx)
	if resp.Error != nil {
		glog.Errorf("cannot save transaction %v", resp.Error)
		c.JSON(500, gin.H{"error": ""})
		return
	}
}

// HandleTransactionQueueGet godoc
// @Summary HandleTransactionQueueGet gets an item from the transaction queue for the transactor
// @Accept mpfd
// @Produce json
// @Param App-Key header string true "Application key header"
// @Success 200 object respTransaction success "transaction details"
// @Success 201 {string} success "no transactions"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 500 {string} error "Internal Server Error"
// @Router /transaction/queue [get]
func HandleTransactionQueueGet(c *gin.Context) {

	var tx models.Transaction
	resp := models.Db.Where("status = 'pending'").Order("created_at").First(&tx)
	if resp.RowsAffected == 0 {
		c.JSON(201, "")
		return
	}

	item := respTransactionQueuedItem{
		Kind: tx.Kind,
	}
	if tx.Kind == "erc721" {

		erc721 := respTransactionQueuedItemAirdropErc721{
			Uid:        tx.Uid,
			WalletAddr: tx.WalletAddr,
			WalletKind: tx.WalletKind,
			IpfsCid:    tx.IpfsCid,
		}

		item.Spec = &erc721

	} else if tx.Kind == "erc20" {

		erc20 := respTransactionQueuedItemAirdropErc20{
			Uid:        tx.Uid,
			WalletAddr: tx.WalletAddr,
			WalletKind: tx.WalletKind,
			Quantity:   tx.TokenQuantity,
		}
		item.Spec = &erc20

	} else {
		glog.Errorf("tx kind not supported %s", tx.Kind)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	// mark transaction as in-process
	tx.Status = models.TRANSACTION_STATUS_PENDING
	resp = models.Db.Save(&tx)
	if resp.Error != nil {
		glog.Errorf("cannot save transaction %v", resp.Error)
		c.JSON(500, gin.H{"error": ""})
		return
	}

	c.JSON(200, item)
}

// HandleTransactionQueueCallback godoc
// @Summary HandleTransactionQueueCallback handles queued callbacks from the transactor
// @Accept mpfd
// @Produce json
// @Param App-Key header string true "Application key header"
// @Success 200 object respObject success "CID of object"
// @Failure 400 {string} error "Request params wrong"
// @Failure 401 {string} error "Unauthorized"
// @Failure 451 {string} error "Cannot find transaction matching UID"
// @Failure 500 {string} error "Internal Server Error"
// @Router /transaction/callback [post]
func HandleTransactionQueueCallback(c *gin.Context) {

	var request reqTransactionQueuedItemCallback
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		glog.Errorf("cannot unmarshall tx callback %v", err)
		c.JSON(400, gin.H{"error": ""})
		return
	}

	var tx models.Transaction
	resp := models.Db.Where("uid = ?", request.Uid).First(&tx)
	if resp.RowsAffected == 0 {
		c.JSON(451, "")
		return
	}
	tx.Cost = request.Cost
	tx.IpfsCid = request.IpfsCid
	tx.TokenQuantity = request.TokenQuantity
	tx.Status = models.TRANSACTION_STATUS_PENDING_CB

	// TODO: make this a cron job with retries

	// call callback to report status
	body, _ := json.Marshal(&request)
	client := &http.Client{}
	req, err := http.NewRequest("POST", tx.CallbackUri, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Key", viper.GetString("services.sync.questori.apiKey"))
	r, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed returning call %v", err)
		glog.Error(err)
		tx.LastError = fmt.Sprintf("failed returning call %v", err)
		tx.Status = models.TRANSACTION_STATUS_ERROR
		tx.ErrorCount = tx.ErrorCount + 1
		models.Db.Save(&tx)
		return
	}
	if r.StatusCode != 200 {
		err = fmt.Errorf("failed returning call %v", r.StatusCode)
		glog.Error(err)
		tx.LastError = fmt.Sprintf("failed returning call %v", r.StatusCode)
		tx.Status = models.TRANSACTION_STATUS_ERROR
		tx.ErrorCount = tx.ErrorCount + 1
		models.Db.Save(&tx)
		return
	}

	// mark db as done
	tx.Status = models.TRANSACTION_STATUS_DONE
	resp = models.Db.Save(&tx)
	if resp.Error != nil {
		glog.Errorf("cannot get transaction %v", resp.Error)
		return
	}
}
