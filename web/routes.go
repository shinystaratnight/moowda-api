package web

import (
	"github.com/labstack/echo/middleware"
	"moowda/apis"
	"moowda/app"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

func AddRoutes(e *echo.Echo, db *gorm.DB) {
	userAPI := apis.NewUserAPI(db)
	e.POST("/register", userAPI.Register)
	e.POST("/login", userAPI.Login)

	r := e.Group("/")
	r.Use(middleware.JWT([]byte(app.Config.JWTSigningKey)))

	topicAPI := apis.NewTopicAPI(db)
	r.POST("/topics", topicAPI.CreateTopic)
	r.GET("/topics", topicAPI.GetTopics)
}
