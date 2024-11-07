package services

import (
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"go.uber.org/fx"
)

type Services struct {
	fx.Out
	UserService      UserService
	AuthService      AuthService
	MosquittoService MosquittoService
	TopicService     TopicService
}

func SetupServices(
	userGateway gateways.UserGateway,
	mosquittoGateway gateways.MosquittoGateway,
	topicGateway gateways.TopicGateway,
) Services {
	return Services{
		UserService: &UserServiceImpl{
			userGateway: userGateway,
		},
		AuthService: &AuthServiceImpl{
			userGateway:      userGateway,
			mosquittoGateway: mosquittoGateway,
		},
		MosquittoService: &MosquittoServiceImpl{
			userGateway:      userGateway,
			mosquittoGateway: mosquittoGateway,
		},
		TopicService: &TopicServiceImpl{
			topicGateway:     topicGateway,
			mosquittoGateway: mosquittoGateway,
			userGateway:      userGateway,
		},
	}
}
