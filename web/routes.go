package web

import (
	"github.com/labstack/echo/middleware"
	"moowda/apis"
	"moowda/app"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

func AddRoutes(e *echo.Echo, db *gorm.DB) {
	r := e.Group("/api")

	userAPI := apis.NewUserAPI(db)
	topicAPI := apis.NewTopicAPI(db)

	// Without Auth
	r.POST("/register", userAPI.Register)
	r.POST("/login", userAPI.Login)

	r.GET("/topics", topicAPI.GetTopics)
	r.GET("/topics/:id", topicAPI.GetTopic)

	// With Auth
	auth := e.Group("/api")
	auth.Use(middleware.JWT([]byte(app.Config.JWTSigningKey)))

	auth.POST("/topics", topicAPI.CreateTopic)

}
