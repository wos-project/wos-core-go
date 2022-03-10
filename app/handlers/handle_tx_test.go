package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/spf13/viper"

	"github.com/wos-project/wos-core-go/app/utils"
)

var _ = func() bool {
	testing.Init()
	return true
}()

// TestTx tests NFT and token transactions.  It requires that the transactor is running
func TestTx(t *testing.T) {

	os.RemoveAll("/var/tmp/mediatmp")
	os.MkdirAll("/var/tmp/mediatmp", 0755)

	router := SetupRouter()
	utils.InitMediaStorage()

	// run the main web server
	go func() {
		err := http.ListenAndServe(":"+viper.GetString("host.hosts.localhost.port"), router)
		assert.Nil(t, err)
	}()

	fmt.Println("!!!!!!make sure transactor and ganache are running!!!!!!")

	contractAddr := ""

	// spin up another web server that will receive the callback
	mockApi := http.NewServeMux()
	mockApi.HandleFunc("/cb/erc721", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "Got callback!")
		var request reqTransactionQueuedItemCallback

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			assert.Fail(t, "failure marshaling cb from transactor")
			fmt.Println("failure marshaling cb from transactor")
			return
		}
		contractAddr = request.ContractAddr
	})
	go func() {
		err := http.ListenAndServe(":7890", mockApi)
		assert.Nil(t, err)
	}()

	// enqueue transaction
	txErc721Spec := reqTransactionEnqueueSpecAirdropErc721{
		Uid:        utils.GenerateBase64Rand(),
		WalletAddr: "d36E5AEaBa5f35997e374Bb5D1A10B770aACE6e8",
		WalletKind: "ethereum",
		IpfsCid:    "a7b969b1a6d25974b2692916abc41312d89bb6a00355ae06ebdb159b89ef5bd8",
		CallbackUri: "http://127.0.0.1:7890/cb/erc721",
	}
	txErc721 := reqTransactionEnqueue{
		Kind: "erc721",
		Spec: &txErc721Spec,
	}
	txNftBody, _ := json.Marshal(&txErc721)

	w := PerformRequest(router, "POST", "/transaction/enqueue", string(txNftBody))
	assert.Equal(t, http.StatusOK, w.Code)

	// wait for the callback with contract address from the transactor
	time.Sleep(5 * time.Second)
	assert.NotEmpty(t, contractAddr)
}
