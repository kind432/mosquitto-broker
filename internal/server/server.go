package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/robboworld/mosquitto-broker/internal/consts"
	http2 "github.com/robboworld/mosquitto-broker/internal/transports/http"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"net/http"
)

func NewServer(
	m consts.Mode,
	lifecycle fx.Lifecycle,
	loggers logger.Loggers,
	handlers http2.Handlers,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) (err error) {
				serverHost := viper.GetString("server_host")
				port := viper.GetString("http_server_port")
				router := gin.Default()
				router.Use(
					gin.Recovery(),
					gin.Logger(),
					AuthMiddleware(loggers.Err),
				)

				switch m {
				case consts.Production:
					handlers.AuthHandler.SetupAuthRoutes(router)
					handlers.UserHandler.SetupUserRoutes(router)
					handlers.MosquittoHandler.SetupMosquittoRoutes(router)
					handlers.TopicHandler.SetupTopicRoutes(router)
				case consts.Development:
					handlers.AuthHandler.SetupAuthRoutes(router)
					handlers.UserHandler.SetupUserRoutes(router)
					handlers.MosquittoHandler.SetupMosquittoRoutes(router)
					handlers.TopicHandler.SetupTopicRoutes(router)
				}

				server := &http.Server{
					Addr: serverHost + ":" + port,
					Handler: cors.New(
						cors.Options{
							AllowedOrigins:   viper.GetStringSlice("cors.allowed_origins"),
							AllowCredentials: viper.GetBool("cors.allow_credentials"),
							AllowedMethods:   viper.GetStringSlice("cors.allowed_methods"),
							AllowedHeaders:   viper.GetStringSlice("cors.allowed_headers"),
						},
					).Handler(router),
					MaxHeaderBytes: 1 << 20,
				}

				loggers.Info.Printf(
					"The app is running in %s mode",
					m,
				)
				loggers.Info.Printf(
					"HTTP address %s:%s",
					serverHost,
					port,
				)
				go func() {
					if err = server.ListenAndServe(); err != nil {
						loggers.Err.Fatalf("Failed to listen and serve: %v", err)
					}
				}()
				return
			},
			OnStop: func(context.Context) error {
				return nil
			},
		})
}
