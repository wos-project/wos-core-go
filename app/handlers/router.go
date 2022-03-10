package handlers

import (
	"path"

	"github.com/golang/glog"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "github.com/wos-project/wos-core-go/app/docs" // docs is generated by Swag CLI, you have to import it.
)

// AuthMiddleware is JWT authorizer
var AuthMiddleware *jwt.GinJWTMiddleware

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(IdentityKey)
	c.JSON(200, gin.H{
		"userID": claims[IdentityKey],
		"uid":    user.(*UserJWT).Uid,
		"text":   "Hello World.",
	})
}

func validateAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		APIKey := c.Request.Header.Get(viper.GetString("auth.apiKey.key"))

		if APIKey != viper.GetString("auth.apiKey.value") {
			c.JSON(401, gin.H{"status": 401, "message": "Authentication failed"})
			return
		}

		return
	}
}

// SetupRouter creates the Gin router
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	AuthMiddleware := SetupAuth(r)
	pathApiVersion := "/" + viper.GetString("apiVersion")

	// these calls do not require JWT authorization (don't have to be logged in)
	v := r.Group(pathApiVersion)
	v.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v.POST("/object/archive/multipart", validateAPIKey(), HandleObjectArchiveUploadMultipart)
	v.GET("/object/archive/:cid", HandleObjectArchiveGet)
	v.POST("/object/index", validateAPIKey(), HandleObjectIndexPost)
	v.GET("/object/:cid/index", HandleObjectIndexGet)
	v.GET("/object/search", HandleObjectSearch)
	v.POST("/object/batchUpload", validateAPIKey(), HandleObjectBatchUploadBegin)
	v.POST("/object/batchUpload/multipart/:sessionId", validateAPIKey(), HandleObjectBatchUploadMultipart)
	v.PUT("/object/batchUpload/:sessionId", validateAPIKey(), HandleObjectBatchUploadEnd)
	v.GET("/layers", HandleLayersGet)

	v.POST("/transaction/enqueue", validateAPIKey(), HandleTransactionEnqueue)
	v.GET("/transaction/queue", validateAPIKey(), HandleTransactionQueueGet)
	v.POST("/transaction/cb", validateAPIKey(), HandleTransactionQueueCallback)

	// setup media storage static content route
	localPath := viper.GetString("media.schemes.localSimple.localPath")
	if localPath != "" {
		webPath := viper.GetString("media.schemes.localSimple.webPath")
		id := viper.GetString("media.schemes.localSimple.id")
		m := r.Group(viper.GetString("media.webMediaPath"))
		if viper.GetBool("media.requireAuth") {
			m.Use(AuthMiddleware.MiddlewareFunc())
			{
				//TODO: don't like this duplicated line
				m.Static(path.Join("/", id, webPath), localPath)
			}
		} else {
			// with this line
			m.Static(path.Join("/", id, webPath), localPath)
		}
	}

	// swagger uses basic auth
	swaggers := gin.Accounts{}
	for k, v := range viper.GetStringMapString("swagger.users") {
		swaggers[k] = v
	}
	if len(swaggers) > 0 {
		basicAuth := gin.BasicAuth(swaggers)
		swagger := r.Group("/" + viper.GetString("apiVersion") + "/swagger")
		swagger.Use(basicAuth)
		{
			swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
	} else {
		glog.Warning("no swagger users")
	}

	return r
}
