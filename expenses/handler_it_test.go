//go:build integration
// +build integration

package expenses_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/tirathawat/assessment/config"
	"github.com/tirathawat/assessment/db"
	"github.com/tirathawat/assessment/expenses"
	"github.com/tirathawat/assessment/router"
	"github.com/tirathawat/assessment/testutils"
	"gorm.io/gorm"
)

func setup() (endpoint string, cleanup func(), err error) {
	database, dbCleanup, err := setupDatabase(config.NewAppConfig())
	if err != nil {
		return "", dbCleanup, err
	}

	r := gin.Default()
	router.Register(r, &router.Handlers{
		Expense: expenses.NewHandler(database),
	})

	server := httptest.NewServer(r)
	endpoint = fmt.Sprintf("%s/%s", server.URL, expenses.Endpoint)

	return endpoint, func() {
		dbCleanup()
		server.Close()
	}, nil
}

func setupDatabase(appConfig *config.AppConfig) (database *gorm.DB, cleanup func(), err error) {
	database, cleanup, err = db.NewConnection(appConfig)
	if err != nil {
		return nil, cleanup, err
	}

	if err = database.Migrator().DropTable(&expenses.Expense{}); err != nil {
		return nil, cleanup, err
	}

	if err = database.AutoMigrate(&expenses.Expense{}); err != nil {
		return nil, cleanup, err
	}

	return database, cleanup, nil
}

func TestITCreate(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    string
		token          string
		want           *expenses.Expense
		wantStatusCode int
	}{
		{
			name:        "Should return 201 when create expense successfully",
			requestBody: expenses.CreateBody,
			token:       expenses.Token,
			want: &expenses.Expense{
				ID:     1,
				Title:  "test expense",
				Amount: 100,
				Note:   "test note",
				Tags:   pq.StringArray([]string{"tag1", "tag2"}),
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Should return 400 when request body is invalid",
			requestBody:    expenses.InvalidCreateBody,
			token:          expenses.Token,
			want:           nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Should return 401 when token is invalid",
			requestBody:    expenses.CreateBody,
			token:          expenses.InvalidToken,
			want:           nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "Should return 401 when token is missing",
			requestBody:    expenses.CreateBody,
			token:          "",
			want:           nil,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpRequest := &testutils.HTTPRequest{
				Method:   http.MethodPost,
				Endpoint: fmt.Sprintf("%s/", endpoint),
				Body:     test.requestBody,
				Token:    test.token,
			}

			createdExpense := &expenses.Expense{}
			statusCode, err := httpRequest.MakeHTTPRequest(createdExpense)
			if err != nil {
				t.Fatal(err)
			}

			if statusCode == http.StatusCreated && !reflect.DeepEqual(createdExpense, test.want) {
				t.Errorf("unexpected expense created: got %v want %v", createdExpense, test.want)
			}

			if statusCode != test.wantStatusCode {
				t.Errorf("unexpected status code: got %v want %v", statusCode, test.wantStatusCode)
			}
		})
	}
}

func TestITGET(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	t.Run("Should return 200 when get expense successfully", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPost,
			Endpoint: fmt.Sprintf("%s/", endpoint),
			Body:     expenses.CreateBody,
			Token:    expenses.Token,
		}

		createdExpense := &expenses.Expense{}
		statusCode, err := httpRequest.MakeHTTPRequest(createdExpense)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusCreated {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusCreated)
		}

		httpRequest = &testutils.HTTPRequest{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("%s/%d", endpoint, createdExpense.ID),
			Body:     "",
			Token:    expenses.Token,
		}

		gettedExpense := &expenses.Expense{}
		statusCode, err = httpRequest.MakeHTTPRequest(gettedExpense)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusOK {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusOK)
		}

		if !reflect.DeepEqual(createdExpense, gettedExpense) {
			t.Errorf("unexpected expense created: got %v want %v", gettedExpense, createdExpense)
		}
	})

	t.Run("Should return 401 when get expense without token", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("%s/1", endpoint),
			Body:     "",
			Token:    "",
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Should return 401 when get expense with invalid token", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("%s/1", endpoint),
			Body:     "",
			Token:    expenses.InvalidToken,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Should return 404 when not found expense", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("%s/99", endpoint),
			Body:     "",
			Token:    expenses.Token,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusNotFound {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusNotFound)
		}
	})

	t.Run("Should return 400 when get expense with invalid id", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("%s/invalid", endpoint),
			Body:     "",
			Token:    expenses.Token,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusBadRequest)
		}
	})
}

