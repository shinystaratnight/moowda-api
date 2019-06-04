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
	topicAPI := apis.NewTopicAPI(db)

	// Without Auth
	e.POST("/register", userAPI.Register)
	e.POST("/login", userAPI.Login)

	e.GET("/topics", topicAPI.GetTopics)

	// With Auth
	r := e.Group("/")
	r.Use(middleware.JWT([]byte(app.Config.JWTSigningKey)))

	r.POST("/topics", topicAPI.CreateTopic)

}
