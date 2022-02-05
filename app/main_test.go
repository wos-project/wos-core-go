package main

import (
	"net/http"
	"testing"
	"flag"

	"github.com/stretchr/testify/assert"

	"github.com/wos-project/wos-core-go/app/handlers"
	"github.com/wos-project/wos-core-go/app/config"
	"github.com/wos-project/wos-core-go/app/models"
	"github.com/wos-project/wos-core-go/app/utils"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	config.ConfigPath = flag.String("config", "config.yaml", "path to YAML config file")
	config.InitializeConfiguration()
	flag.Parse()
	utils.InitMediaStorage()
	models.OpenDatabase()
	models.DropAllTables()
	models.CloseDatabase()
	models.InitializeDatabase()
}

func TestPing(t *testing.T) {
	router := handlers.SetupRouter()
	w := handlers.PerformRequest(router, "GET", "/ping", "")
	assert.Equal(t, http.StatusOK, w.Code)
}
