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
	ErrIDMismatch   = errors.New("id mismatch")
	ErrNotFound     = errors.New("expense not found")
	ErrGetFailed    = errors.New("failed to get expense")
	ErrUpdateFailed = errors.New("failed to update expense")
	ErrListFailed   = errors.New("failed to list expenses")
)

type Handler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	List(c *gin.Context)
}

type DB interface {
	Create(value interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
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

func (h *handler) Update(c *gin.Context) {
	var body Expense
	if err := c.ShouldBindJSON(&body); err != nil {
		logs.Error().Err(err).Msg("failed to bind request body")
		c.JSON(http.StatusBadRequest, errs.Error(err))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logs.Error().Err(err).Msgf("invalid id: %s", c.Param("id"))
		c.JSON(http.StatusBadRequest, errs.Error(ErrInvalidID))
		return
	}

	if body.ID != id {
		logs.Error().Err(err).Msgf("id mismatch: %d != %d", id, body.ID)
		c.JSON(http.StatusBadRequest, errs.Error(ErrIDMismatch))
		return
	}

	var expense Expense
	err = h.db.First(&expense, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Error().Err(err).Msgf("expense not found: %d", id)
		c.JSON(http.StatusNotFound, errs.Error(ErrNotFound))
		return
	}

	if err != nil {
		logs.Error().Err(err).Msgf("failed to get expense: %d", id)
		c.JSON(http.StatusInternalServerError, errs.Error(ErrGetFailed))
		return
	}

	if err := h.db.Save(&body).Error; err != nil {
		logs.Error().Err(err).Msgf("failed to update expense: %v", body)
		c.JSON(http.StatusInternalServerError, errs.Error(ErrUpdateFailed))
		return
	}

	c.JSON(http.StatusOK, body)
}

func (h *handler) List(c *gin.Context) {
	var expenses []Expense
	err := h.db.Find(&expenses).Error
	if err != nil {
		logs.Error().Err(err).Msg("failed to list expenses")
		c.JSON(http.StatusInternalServerError, ErrListFailed)
		return
	}

	c.JSON(http.StatusOK, expenses)
}
