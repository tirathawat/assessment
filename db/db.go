package db

import (
	"github.com/tirathawat/assessment/config"
	"github.com/tirathawat/assessment/expenses"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(dbConfig *config.AppConfig) (db *gorm.DB, cleanup func(), err error) {
	db, err = gorm.Open(postgres.Open(dbConfig.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	cleanup = func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}

	err = db.AutoMigrate(&expenses.Expense{})
	return db, cleanup, err
}