func TestITUpdate(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	t.Run("Should return 200 when update expense successfully", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPost,
			Endpoint: fmt.Sprintf("%s/", endpoint),
			Body:     expenses.CreateBody,
			Token:    expenses.Token,
		}

		createdExpense := &expenses.Expense{}
		statusCode, err := httpRequest.MakeHTTPRequest(createdExpense)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusCreated {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusCreated)
		}

		httpRequest = &testutils.HTTPRequest{
			Method:   http.MethodPut,
			Endpoint: fmt.Sprintf("%s/%d", endpoint, createdExpense.ID),
			Body:     expenses.UpdateBody,
			Token:    expenses.Token,
		}

		updatedExpense := &expenses.Expense{}
		statusCode, err = httpRequest.MakeHTTPRequest(updatedExpense)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusOK {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusOK)
		}

		want := &expenses.Expense{
			ID:     createdExpense.ID,
			Title:  "test expense update",
			Amount: 200,
			Note:   "test note update",
			Tags:   pq.StringArray([]string{"tag1", "tag2"}),
		}

		if !reflect.DeepEqual(updatedExpense, want) {
			t.Errorf("unexpected expense created: got %v want %v", updatedExpense, want)
		}
	})

	t.Run("Should return 401 when update expense without token", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPut,
			Endpoint: fmt.Sprintf("%s/1", endpoint),
			Body:     expenses.UpdateBody,
			Token:    "",
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Should return 401 when update expense with invalid token", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPut,
			Endpoint: fmt.Sprintf("%s/1", endpoint),
			Body:     expenses.UpdateBody,
			Token:    expenses.InvalidToken,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusUnauthorized)
		}
	})

	t.Run("Should return 400 when update expense with invalid body", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPut,
			Endpoint: fmt.Sprintf("%s/1", endpoint),
			Body:     expenses.InvalidUpdateBody,
			Token:    expenses.Token,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusBadRequest)
		}
	})

	t.Run("Should return 400 when update expense with invalid id", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPut,
			Endpoint: fmt.Sprintf("%s/invalid-id", endpoint),
			Body:     expenses.UpdateBody,
			Token:    expenses.Token,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusBadRequest)
		}
	})

	t.Run("Should return 404 when update expense with not found id", func(t *testing.T) {
		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPut,
			Endpoint: fmt.Sprintf("%s/999", endpoint),
			Body:     expenses.UpdateBody,
			Token:    expenses.Token,
		}

		statusCode, err := httpRequest.MakeHTTPRequest(&expenses.Expense{})
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusBadRequest)
		}
	})
}

func TestITList(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	t.Run("Should return 200 when list expenses", func(t *testing.T) {
		createdExpenses := make([]*expenses.Expense, 2)

		httpRequest := &testutils.HTTPRequest{
			Method:   http.MethodPost,
			Endpoint: fmt.Sprintf("%s/", endpoint),
			Body:     expenses.CreateBody,
			Token:    expenses.Token,
		}

		createdExpenses[0] = &expenses.Expense{}
		statusCode, err := httpRequest.MakeHTTPRequest(createdExpenses[0])
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusCreated {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusCreated)
		}

		createdExpenses[1] = &expenses.Expense{}
		statusCode, err = httpRequest.MakeHTTPRequest(createdExpenses[1])
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusCreated {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusCreated)
		}

		httpRequest = &testutils.HTTPRequest{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("%s/", endpoint),
			Body:     "",
			Token:    expenses.Token,
		}

		gettedExpenses := []*expenses.Expense{}
		statusCode, err = httpRequest.MakeHTTPRequest(&gettedExpenses)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode != http.StatusOK {
			t.Errorf("unexpected status code: got %v want %v", statusCode, http.StatusOK)
		}

		if !reflect.DeepEqual(gettedExpenses, createdExpenses) {
			t.Errorf("unexpected expenses: got %v want %v", gettedExpenses, createdExpenses)
		}
	})
}
