//go:build unit
// +build unit

package expenses_test

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/tirathawat/assessment/expenses"
	"github.com/tirathawat/assessment/testutils"
	"gorm.io/gorm"
)

const (
	createMethod = "Create"
	firstMethod  = "First"
	saveMethod   = "Save"
	findMethod   = "Find"
)

type MockDB struct {
	returnValue   interface{}
	currentMethod int
	methodsToCall map[string]bool
	dbs           []*gorm.DB
}

func (m *MockDB) call() int {
	index := m.currentMethod
	m.currentMethod++
	return index
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	if m.returnValue != nil {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(m.returnValue).Elem())
	}
	m.methodsToCall["Create"] = true
	return m.dbs[m.call()]
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.returnValue != nil {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(m.returnValue).Elem())
	}
	m.methodsToCall["First"] = true
	return m.dbs[m.call()]
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	if m.returnValue != nil {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(m.returnValue).Elem())
	}
	m.methodsToCall["Save"] = true
	return m.dbs[m.call()]
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.returnValue != nil {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(m.returnValue).Elem())
	}
	m.methodsToCall["Find"] = true
	return m.dbs[m.call()]
}

func (m *MockDB) Verify(t *testing.T) {
	for methodName, called := range m.methodsToCall {
		if !called {
			t.Errorf("expected %s to be called", methodName)
		}
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		want           *expenses.Expense
		mockDB         *MockDB
		httpRequest    *testutils.HTTPRequest
		wantStatusCode int
	}{
		{
			name: "Should return 201 when create expense successfully",
			want: &expenses.Expense{
				ID:     1,
				Title:  "test expense",
				Amount: 100,
				Note:   "test note",
				Tags:   pq.StringArray([]string{"tag1", "tag2"}),
			},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{
					ID:     1,
					Title:  "test expense",
					Amount: 100,
					Note:   "test note",
					Tags:   pq.StringArray([]string{"tag1", "tag2"}),
				},
				dbs: []*gorm.DB{{}},
				methodsToCall: map[string]bool{
					createMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPost,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.CreateBody,
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "Should return 400 when request body is invalid",
			mockDB: &MockDB{
				dbs: []*gorm.DB{{}},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPost,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.InvalidCreateBody,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Should return 400 when request body is invalid",
			mockDB: &MockDB{
				dbs: []*gorm.DB{{}},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPost,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.InvalidCreateBody,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Should return 500 when database error",
			mockDB: &MockDB{
				dbs: []*gorm.DB{{Error: errors.New("database error")}},
				methodsToCall: map[string]bool{
					createMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPost,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.CreateBody,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			createdExpense := &expenses.Expense{}
			statusCode, err := test.httpRequest.MakeTestHTTPRequest(expenses.NewHandler(test.mockDB).Create, createdExpense)
			if err != nil {
				t.Fatal(err)
			}

			if statusCode != test.wantStatusCode {
				t.Errorf("unexpected status code: got %v want %v", statusCode, test.wantStatusCode)
			}

			if statusCode == http.StatusCreated && !reflect.DeepEqual(createdExpense, test.want) {
				t.Errorf("unexpected expense created: got %v want %v", createdExpense, test.want)
			}

			test.mockDB.Verify(t)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		want           *expenses.Expense
		mockDB         *MockDB
		httpRequest    *testutils.HTTPRequest
		wantStatusCode int
	}{
		{
			name: "Should return 200 when get expense successfully",
			id:   "1",
			want: &expenses.Expense{
				Title:  "test expense",
				Amount: 100,
				Note:   "test note",
				Tags:   pq.StringArray([]string{"tag1", "tag2"}),
			},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{
					Title:  "test expense",
					Amount: 100,
					Note:   "test note",
					Tags:   pq.StringArray([]string{"tag1", "tag2"}),
				},
				dbs: []*gorm.DB{{}},
				methodsToCall: map[string]bool{
					firstMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodGet,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Should return 404 when expense not found",
			id:   "1",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{Error: gorm.ErrRecordNotFound}},
				methodsToCall: map[string]bool{
					firstMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodGet,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Should return 500 when database error",
			id:   "1",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{Error: errors.New("database error")}},
				methodsToCall: map[string]bool{
					firstMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodGet,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Should return 400 when id is not a number",
			id:   "invalid",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{}},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodGet,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expense := &expenses.Expense{}
			statusCode, err := test.httpRequest.MakeTestHTTPRequest(expenses.NewHandler(test.mockDB).Get, expense, gin.Param{Key: "id", Value: test.id})
			if err != nil {
				t.Fatal(err)
			}

			if statusCode != test.wantStatusCode {
				t.Errorf("unexpected status code: got %v want %v", statusCode, test.wantStatusCode)
			}

			if statusCode == http.StatusOK && !reflect.DeepEqual(expense, test.want) {
				t.Errorf("unexpected expense created: got %v want %v", expense, test.want)
			}

			test.mockDB.Verify(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		want           *expenses.Expense
		mockDB         *MockDB
		httpRequest    *testutils.HTTPRequest
		wantStatusCode int
	}{
		{
			name: "Should return 200 when update expense successfully",
			id:   "1",
			want: &expenses.Expense{
				ID:     1,
				Title:  "test expense",
				Amount: 100,
				Note:   "test note",
				Tags:   pq.StringArray([]string{"tag1", "tag2"}),
			},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{
					ID:     1,
					Title:  "test expense",
					Amount: 100,
					Note:   "test note",
					Tags:   pq.StringArray([]string{"tag1", "tag2"}),
				},
				dbs: []*gorm.DB{{}, {}},
				methodsToCall: map[string]bool{
					firstMethod: false,
					saveMethod:  false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPut,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.UpdateBody,
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Should return 400 when id is not a number",
			id:   "invalid",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{}},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPut,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.UpdateBody,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Should return 400 when body is invalid",
			id:   "1",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{}},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPut,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.InvalidUpdateBody,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Should return 404 when expense not found",
			id:   "1",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{Error: gorm.ErrRecordNotFound}},
				methodsToCall: map[string]bool{
					firstMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPut,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.UpdateBody,
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Should return 500 when find database error",
			id:   "1",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{Error: errors.New("error")}, {}},
				methodsToCall: map[string]bool{
					firstMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPut,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.UpdateBody,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Should return 500 when save database error",
			id:   "1",
			want: &expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &expenses.Expense{},
				dbs:         []*gorm.DB{{}, {Error: errors.New("error")}},
				methodsToCall: map[string]bool{
					firstMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodPut,
				Endpoint: fmt.Sprintf("%s/", expenses.Endpoint),
				Body:     expenses.UpdateBody,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expense := &expenses.Expense{}
			statusCode, err := test.httpRequest.MakeTestHTTPRequest(expenses.NewHandler(test.mockDB).Update, expense, gin.Param{Key: "id", Value: test.id})
			if err != nil {
				t.Fatal(err)
			}

			if statusCode != test.wantStatusCode {
				t.Errorf("unexpected status code: got %v want %v", statusCode, test.wantStatusCode)
			}

			if statusCode == http.StatusOK && !reflect.DeepEqual(expense, test.want) {
				t.Errorf("unexpected expense created: got %v want %v", expense, test.want)
			}

			test.mockDB.Verify(t)
		})
	}

}

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		want           []expenses.Expense
		mockDB         *MockDB
		httpRequest    *testutils.HTTPRequest
		wantStatusCode int
	}{
		{
			name: "Should return 200 when expenses list successfully",
			want: []expenses.Expense{{
				ID:     1,
				Title:  "test expense",
				Amount: 100,
				Note:   "test note",
				Tags:   pq.StringArray([]string{"tag1", "tag2"}),
			}},
			mockDB: &MockDB{
				returnValue: &[]expenses.Expense{{
					ID:     1,
					Title:  "test expense",
					Amount: 100,
					Note:   "test note",
					Tags:   pq.StringArray([]string{"tag1", "tag2"}),
				}},
				dbs: []*gorm.DB{{}},
				methodsToCall: map[string]bool{
					findMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodGet,
				Endpoint: expenses.Endpoint,
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Should return 500 when database error",
			want: []expenses.Expense{},
			mockDB: &MockDB{
				returnValue: &[]expenses.Expense{},
				dbs:         []*gorm.DB{{Error: errors.New("error")}},
				methodsToCall: map[string]bool{
					findMethod: false,
				},
			},
			httpRequest: &testutils.HTTPRequest{
				Method:   http.MethodGet,
				Endpoint: expenses.Endpoint,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list := []expenses.Expense{}
			statusCode, err := test.httpRequest.MakeTestHTTPRequest(expenses.NewHandler(test.mockDB).List, &list)
			if err != nil {
				t.Fatal(err)
			}

			if statusCode != test.wantStatusCode {
				t.Errorf("unexpected status code: got %v want %v", statusCode, test.wantStatusCode)
			}

			if statusCode == http.StatusOK && !reflect.DeepEqual(list, test.want) {
				t.Errorf("unexpected expenses list: got %v want %v", list, test.want)
			}

			test.mockDB.Verify(t)
		})
	}
}
