package database

import (
	"context"
	"fmt"
	"log"

	"github.com/aerscs/theca-public/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database defines the interface for interacting with the database
type Database interface {
	// AutoMigrate performs migration of model structures to database tables
	AutoMigrate(dst ...any) error
	// GetDB returns the database connection
	GetDB() *gorm.DB
	// Close closes the database connection
	Close() error
	// MigrateModels runs migration for the provided models
	MigrateModels(models ...any) error
	// CreateIndexes creates indexes for the provided models
	CreateIndexes() error
}

// GormDatabase implements the Database interface using GORM
type GormDatabase struct {
	Conn *gorm.DB
}

// AutoMigrate implements the method of migrating models to the database
func (g *GormDatabase) AutoMigrate(dst ...any) error {
	return g.Conn.AutoMigrate(dst...)
}

// GetDB returns the current database connection
func (g *GormDatabase) GetDB() *gorm.DB {
	return g.Conn
}

// Close closes the connection to the database
func (g *GormDatabase) Close() error {
	sqlDB, err := g.Conn.DB()
	if err != nil {
		return fmt.Errorf("error getting SQL connection: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}

	return nil
}

// MigrateModels runs migration for the provided models
func (g *GormDatabase) MigrateModels(models ...any) error {
	if len(models) == 0 {
		log.Println("[WARN]: no models provided for migration")
		return nil
	}

	if err := g.Conn.AutoMigrate(models...); err != nil {
		return fmt.Errorf("error migrating models: %w", err)
	}

	log.Println("Models migration successfully completed")
	return nil
}

// ConnectDatabase establishes a connection to the database
// and returns an interface for working with it
func ConnectDatabase(ctx context.Context, cfg *config.Config) (Database, error) {
	var conn *gorm.DB
	var err error

	if cfg.IsLocalRun {
		// Используем SQLite для локального запуска
		conn, err = gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			return nil, fmt.Errorf("error connecting to SQLite database: %w", err)
		}
		log.Printf("Successfully connected to SQLite database at %s", cfg.SQLitePath)
	} else {
		// Используем PostgreSQL для обычного запуска
		psqlInfo := fmt.Sprintf(
			"host=%s user=%s dbname=%s port=%d password=%s sslmode=%s",
			cfg.PGName, cfg.PGUser, cfg.PGDB, cfg.PGPort, cfg.PGPassword, cfg.PGSSLMode,
		)

		conn, err = gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			return nil, fmt.Errorf("error connecting to PostgreSQL database: %w", err)
		}

		// Connection pool settings
		sqlDB, err := conn.DB()
		if err != nil {
			return nil, fmt.Errorf("error getting SQL DB: %w", err)
		}

		// Standard connection pool settings
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		log.Printf("Successfully connected to PostgreSQL database")
	}

	db := &GormDatabase{Conn: conn}
	return db, nil
}

func (g *GormDatabase) CreateIndexes() error {
	if err := g.Conn.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);").Error; err != nil {
		return fmt.Errorf("filed to create username index: %w", err)
	}

	if err := g.Conn.Exec("CREATE INDEX IF NOT EXISTS idx_users_email_username ON users (email, username);").Error; err != nil {
		return fmt.Errorf("failed to create composite email_username index: %w", err)
	}
	return nil
}
