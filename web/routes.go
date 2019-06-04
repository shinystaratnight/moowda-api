package web

import (
	"moowda/apis"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

func AddRoutes(e *echo.Echo, db *gorm.DB) {
	userAPI := apis.NewUserAPI(db)
	e.POST("/register", userAPI.Register)
	e.POST("/login", userAPI.Login)

	topicAPI := apis.NewTopicAPI(db)
	e.POST("/topics", topicAPI.CreateTopic)
	e.GET("/topics", topicAPI.GetTopics)
}
