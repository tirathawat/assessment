package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tirathawat/assessment/errs"
	"github.com/tirathawat/assessment/logs"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidToken = errors.New("invalid token")
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			logs.Error().Msg("unauthorized")
			c.JSON(http.StatusUnauthorized, errs.Error(ErrUnauthorized))
			c.Abort()
			return
		}

		if _, err := time.Parse("January 2, 2006", token); err != nil {
			logs.Error().Err(err).Msg("invalid token")
			c.JSON(http.StatusUnauthorized, errs.Error(ErrInvalidToken))
			c.Abort()
			return
		}

		c.Next()
	}
}
