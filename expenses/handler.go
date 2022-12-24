package expenses

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/tirathawat/assessment/errs"
	"github.com/tirathawat/assessment/logs"
	"gorm.io/gorm"
)

var (
	ErrCreateFailed = errors.New("failed to create expense")
	ErrInvalidID    = errors.New("invalid id")
	ErrNotFound     = errors.New("expense not found")
	ErrGetFailed    = errors.New("failed to get expense")
)

type Handler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
}

type DB interface {
	Create(value interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
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

func (h *handler) Get(c *gin.Context) {
	var expense Expense

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logs.Error().Err(err).Msgf("invalid id: %s", c.Param("id"))
		c.JSON(http.StatusBadRequest, errs.Error(ErrInvalidID))
		return
	}

	err = h.db.First(&expense, "id = ?", id).Error
	if err == nil {
		c.JSON(http.StatusOK, expense)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Error().Err(err).Msgf("expense not found: %d", id)
		c.JSON(http.StatusNotFound, errs.Error(ErrNotFound))
		return
	}

	logs.Error().Err(err).Msgf("failed to get expense: %d", id)
	c.JSON(http.StatusInternalServerError, errs.Error(ErrGetFailed))
}
