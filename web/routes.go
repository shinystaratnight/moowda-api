package web

import (
	"fmt"
	"moowda/middleware"
	"moowda/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	"moowda/apis"
	"moowda/app"
	"moowda/models"
	"moowda/sockets"
	"moowda/storage"
)

const (
	UserContextKey = "user"
)

func AddRoutes(e *echo.Echo, db *gorm.DB) {
	// Possible Without Auth Config
	skipJwtConfig := middleware.DefaultJWTConfig
	skipJwtConfig.SigningKey = []byte(app.Config.JWTSigningKey)
	skipJwtConfig.SuccessHandler = func(ctx echo.Context) {
		token := ctx.Get(UserContextKey).(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		userID := claims["userID"]

		user := new(models.User)
		if err := db.Where("id = ?", userID).Find(user).Error; err != nil {
			ctx.Error(err)
			return
		}
		ctx.Set(UserContextKey, user)
	}
	skipJwtConfig.ErrorHandler = func(err error) error {
		// skip errors because endpoints could be called without auth
		fmt.Println("skip auth")
		return nil
	}

	r := e.Group("/api")

	skipAuth := e.Group("/api")
	skipAuth.Use(middleware.JWTWithConfig(skipJwtConfig))

	skipAuthWS := e.Group("/ws")
	skipAuthWS.Use(middleware.JWTWithConfig(skipJwtConfig))

	// Services
	mailService := services.NewEmailService(app.Config.SendgridAPIKey)
	notificationService := services.NewNotificationService(mailService)

	fileStorage, err := storage.Adapters[app.Config.StorageAdapter](app.Config.StorageConfig)
	if err != nil {
		panic(fmt.Errorf("file storage wasnt able to start due to the error: %v", err))
	}
	topicService := services.NewTopicService(db)

	topicsHub := sockets.RunTopicsHub(skipAuthWS, topicService)
	messagesHub := sockets.RunMessagesHub(skipAuthWS, topicService)

	// Configure API
	userAPI := apis.NewUserAPI(db, notificationService)
	topicAPI := apis.NewTopicAPI(db, topicsHub, messagesHub)
	imagesAPI := apis.NewImagesAPI(db, fileStorage)

	// Without Auth
	r.POST("/register", userAPI.Register)
	r.POST("/login", userAPI.Login)

	r.GET("/topics/:id", topicAPI.GetTopic)

	r.POST("/restore-request", userAPI.RestoreRequest)
	r.POST("/restore", userAPI.Restore)

	r.GET("/topics/:id/messages", topicAPI.GetTopicMessages)

	skipAuth.GET("/topics", topicAPI.GetTopics)

	// With Auth Config
	jwtConfig := middleware.DefaultJWTConfig
	jwtConfig.SigningKey = []byte(app.Config.JWTSigningKey)
	jwtConfig.SuccessHandler = func(ctx echo.Context) {
		token := ctx.Get(middleware.DefaultJWTConfig.ContextKey).(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		userID := claims["userID"]

		user := new(models.User)
		if err := db.Where("id = ?", userID).Find(user).Error; err != nil {
			ctx.Error(err)
			return
		}
		ctx.Set(UserContextKey, user)
	}

	// Auth Group
	auth := e.Group("/api")
	auth.Use(middleware.JWTWithConfig(jwtConfig))

	auth.GET("/me", userAPI.Me)
	auth.POST("/topics", topicAPI.CreateTopic)
	auth.POST("/topics/:id/messages", topicAPI.CreateTopicMessage)
	auth.POST("/topics/:topicID/messages/:messageID/read", topicAPI.ReadTopicMessage)
	auth.POST("/images", imagesAPI.Upload)

}
