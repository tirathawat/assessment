package expenses

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/tirathawat/assessment/errs"
	"github.com/tirathawat/assessment/logs"
	"gorm.io/gorm"
)

var (
	ErrCreateFailed = errors.New("failed to create expense")
)

type Handler interface {
	Create(c *gin.Context)
}

type DB interface {
	Create(value interface{}) *gorm.DB
}

type handler struct {
	db DB
}

func NewHandler(db DB) Handler {
	return &handler{db}
}

func (h *handler) Create(c *gin.Context) {
	var body CreateRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		logs.Error().Err(err).Msg("failed to bind request body")
		c.JSON(http.StatusBadRequest, errs.Error(err))
		return
	}

	expense := Expense{
		Title:  body.Title,
		Amount: body.Amount,
		Note:   body.Note,
		Tags:   pq.StringArray(body.Tags),
	}

	if err := h.db.Create(&expense).Error; err != nil {
		logs.Error().Err(err).Msgf("failed to create expense: %v", expense)
		c.JSON(http.StatusInternalServerError, errs.Error(ErrCreateFailed))
		return
	}

	c.JSON(http.StatusCreated, expense)
}
