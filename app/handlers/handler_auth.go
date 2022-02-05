package handlers

import (
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type reqLogin struct {
	Email           string `json:"email"`
	PhoneNum        string `json:"phonenum"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	RememberMeToken string `json:"rememberMeToken"`
}

var IdentityKey = "jwtid"

// User encoded into JWT
type UserJWT struct {
	Uid             string
	RememberMeToken string
}

// hashPasswordBcrypt hashes and salts a password with bcrypt
func hashPasswordBcrypt(password string) (string, error) {
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPw), nil
}

// compareHashPasswordBcrypt compares a clear and hashed password
func compareHashPasswordBcrypt(hashedPassword string, clearPassword string) bool {
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(clearPassword))
	if err != nil {
		return false
	}
	return true
}

// SetupAuth Authenticates the user (logs in), sets up the JWT auth token, and returns username and rememberMeToken,
// requires the App-Key header to be present.
// There are two ways of logging in, supplying username/password or supplying RememberMeToken header.
// In both cases the user needs to activate the account first by clicking confirm email.
func SetupAuth(r *gin.Engine) *jwt.GinJWTMiddleware {

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       viper.GetString("auth.realm"),
		Key:         []byte(viper.GetString("auth.key")),
		Timeout:     time.Duration(viper.GetInt64("auth.timeoutSeconds")) * time.Second,
		MaxRefresh:  time.Duration(viper.GetInt64("auth.maxRefreshSeconds")) * time.Second,
		IdentityKey: IdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*UserJWT); ok {
				return jwt.MapClaims{
					IdentityKey: v.Uid,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &UserJWT{
				Uid: claims[IdentityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			// Insert auth here
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// NOTE: not sure if we need to check users here
			if _, ok := data.(*UserJWT); ok {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	if err != nil {
		glog.Fatal("JWT Error:" + err.Error())
	}

	v := r.Group(viper.GetString("apiVersion"))
	v.POST("/login", authMiddleware.LoginHandler)
	v.POST("/user/login", authMiddleware.LoginHandler)
	v.GET("/user/refreshToken", authMiddleware.RefreshHandler)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		glog.Errorf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	return authMiddleware
}

// ValJWT validates the JWT and returns user UID
func ValJWT(c *gin.Context) string {
	userjwt, _ := c.Get(IdentityKey)
	return userjwt.(*UserJWT).Uid
}

// GetUserFromJWT validates the JWT and returns user model
func GetUserFromJWT(c *gin.Context) (string, error) {
	userjwt, _ := c.Get(IdentityKey)
	userUid := userjwt.(*UserJWT).Uid
	return userUid, nil
}
