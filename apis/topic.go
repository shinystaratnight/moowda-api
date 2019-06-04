package apis

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"moowda/models"
	"net/http"
)

type TopicAPI struct {
	db *gorm.DB
}

func NewTopicAPI(db *gorm.DB) *TopicAPI {
	return &TopicAPI{db: db}
}

func (s *TopicAPI) CreateTopic(c echo.Context) error {
	topic := new(models.Topic)
	if err := c.Bind(topic); err != nil {
		return err
	}

	var user models.User
	if err := s.db.First(&user).Error; err != nil {
		return err
	}

	topic.OwnerID = user.ID

	if err := s.db.Create(topic).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topic)
}

func (s *TopicAPI) GetTopics(c echo.Context) error {
	var topics []models.TopicCard

	if err := s.db.Select("id, title, (?) as messages_count", s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = topics_topic.id").QueryExpr()).Find(&topics).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topics)
}
