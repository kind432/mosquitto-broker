package http

import (
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
)

type Handlers struct {
	AuthHandler      *authHandler
	UserHandler      userHandler
	MosquittoHandler mosquittoHandler
	TopicHandler     topicHandler
}

func SetupHandlers(
	loggers logger.Loggers,
	authService services.AuthService,
	userService services.UserService,
	mosquittoService services.MosquittoService,
	topicService services.TopicService,
) Handlers {
	return Handlers{
		AuthHandler: NewAuthHandler(loggers, authService),
		UserHandler: userHandler{
			loggers: loggers,
			user:    userService,
		},
		MosquittoHandler: mosquittoHandler{
			loggers:   loggers,
			mosquitto: mosquittoService,
		},
		TopicHandler: topicHandler{
			loggers: loggers,
			topic:   topicService,
		},
	}
}
