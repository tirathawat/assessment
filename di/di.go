package di

import (
	"github.com/tirathawat/assessment/config"
	"github.com/tirathawat/assessment/db"
	"github.com/tirathawat/assessment/expenses"
	"github.com/tirathawat/assessment/logs"
	"github.com/tirathawat/assessment/router"
	"github.com/tirathawat/assessment/srv"
)

func InitializeApplication() (server srv.Server, cleanup func(), err error) {
	logs.Setup()
	appConfig := config.NewAppConfig()
	database, cleanup, err := db.NewConnection(appConfig)
	server = srv.NewServer(appConfig, &router.Handlers{
		Expense: expenses.NewHandler(database),
	})
	return server, cleanup, err
}
