//go:build integration
// +build integration

package expenses_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/tirathawat/assessment/config"
	"github.com/tirathawat/assessment/db"
	"github.com/tirathawat/assessment/expenses"
	"github.com/tirathawat/assessment/router"
)

func TestITCreate(t *testing.T) {
	appConfig := config.NewAppConfig()
	database, cleanup, err := db.NewConnection(appConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	r := gin.Default()

	router.Register(r, &router.Handlers{
		Expense: expenses.NewHandler(database),
	})

	server := httptest.NewServer(r)
	endpoint := fmt.Sprintf("%s/expenses", server.URL)

	t.Run("Should return 201 when create expense successfully", func(t *testing.T) {
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		byteBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		var createdExpense expenses.Expense
		if err := json.Unmarshal(byteBody, &createdExpense); err != nil {
			t.Fatal(err)
		}

		want := expenses.Expense{
			ID:     1,
			Title:  "test expense",
			Amount: 100,
			Note:   "test note",
			Tags:   pq.StringArray([]string{"tag1", "tag2"}),
		}

		if !reflect.DeepEqual(createdExpense, want) {
			t.Errorf("unexpected expense created: got %v want %v", createdExpense, want)
		}
	})

	t.Run("Should return 400 when create expense with invalid payload", func(t *testing.T) {
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusBadRequest {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusBadRequest)
		}
	})

	t.Run("Should return 401 when create expense without token", func(t *testing.T) {
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("Should return 401 when create expense with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "invalid token")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})
}
