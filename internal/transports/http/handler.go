package http

import (
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
)

type Handlers struct {
	AuthHandler      AuthHandler
	UserHandler      UserHandler
	MosquittoHandler MosquittoHandler
	TopicHandler     TopicHandler
}

func SetupHandlers(
	loggers logger.Loggers,
	authService services.AuthService,
	userService services.UserService,
	mosquittoService services.MosquittoService,
	topicService services.TopicService,
) Handlers {
	return Handlers{
		AuthHandler: AuthHandler{
			loggers:     loggers,
			authService: authService,
		},
		UserHandler: UserHandler{
			loggers:     loggers,
			userService: userService,
		},
		MosquittoHandler: MosquittoHandler{
			loggers:          loggers,
			mosquittoService: mosquittoService,
		},
		TopicHandler: TopicHandler{
			loggers:      loggers,
			topicService: topicService,
		},
	}
}
