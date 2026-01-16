package db

import (
	"database/sql"

	"fmt"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/user/crypto-invoice-generator/backend/internal/config"
	"github.com/user/crypto-invoice-generator/backend/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB(cfg *config.DBConfig) *gorm.DB {
	sqlDB := setupDB(cfg)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	err = gormDB.AutoMigrate(
		&models.Invoice{},
		&models.AppState{},
	)
	if err != nil {
		logrus.Fatalf("Failed to open GORM DB: %v", err)
	}

	if cfg.AppEnv == "debug" {
		gormDB = gormDB.Debug()
		logrus.Info("GORM debug mode enabled")
	}
	return gormDB
}

func setupDB(cfg *config.DBConfig) *sql.DB {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SslMode,
	)

	if cfg.SslMode != "disable" {
		authPemPath := filepath.Join(".", "config", "auth.pem")
		dataSourceName += fmt.Sprintf(" sslrootcert=%s", authPemPath)
	}

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		logrus.Fatalf("Failed to connect to DB: %v", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLife) * time.Second)

	return db
}
