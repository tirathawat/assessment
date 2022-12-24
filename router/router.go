package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tirathawat/assessment/middleware"
)

func Register(router *gin.Engine, h *Handlers) {
	expenses := router.Group("/expenses").Use(middleware.Auth())
	{
		expenses.POST("/", h.Expense.Create)
		expenses.GET("/:id", h.Expense.Get)
	}
}
