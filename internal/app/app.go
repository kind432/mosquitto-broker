package app

import (
	"log"
	"os"

	"go.uber.org/fx"

	"github.com/robboworld/mosquitto-broker/internal/configs"
	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"github.com/robboworld/mosquitto-broker/internal/mosquitto"
	"github.com/robboworld/mosquitto-broker/internal/server"
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/internal/transports/http"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
)

func InvokeWith(m consts.Mode, options ...fx.Option) *fx.App {
	if err := configs.Init(m); err != nil {
		log.Fatalf("%s", err.Error())
	}
	di := []fx.Option{
		fx.Provide(func() consts.Mode { return m }),
		fx.Provide(logger.InitLogger),
		fx.Provide(mosquitto.NewMosquitto),
		fx.Provide(db.InitPostgresClient),
		fx.Provide(gateways.SetupGateways),
		fx.Provide(services.SetupServices),
		fx.Provide(http.SetupHandlers),
	}
	for _, option := range options {
		di = append(di, option)
	}
	return fx.New(di...)
}

func RunApp() {
	if len(os.Args) == 2 && (consts.Mode(os.Args[1]) == consts.Development ||
		consts.Mode(os.Args[1]) == consts.Production) {
		InvokeWith(consts.Mode(os.Args[1]), fx.Invoke(server.NewServer)).Run()
	} else {
		InvokeWith(consts.Development, fx.Invoke(server.NewServer)).Run()
	}
}
