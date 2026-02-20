package gateways

import (
	"github.com/robboworld/mosquitto-broker/internal/models"
	"go.uber.org/fx"

	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/mosquitto"
)

type UserGateway interface {
	Create(user models.UserCore) error
	GetById(id uint) (models.UserCore, error)
	GetByEmail(email string) (models.UserCore, error)
	DoesExistEmail(id uint, email string) (bool, error)
	SetMosquittoOn(id uint, mosquittoOn bool) error
}

type MosquittoGateway interface {
	WriteMosquittoPasswd(email, password string)
	WriteNewUserToAcl(email string)
	WriteNewTopicToAcl(email, name string, canRead, canWrite bool)
	WriteUpdatedTopicToAcl(email, name string, canRead, canWrite bool)
	DeleteTopicFromAcl(username, name string)
	MosquittoLaunch(mosquittoOn bool)
	MosquittoStop()
}

type TopicGateway interface {
	Create(topic models.TopicCore) (models.TopicCore, error)
	GetById(id uint) (models.TopicCore, error)
	GetByUserId(userId uint, offset, limit int) (topics []models.TopicCore, countRows uint, err error)
	GetAll(offset, limit int) (topics []models.TopicCore, countRows uint, err error)
	UpdatePermissions(topic models.TopicCore) (models.TopicCore, error)
	Delete(id uint) error
	DoesExist(id, userId uint, name string) (bool, error)
}

type Gateways struct {
	fx.Out
	UserGateway      UserGateway
	MosquittoGateway MosquittoGateway
	TopicGateway     TopicGateway
}

func New(
	pc db.PostgresClient,
	mosquitto mosquitto.Mosquitto,
) Gateways {
	return Gateways{
		UserGateway:      NewUserGateway(pc),
		MosquittoGateway: NewMosquittoGateway(mosquitto),
		TopicGateway:     NewTopicGateway(pc),
	}
}
