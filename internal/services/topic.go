package services

import (
	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"net/http"
)

type TopicService interface {
	CreateTopic(topic models.TopicCore, clientId uint) (newTopic models.TopicCore, err error)
	DeleteTopic(id uint, clientId uint, clientRole models.Role) (err error)
	UpdateTopicPermissions(topic models.TopicCore, clientId uint, clientRole models.Role) (updatedTopic models.TopicCore, err error)
	GetTopicById(id uint, clientId uint, clientRole models.Role) (topic models.TopicCore, err error)
	GetAllTopics(page, pageSize *int, clientId uint, clientRole models.Role) (topics []models.TopicCore, countRows uint, err error)
}

type TopicServiceImpl struct {
	topicGateway     gateways.TopicGateway
	userGateway      gateways.UserGateway
	mosquittoGateway gateways.MosquittoGateway
}

func (t TopicServiceImpl) CreateTopic(topic models.TopicCore, clientId uint) (newTopic models.TopicCore, err error) {
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

	newTopic, err = t.topicGateway.CreateTopic(topic)
	if err != nil {
		return models.TopicCore{}, err
	}
	return newTopic, nil
}

func (t TopicServiceImpl) UpdateTopicPermissions(topic models.TopicCore, clientId uint, clientRole models.Role) (updatedTopic models.TopicCore, err error) {
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

func (t TopicServiceImpl) GetTopicById(id uint, clientId uint, clientRole models.Role) (topic models.TopicCore, err error) {
	topic, err = t.topicGateway.GetTopicById(id)
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

func (t TopicServiceImpl) GetAllTopics(page, pageSize *int, clientId uint, clientRole models.Role) (topics []models.TopicCore, countRows uint, err error) {
	offset, limit := utils.GetOffsetAndLimit(page, pageSize)
	if clientRole.String() != models.RoleSuperAdmin.String() {
		return t.topicGateway.GetTopicsByUserId(clientId, offset, limit)
	}
	return t.topicGateway.GetAllTopics(offset, limit)
}

func (t TopicServiceImpl) DeleteTopic(id uint, clientId uint, clientRole models.Role) (err error) {
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
