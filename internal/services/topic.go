package services

import (
	"net/http"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type topicService struct {
	topicGateway     gateways.TopicGateway
	userGateway      gateways.UserGateway
	mosquittoGateway gateways.MosquittoGateway
}

func NewTopicService(
	topicGateway gateways.TopicGateway,
	userGateway gateways.UserGateway,
	mosquittoGateway gateways.MosquittoGateway,
) *topicService {
	return &topicService{
		topicGateway:     topicGateway,
		userGateway:      userGateway,
		mosquittoGateway: mosquittoGateway,
	}
}

func (t *topicService) CreateTopic(topic models.TopicCore, clientId uint) (models.TopicCore, error) {
	user, err := t.userGateway.GetUserById(clientId)
	if err != nil {
		return models.TopicCore{}, err
	}

	exist, err := t.topicGateway.DoesExistTopic(0, user.ID, topic.Name)
	if err != nil {
		return models.TopicCore{}, err
	}
	if exist {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusBadRequest,
			Message: consts.ErrTopicIsExist,
		}
	}

	t.mosquittoGateway.WriteNewTopicToAcl(user.Email, topic.Name, topic.CanRead, topic.CanWrite)

	return t.topicGateway.CreateTopic(topic)
}

func (t *topicService) UpdateTopicPermissions(topic models.TopicCore, clientId uint, clientRole models.Role) (models.TopicCore, error) {
	currentTopic, err := t.topicGateway.GetTopicById(topic.ID)
	if err != nil {
		return models.TopicCore{}, err
	}
	if clientRole.String() != models.RoleSuperAdmin.String() && currentTopic.UserId != clientId {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusForbidden,
			Message: consts.ErrAccessDenied,
		}
	}

	user, err := t.userGateway.GetUserById(clientId)
	if err != nil {
		return models.TopicCore{}, err
	}

	t.mosquittoGateway.WriteUpdatedTopicToAcl(user.Email, currentTopic.Name, topic.CanRead, topic.CanWrite)
	return t.topicGateway.UpdateTopicPermissions(topic)
}

func (t *topicService) GetTopicById(id uint, clientId uint, clientRole models.Role) (models.TopicCore, error) {
	topic, err := t.topicGateway.GetTopicById(id)
	if err != nil {
		return models.TopicCore{}, err
	}
	if clientRole.String() != models.RoleSuperAdmin.String() && topic.UserId != clientId {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusForbidden,
			Message: consts.ErrAccessDenied,
		}
	}

	return topic, nil
}

func (t *topicService) GetAllTopics(page, pageSize *int, clientId uint, clientRole models.Role) ([]models.TopicCore, uint, error) {
	offset, limit := utils.GetOffsetAndLimit(page, pageSize)
	if clientRole.String() != models.RoleSuperAdmin.String() {
		return t.topicGateway.GetTopicsByUserId(clientId, offset, limit)
	}
	return t.topicGateway.GetAllTopics(offset, limit)
}

func (t *topicService) DeleteTopic(id uint, clientId uint, clientRole models.Role) (err error) {
	topic, err := t.topicGateway.GetTopicById(id)
	if err != nil {
		return err
	}
	if clientRole.String() != models.RoleSuperAdmin.String() && topic.UserId != clientId {
		return utils.ResponseError{
			Code:    http.StatusForbidden,
			Message: consts.ErrAccessDenied,
		}
	}
	user, err := t.userGateway.GetUserById(clientId)
	if err != nil {
		return err
	}

	t.mosquittoGateway.DeleteTopicFromAcl(user.Email, topic.Name)
	return t.topicGateway.DeleteTopic(id)
}
