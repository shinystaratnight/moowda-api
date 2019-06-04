package apis

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"moowda/models"
	"net/http"
	"strconv"
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

func (s *TopicAPI) GetTopic(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var topic models.TopicDetail

	if err := s.db.Where("id = ?", id).
		Select("id, title, (?) as messages_count, (?) as last_message_date",
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = ?", id).QueryExpr(),
			s.db.Table("topics_topicmessage").Select("created_at").Where("topics_topicmessage.topic_id = ?", id).Order("id DESC").Limit(1).QueryExpr(),
		).
		Find(&topic).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topic)
}
