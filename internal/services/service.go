package services

import (
	"github.com/robboworld/mosquitto-broker/internal/models"
	"go.uber.org/fx"

	"github.com/robboworld/mosquitto-broker/internal/gateways"
)

type UserService interface {
	GetById(id uint, clientId uint, clientRole models.Role) (models.UserCore, error)
}

type AuthService interface {
	SignUp(newUser models.UserCore) error
	SignIn(email, password string) (Tokens, error)
	Refresh(token string) (string, error)
}

type MosquittoService interface {
	Launch(id uint, mosquittoOn bool) error
	Stop()
}

type TopicService interface {
	Create(topic models.TopicCore, clientId uint) (models.TopicCore, error)
	GetById(id uint, clientId uint, clientRole models.Role) (models.TopicCore, error)
	GetAll(page, pageSize *int, clientId uint, clientRole models.Role) (topics []models.TopicCore, countRows uint, err error)
	UpdatePermissions(topic models.TopicCore, clientId uint, clientRole models.Role) (models.TopicCore, error)
	Delete(id uint, clientId uint, clientRole models.Role) error
}

type Services struct {
	fx.Out
	UserService      UserService
	AuthService      AuthService
	MosquittoService MosquittoService
	TopicService     TopicService
}

func New(
	userGateway gateways.UserGateway,
	mosquittoGateway gateways.MosquittoGateway,
	topicGateway gateways.TopicGateway,
) Services {
	return Services{
		UserService:      NewUserService(userGateway),
		AuthService:      NewAuthService(userGateway, mosquittoGateway),
		MosquittoService: NewMosquittoService(userGateway, mosquittoGateway),
		TopicService:     NewTopicService(topicGateway, userGateway, mosquittoGateway),
	}
}
