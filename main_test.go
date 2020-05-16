package helper

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Server() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/products", BrowseProduct).Methods("GET")
	router.HandleFunc("/api/products", CreateProduct).Methods("POST")
	router.HandleFunc("/api/products/{id}", ShowProduct).Methods("GET")
	router.HandleFunc("/api/products/{id}", UpdateProduct).Methods("PATCH")
	router.HandleFunc("/api/products/{id}", DeleteProduct).Methods("DELETE")
	return router
}

func Test_BrowseProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' when open database", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "price"}).
		AddRow(1, "Product1", 100).
		AddRow(2, "Product2", 100)
	mock.ExpectQuery("SELECT id, name, price FROM products").
		WillReturnRows(rows)

	mysqlDB = db
	request, _ := http.NewRequest("GET", "/api/products", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Unexpected response code")
	responseBody, _ := ioutil.ReadAll(response.Body)
	expectedResponse := `{"data":[{"type":"products","id":"1","attributes":{"name":"Product1","price":100},"links":{"self":"http://localhost:8000/api/products/1"}},{"type":"products","id":"2","attributes":{"name":"Product2","price":100},"links":{"self":"http://localhost:8000/api/products/2"}}],"meta":{"total":2}}`
	assert.Equal(t, string(bytes.TrimSpace(responseBody)), expectedResponse, "Response not match")
}

func Test_CreateProduct(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"name":  "Product 1",
				"price": 200,
			},
		},
	}

	body, _ := json.Marshal(data)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Logf("an error '%s' when open database", err)
	}
	defer db.Close()

	mock.ExpectPrepare("INSERT INTO products (name, price) values (?, ?)")
	mock.ExpectExec("INSERT INTO products (name, price) values (?, ?)").
		WithArgs("Product 1", 200).
		WillReturnResult(sqlmock.NewResult(1, 1))

	request, _ := http.NewRequest("POST", "/api/products", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	mysqlDB = db
	Server().ServeHTTP(response, request)

	expectedResponse := `{"data":{"type":"products","id":"1","attributes":{"name":"Product 1","price":200},"links":{"self":"http://localhost:8000/api/products/1"}}}`
	responseBody, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, 201, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)), "Response body not match")
}

func Test_GetProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "price"}).
		AddRow(1, "Product 1", 200)
	mock.ExpectQuery("SELECT id, name, price FROM products WHERE id = 1").
		WillReturnRows(rows)
	mysqlDB = db

	request, _ := http.NewRequest("GET", "/api/products/1", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	body, _ := ioutil.ReadAll(response.Body)
	expectedResponse := `{"data":{"type":"products","id":"1","attributes":{"name":"Product 1","price":200},"links":{"self":"http://localhost:8000/api/products/1"}}}`
	assert.Equal(t, http.StatusOK, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(body)))
}

func Test_UpdateProduct(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"name":  "Update product",
				"price": 250,
			},
		},
	}
	body, _ := json.Marshal(data)
	request, _ := http.NewRequest("PATCH", "/api/products/1", bytes.NewReader(body))
	response := httptest.NewRecorder()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	mock.ExpectPrepare("UPDATE products SET name = ?, price = ? WHERE id = ?")
	mock.ExpectExec("UPDATE products SET name = ?, price = ? WHERE id = ?").
		WithArgs("1", "Update product", 250).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mysqlDB = db
	Server().ServeHTTP(response, request)

	responseBody, _ := ioutil.ReadAll(response.Body)
	expectedResponse := `{"data":{"type":"products","id":"1","attributes":{"name":"Update product","price":250},"links":{"self":"http://localhost:8000/api/products/1"}}}`
	assert.Equal(t, http.StatusOK, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)), "Unexpected response")
}

func Test_DeleteProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM products WHERE id = ?").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mysqlDB = db

	request, _ := http.NewRequest("DELETE", "/api/products/1", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code, "Invalid response code")
}
