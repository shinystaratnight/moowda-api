package web

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"moowda/apis"
	"moowda/app"
	"moowda/models"
	"moowda/sockets"
	"moowda/storage"
)

func AddRoutes(e *echo.Echo, db *gorm.DB, hub *sockets.Hub) {
	r := e.Group("/api")

	userAPI := apis.NewUserAPI(db)
	topicAPI := apis.NewTopicAPI(db, hub)

	fileStorage, err := storage.Adapters[app.Config.StorageAdapter](app.Config.StorageConfig)
	if err != nil {
		panic(fmt.Errorf("file storage wasnt able to start due to the error: %v", err))
	}
	imagesAPI := apis.NewImagesAPI(db, fileStorage)

	// Without Auth
	r.POST("/register", userAPI.Register)
	r.POST("/login", userAPI.Login)

	r.GET("/topics", topicAPI.GetTopics)
	r.GET("/topics/:id", topicAPI.GetTopic)

	r.POST("/restore-request", userAPI.RestoreRequest)
	r.POST("/restore", userAPI.Restore)

	r.GET("/topics/:id/messages", topicAPI.GetTopicMessages)

	// With Auth
	auth := e.Group("/api")

	jwtConfig := middleware.DefaultJWTConfig
	jwtConfig.SigningKey = []byte(app.Config.JWTSigningKey)
	jwtConfig.SuccessHandler = func(ctx echo.Context) {
		token := ctx.Get("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		userID := claims["userID"]

		user := new(models.User)
		if err := db.Where("id = ?", userID).Find(user).Error; err != nil {
			ctx.Error(err)
			return
		}
		ctx.Set("user", user)
	}
	auth.Use(middleware.JWTWithConfig(jwtConfig))

	auth.GET("/me", userAPI.Me)
	auth.POST("/topics", topicAPI.CreateTopic)
	auth.POST("/topics/:id/messages", topicAPI.CreateTopicMessage)
	auth.POST("/topics/:topicID/messages/:messageID/read", topicAPI.ReadTopicMessage)
	auth.POST("/images", imagesAPI.Upload)

}
