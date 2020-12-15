package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var mysqlDB *sql.DB

//func init() {
//	err := godotenv.Load()
//	if err != nil {
//		log.Fatal("Error loading .env file")
//	}
//}

// Buku Struct (model)
type Buku struct {
	Id 			int64 `jsonapi:"primary,buku"`
	Isbn 		string `jsonapi:"attr,isbn"`
	Judul 		string `jsonapi:"attr,judul"`
	Pengarang 	string `jsonapi:"attr,pengarang"`
}

func (buku Buku) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("http://localhost:8080/api/buku/%d", buku.Id),
	}
}

func main() {
	// Init DB Connection
	mysqlDB = connect()
	defer mysqlDB.Close()

	// Init Router
	router := mux.NewRouter().StrictSlash(true)

	// Router Handlers / Endpoints
	router.HandleFunc("/api/buku",getAllBuku).Methods("GET")
	router.HandleFunc("/api/buku/{nim}",getBuku).Methods("GET")
	router.HandleFunc("/api/buku",addBuku).Methods("POST")
	router.HandleFunc("/api/buku/{id}",updateBuku).Methods("PUT")
	router.HandleFunc("/api/buku/{id}",deleteBuku).Methods("DELETE")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env("PORT", "8080")), router))
}

func deleteBuku(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	bukuID := mux.Vars(request)["id"]

	result, err := mysqlDB.Exec("DELETE FROM buku WHERE id = ?", bukuID)
	checkError(err)
	affected, err := result.RowsAffected()
	if affected == 0 {
		writer.WriteHeader(http.StatusNotFound)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title:  "NotFound",
			Status: strconv.Itoa(http.StatusNotFound),
			Detail: fmt.Sprintf("Product with id %s not found", bukuID),
		}})
	}

	writer.WriteHeader(http.StatusNoContent)
}

func updateBuku(writer http.ResponseWriter, request *http.Request) {
	bukuID := mux.Vars(request)["id"]
	var buku Buku
	err := jsonapi.UnmarshalPayload(request.Body, &buku)
	if err != nil {
		writer.Header().Set("Content-Type", jsonapi.MediaType)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title:  "ValidationError",
			Detail: "Given request is invalid",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
		}})
		return
	}

	query, err := mysqlDB.Prepare("UPDATE buku SET isbn = ?, judul = ?, pengarang = ? WHERE id = ?")
	query.Exec(buku.Isbn, buku.Judul, buku.Pengarang, bukuID)
	checkError(err)

	buku.Id, _ = strconv.ParseInt(bukuID, 10, 64)
	renderJson(writer, &buku)
}

func getBuku(writer http.ResponseWriter, request *http.Request) {
	bukuID := mux.Vars(request)["id"]

	query, err := mysqlDB.Query("SELECT id, isbn, judul, pengarang FROM buku WHERE id = " + bukuID)
	checkError(err)
	var buku Buku
	for query.Next() {
		if err := query.Scan(&buku.Id, &buku.Isbn, &buku.Judul, &buku.Pengarang); err != nil {
			log.Print(err)
		}
	}

	renderJson(writer, &buku)
}

func addBuku(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", jsonapi.MediaType)

	var buku Buku
	err := jsonapi.UnmarshalPayload(request.Body, &buku)
	if err != nil {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title:  "ValidationError",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
			Detail: "Given request body was invalid",
		}})
		return
	}

	query, err := mysqlDB.Prepare("INSERT INTO buku (isbn, judul, pengarang) values (?, ?, ?)")
	checkError(err)
	result, err := query.Exec(buku.Isbn, buku.Judul, buku.Pengarang)
	checkError(err)
	productID, err := result.LastInsertId()
	checkError(err)

	buku.Id = productID
	writer.WriteHeader(http.StatusCreated)
	renderJson(writer, &buku)
}

func getAllBuku(writer http.ResponseWriter, _ *http.Request) {
	rows, err := mysqlDB.Query("SELECT id, isbn ,judul, pengarang FROM buku")
	checkError(err)

	var buku []*Buku
	log.Print(rows)
	for rows.Next() {
		var bku Buku
		if err := rows.Scan(&bku.Id, &bku.Isbn, &bku.Judul, &bku.Pengarang); err != nil && err != sql.ErrNoRows {
			checkError(err)
		} else {
			buku = append(buku, &bku)
		}
	}
	renderJson(writer, buku)
}
