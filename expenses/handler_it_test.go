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

func setup() (endpoint string, cleanup func(), err error) {
	appConfig := config.NewAppConfig()
	database, dbCleanup, err := db.NewConnection(appConfig)
	if err != nil {
		return "", nil, err
	}

	r := gin.Default()
	router.Register(r, &router.Handlers{
		Expense: expenses.NewHandler(database),
	})

	if err := database.Migrator().DropTable(&expenses.Expense{}); err != nil {
		return "", dbCleanup, err
	}

	if err := database.AutoMigrate(&expenses.Expense{}); err != nil {
		return "", dbCleanup, err
	}

	server := httptest.NewServer(r)
	endpoint = fmt.Sprintf("%s/expenses", server.URL)

	return endpoint, func() {
		dbCleanup()
		server.Close()
	}, nil
}

func TestITCreate(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

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

func TestITGET(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	t.Run("Should return 200 when get expense successfully", func(t *testing.T) {
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")

		client := http.Client{}
		if _, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		req, err = http.NewRequest("GET", endpoint+"/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "January 2, 2006")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		byteBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		var expense expenses.Expense
		if err := json.Unmarshal(byteBody, &expense); err != nil {
			t.Fatal(err)
		}

		want := expenses.Expense{
			ID:     1,
			Title:  "test expense",
			Amount: 100,
			Note:   "test note",
			Tags:   pq.StringArray([]string{"tag1", "tag2"}),
		}

		if !reflect.DeepEqual(expense, want) {
			t.Errorf("unexpected expense created: got %v want %v", expense, want)
		}
	})

	t.Run("Should return 401 when get expense without token", func(t *testing.T) {
		req, err := http.NewRequest("GET", endpoint+"/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusUnauthorized {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("Should return 401 when get expense with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", endpoint+"/1", nil)
		if err != nil {
			t.Fatal(err)
		}
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

	t.Run("Should return 404 when not found expense", func(t *testing.T) {
		req, err := http.NewRequest("GET", endpoint+"/99", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "January 2, 2006")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusNotFound {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusNotFound)
		}
	})

	t.Run("Should return 400 when get expense with invalid id", func(t *testing.T) {
		req, err := http.NewRequest("GET", endpoint+"/abc", nil)
		if err != nil {
			t.Fatal(err)
		}
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
}

func TestITSave(t *testing.T) {
	endpoint, cleanup, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	t.Run("Should return 200 when update expense successfully", func(t *testing.T) {
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")

		client := http.Client{}
		if _, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		req, err = http.NewRequest("PUT", endpoint+"/1", strings.NewReader(`{"id":1,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusOK {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusOK)
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
			Title:  "test expense update",
			Amount: 200,
			Note:   "test note update",
			Tags:   pq.StringArray([]string{"tag1", "tag2"}),
		}

		if !reflect.DeepEqual(createdExpense, want) {
			t.Errorf("unexpected expense created: got %v want %v", createdExpense, want)
		}
	})

	t.Run("Should return 401 when update expense without token", func(t *testing.T) {
		req, err := http.NewRequest("PUT", endpoint+"/1", strings.NewReader(`{"id":1,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]}`))
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

	t.Run("Should return 401 when update expense with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("PUT", endpoint+"/1", strings.NewReader(`{"id":1,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]}`))
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

	t.Run("Should return 400 when update expense with invalid body", func(t *testing.T) {
		req, err := http.NewRequest("PUT", endpoint+"/1", strings.NewReader(`{"id":1,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]`))
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

	t.Run("Should return 400 when update expense with invalid id", func(t *testing.T) {
		req, err := http.NewRequest("PUT", endpoint+"/asdsad", strings.NewReader(`{"id":1,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]`))
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

	t.Run("Should return 404 when update expense with not found id", func(t *testing.T) {
		req, err := http.NewRequest("PUT", endpoint+"/999", strings.NewReader(`{"id":999,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]}`))
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

		if status := resp.StatusCode; status != http.StatusNotFound {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusNotFound)
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
		req, err := http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")

		client := http.Client{}
		if _, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		req, err = http.NewRequest("POST", endpoint, strings.NewReader(`{"title":"test expense 2","amount":200,"note":"test note 2","tags":["tag1","tag2"]}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "January 2, 2006")
		if _, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		req, err = http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "January 2, 2006")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if status := resp.StatusCode; status != http.StatusOK {
			t.Errorf("unexpected status code: got %v want %v", status, http.StatusOK)
		}

		var actual []expenses.Expense
		if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
			t.Fatal(err)
		}

		want := []expenses.Expense{
			{ID: 1, Title: "test expense", Amount: 100, Note: "test note", Tags: []string{"tag1", "tag2"}},
			{ID: 2, Title: "test expense 2", Amount: 200, Note: "test note 2", Tags: []string{"tag1", "tag2"}},
		}

		if !reflect.DeepEqual(actual, want) {
			t.Errorf("unexpected expenses: got %v want %v", actual, want)
		}
	})
}
