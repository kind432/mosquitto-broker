package db

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
)

type PostgresClient struct {
	Db         *gorm.DB
	InfoLogger *log.Logger
}

func InitPostgresClient(m consts.Mode, loggers logger.Loggers) (postgresClient PostgresClient, err error) {
	// set stdout gorm logger depends on app mode
	var dbLogger gormLogger.Interface
	switch m {
	case consts.Production:
		gormF, err := os.OpenFile(viper.GetString("logger.gorm"), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			loggers.Err.Fatalf("%s", err.Error())
		}
		defer func(gormF *os.File) {
			err := gormF.Close()
			if err != nil {
				loggers.Err.Fatalf("%s", err.Error())
			}
		}(gormF)
		dbLogger = gormLogger.New(
			log.New(gormF, "[GORM]\t", log.LstdFlags),
			gormLogger.Config{
				SlowThreshold:             time.Second,     // Slow SQL threshold
				LogLevel:                  gormLogger.Info, // Log level
				IgnoreRecordNotFoundError: false,           // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,            // Disable color
			},
		)
	case consts.Development:
		dbLogger = gormLogger.New(
			log.New(os.Stdout, "[GORM]\t", log.LstdFlags),
			gormLogger.Config{
				SlowThreshold:             time.Second,     // Slow SQL threshold
				LogLevel:                  gormLogger.Info, // Log level
				IgnoreRecordNotFoundError: false,           // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,            // Disable color
			},
		)
	}

	db, err := gorm.Open(postgres.Open(viper.GetString("postgres_dsn")), &gorm.Config{Logger: dbLogger})
	if err != nil {
		loggers.Err.Fatalf("Failed to initialize postgres client: %s", err.Error())
		return
	}
	postgresClient = PostgresClient{
		Db:         db,
		InfoLogger: loggers.Info,
	}
	if migrateErr := postgresClient.Migrate(); migrateErr != nil {
		loggers.Err.Fatalf("Failed to migrate: %s", migrateErr.Error())
	}
	return postgresClient, err
}

func (c *PostgresClient) Migrate() (err error) {
	err = c.Db.AutoMigrate(
		&models.UserCore{},
		&models.TopicCore{},
	)
	if err != nil {
		return err
	}
	return nil
}
