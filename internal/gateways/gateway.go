package gateways

import (
	"github.com/robboworld/mosquitto-broker/internal/models"
	"go.uber.org/fx"

	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/mosquitto"
)

type UserGateway interface {
	CreateUser(user models.UserCore) error
	GetUserById(id uint) (models.UserCore, error)
	GetUserByEmail(email string) (models.UserCore, error)
	DoesExistEmail(id uint, email string) (bool, error)
	SetMosquittoOn(id uint, mosquittoOn bool) error
}

type MosquittoGateway interface {
	MosquittoLaunch(mosquittoOn bool)
	MosquittoStop()
	WriteMosquittoPasswd(email, password string)
	WriteNewUserToAcl(email string)
	WriteNewTopicToAcl(email, name string, canRead, canWrite bool)
	WriteUpdatedTopicToAcl(email, name string, canRead, canWrite bool)
	DeleteTopicFromAcl(username, name string)
}

type TopicGateway interface {
	CreateTopic(topic models.TopicCore) (models.TopicCore, error)
	DeleteTopic(id uint) error
	UpdateTopicPermissions(topic models.TopicCore) (models.TopicCore, error)
	DoesExistTopic(id, userId uint, name string) (bool, error)
	GetTopicById(id uint) (models.TopicCore, error)
	GetTopicsByUserId(userId uint, offset, limit int) (topics []models.TopicCore, countRows uint, err error)
	GetAllTopics(offset, limit int) (topics []models.TopicCore, countRows uint, err error)
}

type Gateways struct {
	fx.Out
	UserGateway      UserGateway
	MosquittoGateway MosquittoGateway
	TopicGateway     TopicGateway
}

func SetupGateways(
	pc db.PostgresClient,
	mosquitto mosquitto.Mosquitto,
) Gateways {
	return Gateways{
		UserGateway:      NewUserGateway(pc),
		MosquittoGateway: NewMosquittoGateway(mosquitto),
		TopicGateway:     NewTopicGateway(pc),
	}
}
