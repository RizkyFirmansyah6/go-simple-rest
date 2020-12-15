package main

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
	router.HandleFunc("/api/buku", getAllBuku).Methods("GET")
	router.HandleFunc("/api/buku", addBuku).Methods("POST")
	router.HandleFunc("/api/buku/{id}", getBuku).Methods("GET")
	//router.HandleFunc("/api/products/{id}", UpdateProduct).Methods("PATCH")
	//router.HandleFunc("/api/products/{id}", DeleteProduct).Methods("DELETE")
	return router
}

func Test_getAllBuku(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' when open database", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "isbn", "judul", "pengarang"}).
		AddRow("1", "12345678", "buku baru", "rizky")
	mock.ExpectQuery("SELECT id, isbn ,judul, pengarang FROM buku").
		WillReturnRows(rows)

	mysqlDB = db
	request, _ := http.NewRequest("GET", "/api/buku", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	assert.Equal(t, 200, response.Code, "Unexpected response code")
	responseBody, _ := ioutil.ReadAll(response.Body)
	expectedResponse := `{"data":[{"type":"buku","id":"1","attributes":{"isbn":"12345678","judul":"buku baru","pengarang":"rizky"},"links":{"self":"http://localhost:8080/api/buku/1"}}],"meta":{"total":1}}`
	assert.Equal(t, string(bytes.TrimSpace(responseBody)), expectedResponse, "Response not match")
}

func Test_addBuku(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"isbn":  "87654321",
				"judul": "Teknik Informatika",
				"pengarang": "ricky",
			},
		},
	}

	body, _ := json.Marshal(data)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Logf("an error '%s' when open database", err)
	}
	defer db.Close()

	mock.ExpectPrepare("INSERT INTO buku (isbn, judul, pengarang) values (?, ?, ?)")
	mock.ExpectExec("INSERT INTO buku (isbn, judul, pengarang) values (?, ?, ?)").
		WithArgs("87654321", "Teknik Informatika", "ricky").
		WillReturnResult(sqlmock.NewResult(1, 1))

	request, _ := http.NewRequest("POST", "/api/buku", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	mysqlDB = db
	Server().ServeHTTP(response, request)

	expectedResponse := `{"data":{"type":"buku","id":"1","attributes":{"isbn":"87654321","judul":"Teknik Informatika","pengarang":"ricky"},"links":{"self":"http://localhost:8080/api/buku/1"}}}`
	responseBody, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, 201, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)), "Response body not match")
}

func Test_getBuku(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "isbn", "judul", "pengarang"}).
		AddRow(1, "12345678", "buku baru", "rizky")
	mock.ExpectQuery("SELECT id, isbn, judul, pengarang FROM buku WHERE id = 1").
		WillReturnRows(rows)
	mysqlDB = db

	request, _ := http.NewRequest("GET", "/api/buku/1", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	body, _ := ioutil.ReadAll(response.Body)
	expectedResponse := `{"data":{"type":"buku","id":"1","attributes":{"isbn":"12345678","judul":"buku baru","pengarang":"rizky"},"links":{"self":"http://localhost:8080/api/buku/1"}}}`
	assert.Equal(t, http.StatusOK, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(body)))
}

func Test_updateBuku(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"isbn":  "666",
				"judul": "test",
				"pengarang":  "ricky nyoo",
			},
		},
	}
	body, _ := json.Marshal(data)
	request, _ := http.NewRequest("PATCH", "/api/buku/1", bytes.NewReader(body))
	response := httptest.NewRecorder()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	mock.ExpectPrepare("UPDATE buku SET isbn = ?, judul = ?, pengarang = ? WHERE id = ?")
	mock.ExpectExec("UPDATE buku SET isbn = ?, judul = ?, pengarang = ? WHERE id = ?").
		WithArgs("1", "666", "test", "ricky nyoo").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mysqlDB = db
	Server().ServeHTTP(response, request)

	responseBody, _ := ioutil.ReadAll(response.Body)
	expectedResponse := `{"data":{"type":"buku","id":"1","attributes":{"isbn":"666","judul":"test","pengarang":"ricky nyoo"},"links":{"self":"http://localhost:8080/api/buku/1"}}}`
	assert.Equal(t, http.StatusOK, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)), "Unexpected response")
}

func Test_deleteBuku(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM buku WHERE id = ?").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mysqlDB = db

	request, _ := http.NewRequest("DELETE", "/api/buku/1", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code, "Invalid response code")
}