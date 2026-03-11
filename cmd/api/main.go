package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"organizational-api/internal/config"
	"organizational-api/internal/handlers"
	"organizational-api/internal/logger"
	"organizational-api/internal/repository"
	"organizational-api/internal/router"
	"organizational-api/internal/service"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	logger.Init()
	logger.Info.Println("Starting Organizational API...")

	cfg := config.Load()

	db, err := connectDB(cfg)
	if err != nil {
		logger.Error.Fatalf("Failed to connect to database: %v", err)
	}

	if err := runMigrations(cfg); err != nil {
		logger.Error.Fatalf("Failed to run migrations: %v", err)
	}

	deptRepo := repository.NewDepartmentRepository(db)
	empRepo := repository.NewEmployeeRepository(db)
	deptService := service.NewDepartmentService(deptRepo, empRepo)
	deptHandler := handlers.NewDepartmentHandler(deptService)

	handler := router.SetupRoutes(deptHandler)

	addr := fmt.Sprintf(":%s", cfg.DB.ServerPort)
	logger.Info.Printf("Server listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error.Fatalf("Server failed: %v", err)
	}
}

func connectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	var db *gorm.DB
	var err error
	maxRetries := 30

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		})

		if err == nil {
			sqlDB, err := db.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil {
					sqlDB.SetMaxOpenConns(100)
					sqlDB.SetMaxIdleConns(10)
					sqlDB.SetConnMaxLifetime(time.Hour)
					logger.Info.Println("Successfully connected to database")
					return db, nil
				}
			}
		}

		logger.Info.Printf("Waiting for database... (%d/%d)", i+1, maxRetries)
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
}

func runMigrations(cfg *config.Config) error {
	dsn := cfg.GetDSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open db for migrations: %v", err)
	}
	defer db.Close()

	migrationsDir := cfg.Migration.Dir
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "/root/migrations"
	}

	logger.Info.Printf("Running migrations from %s", migrationsDir)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	logger.Info.Println("Migrations completed successfully")
	return nil
}
