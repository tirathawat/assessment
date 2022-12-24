package expenses_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/tirathawat/assessment/expenses"
	"gorm.io/gorm"
)

type MockDB struct {
	expense       *expenses.Expense
	methodsToCall map[string]bool
	db            *gorm.DB
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	if m.expense != nil {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(m.expense).Elem())
	}

	m.methodsToCall["Create"] = true
	return m.db
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.expense != nil {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(m.expense).Elem())
	}

	m.methodsToCall["First"] = true
	return m.db
}

func (m *MockDB) ExpectToCall(methodName string) {
	if m.methodsToCall == nil {
		m.methodsToCall = make(map[string]bool)
	}

	m.methodsToCall[methodName] = false
}

func TestCreate(t *testing.T) {
	t.Run("Should return 201 when create expense successfully", func(t *testing.T) {
		want := expenses.Expense{
			Title:  "test expense",
			Amount: 100,
			Note:   "test note",
			Tags:   pq.StringArray([]string{"tag1", "tag2"}),
		}

		mockDB := &MockDB{
			expense: &want,
			db:      &gorm.DB{},
		}

		mockDB.ExpectToCall("Create")

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("POST", "/expenses", strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req

		h.Create(ctx)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusCreated)
		}

		if !mockDB.methodsToCall["Create"] {
			t.Errorf("expected Create method to be called on mock DB")
		}

		body, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Fatal(err)
		}

		var createdExpense expenses.Expense
		if err := json.Unmarshal(body, &createdExpense); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(createdExpense, want) {
			t.Errorf("unexpected expense created: got %v want %v", createdExpense, want)
		}
	})

	t.Run("Should return 400 when request body is invalid", func(t *testing.T) {
		mockDB := &MockDB{
			db: &gorm.DB{},
		}

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("POST", "/expenses", strings.NewReader(`{"title":"test expense","note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req

		h.Create(ctx)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusBadRequest)
		}
	})

	t.Run("Should return 500 when create expense failed", func(t *testing.T) {
		mockDB := &MockDB{
			db: &gorm.DB{
				Error: errors.New("db error"),
			},
		}

		mockDB.ExpectToCall("Create")

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("POST", "/expenses", strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req

		h.Create(ctx)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusInternalServerError)
		}

		if !mockDB.methodsToCall["Create"] {
			t.Errorf("expected Create method to be called on mock DB")
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("Should return 200 when get expense successfully", func(t *testing.T) {
		want := expenses.Expense{
			Title:  "test expense",
			Amount: 100,
			Note:   "test note",
			Tags:   pq.StringArray([]string{"tag1", "tag2"}),
		}

		mockDB := &MockDB{
			expense: &want,
			db:      &gorm.DB{},
		}

		mockDB.ExpectToCall("First")

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("GET", "/expenses", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req
		ctx.Params = []gin.Param{
			{
				Key:   "id",
				Value: "1",
			},
		}

		h.Get(ctx)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusOK)
		}

		if !mockDB.methodsToCall["First"] {
			t.Errorf("expected First method to be called on mock DB")
		}

		body, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Fatal(err)
		}

		var expense expenses.Expense
		if err := json.Unmarshal(body, &expense); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expense, want) {
			t.Errorf("unexpected expense created: got %v want %v", expense, want)
		}
	})

	t.Run("Should return 404 when expense not found", func(t *testing.T) {
		mockDB := &MockDB{
			db: &gorm.DB{
				Error: gorm.ErrRecordNotFound,
			},
		}

		mockDB.ExpectToCall("First")

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("GET", "/expenses", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req
		ctx.Params = []gin.Param{
			{
				Key:   "id",
				Value: "1",
			},
		}

		h.Get(ctx)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusNotFound)
		}

		if !mockDB.methodsToCall["First"] {
			t.Errorf("expected First method to be called on mock DB")
		}
	})

	t.Run("Should return 500 when get expense failed", func(t *testing.T) {
		mockDB := &MockDB{
			db: &gorm.DB{
				Error: errors.New("db error"),
			},
		}

		mockDB.ExpectToCall("First")

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("GET", "/expenses", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req
		ctx.Params = []gin.Param{
			{
				Key:   "id",
				Value: "1",
			},
		}

		h.Get(ctx)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusInternalServerError)
		}

		if !mockDB.methodsToCall["First"] {
			t.Errorf("expected First method to be called on mock DB")
		}
	})

	t.Run("Should return 400 when id is not a number", func(t *testing.T) {
		mockDB := &MockDB{}

		h := expenses.NewHandler(mockDB)

		req, err := http.NewRequest("GET", "/expenses", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rr)
		ctx.Request = req
		ctx.Params = []gin.Param{
			{
				Key:   "id",
				Value: "abc",
			},
		}

		h.Get(ctx)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}
