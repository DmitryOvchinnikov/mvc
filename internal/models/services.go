package models

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServicesConfig func(services *Services) error

func WithGORM(dsn string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

func WithGORMLogMode(dsn string, mode bool) ServicesConfig {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	return func(s *Services) error {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

func WithTable() ServicesConfig {
	return func(s *Services) error {
		s.Table = NewTableService(s.db)
		return nil
	}
}

func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		err := cfg(*s)
		if err != nil {
			return nil, err
		}
	}

	return &s, nil
}

type Services struct {
	Table TableService
	db    *gorm.DB
}

// Close closes the DB connections.
func (s *Services) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// AutoMigrate will attempt to automatically migrate all tables.
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&Table{})
}
