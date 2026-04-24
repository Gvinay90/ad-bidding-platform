package db

import (
	"fmt"

	"github.com/Gvinay90/ad-bidding-platform/internal/platform/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.DNS)
	case "mysql":
		dialector = mysql.Open(cfg.DNS)
	case "sqlite":
		dialector = sqlite.Open(cfg.DNS)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	return gdb, nil
}
