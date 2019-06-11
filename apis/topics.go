package apis

import (
	"fmt"
	"moowda/sockets"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	apiErrors "moowda/errors"
	"moowda/models"
)

type TopicAPI struct {
	db          *gorm.DB
	topicsHub   *sockets.Hub
	messagesHub *sockets.Hub
}

func NewTopicAPI(db *gorm.DB, topicsHub *sockets.Hub, messagesHub *sockets.Hub) *TopicAPI {
	return &TopicAPI{db: db, topicsHub: topicsHub, messagesHub: messagesHub}
}

func (s *TopicAPI) CreateTopic(c echo.Context) error {
	user := c.Get("user").(*models.User)

	topic := new(models.Topic)
	if err := c.Bind(topic); err != nil {
		return err
	}
	topic.OwnerID = user.ID

	if err := s.db.Create(topic).Error; err != nil {
		return err
	}

	s.topicsHub.BroadcastTopic(topic)

	return c.JSON(http.StatusOK, topic)
}

func (s *TopicAPI) GetTopics(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	fmt.Printf(">>> %v, %v", user, ok)

	var topics []models.TopicCard

	var query *gorm.DB
	if ok {
		query = s.db.Select("id, title, (?) as unread_messages_count, (?) as messages_count",
			s.db.Table("topics_topicmessage").
				Select("COUNT(*)").
				Where("topics_topicmessage.topic_id = topics_topic.id and topics_topicmessage.user_id <> ? and topics_topicmessage.id > (select coalesce((?), 0))", user.ID,
					s.db.Table("topics_topicmessageread").Select("coalesce(message_id, 0)").Where("topics_topicmessageread.topic_id = topics_topicmessage.topic_id").Order("id desc").Limit(1).QueryExpr(),
				).QueryExpr(),
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = topics_topic.id").QueryExpr(),
		).Find(&topics)
	} else {
		query = s.db.Select("id, title, (?) as unread_messages_count, (?) as messages_count",
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = topics_topic.id").QueryExpr(),
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = topics_topic.id").QueryExpr(),
		).Find(&topics)
	}
	if err := query.Error; err != nil {
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
		).Find(&topic).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topic)
}

func (s *TopicAPI) CreateTopicMessage(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("id"))

	createTopicMessageRequest := new(models.CreateTopicMessageRequest)
	if err := c.Bind(createTopicMessageRequest); err != nil {
		return err
	}
	if err := createTopicMessageRequest.Validate(); err != nil {
		return apiErrors.InvalidData(err.(validation.Errors))
	}

	user := c.Get("user").(*models.User)

	message := models.TopicMessage{
		Content: createTopicMessageRequest.Content,
		TopicID: uint(topicID),
		UserID:  user.ID,
	}

	var image models.Image
	for _, imageID := range createTopicMessageRequest.Images {
		if err := s.db.Where("id = ?", imageID).First(&image).Error; err != nil {
			return err
		}

		message.Images = append(message.Images, models.TopicMessageImage{
			ImageID: image.ID,
		})
	}

	if err := s.db.Create(&message).Error; err != nil {
		return err
	}

	if err := s.db.Preload("Topic").Preload("User").Preload("Images.Image").Where("id = ?", message.ID).Find(&message).Error; err != nil {
		return err
	}

	var newTopic models.TopicCard
	if err := s.db.Where("id = ?", message.TopicID).
		Select("id, title, (?) as unread_messages_count, (?) as messages_count",
			0,
			s.db.Table("topics_topicmessage").Select("count(*)").Where("topics_topicmessage.topic_id = ?", message.TopicID).QueryExpr(),
		).Find(&newTopic).Error; err != nil {
		return err
	}

	s.topicsHub.BroadcastTopic(&message.Topic)
	s.messagesHub.BroadcastMessage(&message)

	return c.JSON(http.StatusOK, message)
}

func (s *TopicAPI) ReadTopicMessage(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("topicID"))
	messageID, _ := strconv.Atoi(c.Param("messageID"))

	messageRead := new(models.TopicMessageRead)
	s.db.Where("topic_id = ?", topicID).Find(&messageRead)

	user := c.Get("user").(*models.User)

	readMessage := models.TopicMessageRead{
		TopicID:        uint(topicID),
		UserID:         user.ID,
		TopicMessageID: uint(messageID),
	}

	var query *gorm.DB
	if messageRead.ID > 0 {
		query = s.db.Model(&readMessage).Where("id = ?", messageRead.ID).Update(readMessage)
	} else {
		query = s.db.Create(&readMessage)
	}

	if err := query.Error; err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (s *TopicAPI) GetTopicMessages(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("id"))

	var messages []models.TopicMessage

	if err := s.db.Preload("User").Preload("Images.Image").Where("topic_id = ?", topicID).Find(&messages).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"count":   len(messages),
		"results": messages,
	})
}
