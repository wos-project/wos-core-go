package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/autotls"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	ginglog "github.com/szuecs/gin-glog"
	"github.com/robfig/cron/v3"

	"github.com/wos-project/wos-core-go/app/docs" // docs is generated by Swag CLI, you have to import it.

	"github.com/wos-project/wos-core-go/app/models"
	"github.com/wos-project/wos-core-go/app/config"
	"github.com/wos-project/wos-core-go/app/utils"
	"github.com/wos-project/wos-core-go/app/handlers"
)

// @title World Object Store Core
// @description World Object Store Core - RESTful API
// @termsOfService https://questori.com/about
// @contact.name API Support
// @contact.url https://questori.com/support
// @contact.email support@questori.com

// @securityDefinitions.apikey App-Key
// @in header
// @name App-Key

// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
func main() {

	config.ConfigPath = flag.String("config", "config.yaml", "path to YAML config file")
	flag.Parse()
	config.InitializeConfiguration()

	glog.Infof("starting %s mode", viper.GetString("mode"))

	utils.InitMediaStorage()

	models.InitializeDatabase()
	defer models.CloseDatabase()

	r := handlers.SetupRouter()
	r.Use(ginglog.Logger(3 * time.Second))

	docs.SwaggerInfo.BasePath = "/" + viper.GetString("apiVersion")
	docs.SwaggerInfo.Version = viper.GetString("apiVersion")

	// cron jobs
	go handlers.CleanupOldTempFiles();
	c := cron.New()
	c.AddFunc(viper.GetString("schedules.cleanupOldTempFiles"), func() {
		handlers.CleanupOldTempFiles();
	})
	c.Start()
	
	switch viper.GetString("host.mode") {
	case "letsEncrypt":
		if err := autotls.Run(r, viper.GetString("host.hosts.letsEncrypt.domain")); err != nil {
			glog.Fatal(err)
		}
	case "localhost":
		if err := http.ListenAndServe(":"+viper.GetString("host.hosts.localhost.port"), r); err != nil {
			glog.Fatal(err)
		}
	default:
		glog.Fatal("unknown host mode")
	}

	switch viper.GetString("swagger.host") {
	case "letsEncrypt":
		docs.SwaggerInfo.Host = viper.GetString("host.hosts.letsEncrypt.domain")
		docs.SwaggerInfo.Schemes = []string{viper.GetString("host.hosts.letsEncrypt.scheme")}
	case "localhost":
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", viper.GetString("host.hosts.localhost.domain"), viper.GetString("host.hosts.localhost.port"))
		docs.SwaggerInfo.Schemes = []string{viper.GetString("host.hosts.localhost.scheme")}
	default:
		glog.Fatal("unknown host mode")
	}
}
