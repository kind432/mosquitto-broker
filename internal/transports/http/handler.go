package http

import (
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
)

type Handlers struct {
	AuthHandler      *authHandler
	UserHandler      *userHandler
	MosquittoHandler *mosquittoHandler
	TopicHandler     *topicHandler
}

func NewHandlers(
	loggers logger.Loggers,
	authService services.AuthService,
	userService services.UserService,
	mosquittoService services.MosquittoService,
	topicService services.TopicService,
) Handlers {
	return Handlers{
		AuthHandler:      NewAuthHandler(loggers, authService),
		UserHandler:      NewUserHandler(loggers, userService),
		MosquittoHandler: NewMosquittoHandler(loggers, mosquittoService),
		TopicHandler:     NewTopicHandler(loggers, topicService),
	}
}
