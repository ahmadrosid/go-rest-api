package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

type Product struct {
	ID			int64 `jsonapi:"primary,products"`
	Name		string `jsonapi:"attr,name"`
	Price		int `jsonapi:"attr,price"`
}

func(product Product) JSONAPILinks() *jsonapi.Links{
	return &jsonapi.Links{
		"self": fmt.Sprintf("http://localhost:8000/api/products/%d", product.ID),
	}
}

func env(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
}

func connect() *sql.DB {
	user := env("DB_USERNAME", "root")
	password := env("DB_PASSWORD", "secret")
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "3306")
	database := env("DB_DATABASE", "")

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	db, err := sql.Open("mysql", connection)
	checkError(err)
	return db
}

func checkError(err interface{}) {
	if err != nil {
		log.Print(err, "\nError connect database")
		return
	}
}

func renderJson(w http.ResponseWriter, product interface{}) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	if payload, err := jsonapi.Marshal(product); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		payloads, ok := payload.(*jsonapi.ManyPayload)
		if ok {
			val := reflect.ValueOf(product)
			payloads.Meta = &jsonapi.Meta{
				"total" : val.Len(),
			}
			json.NewEncoder(w).Encode(payloads)
		}else{
			json.NewEncoder(w).Encode(payload)
		}
	}
}

func listProduct(res http.ResponseWriter, req *http.Request) {
	db :=  connect()
	defer db.Close()

	rows, err := db.Query("SELECT id, name, price FROM products")
	checkError(err)

	var products []*Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			checkError(err)
		} else {
			products = append(products, &product)
		}
	}
	renderJson(res, products)
}

func createProduct(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", jsonapi.MediaType)
	db := connect()
	defer db.Close()

	var product Product
	err := jsonapi.UnmarshalPayload(req.Body, &product)
	if err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(res, []*jsonapi.ErrorObject{{
			Title: "ValidationError",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
			Detail: "Given request body was invalid",
		}})
		return
	}

	query, err := db.Prepare("INSERT INTO products (name, price) values (?, ?)")
	checkError(err)
	result, err := query.Exec(product.Name, product.Price)
	checkError(err)
	productID, err := result.LastInsertId()
	checkError(err)

	product.ID = productID
	res.WriteHeader(http.StatusCreated)
	renderJson(res, &product)
}

func deleteProduct(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	productID := mux.Vars(req)["id"]
	db := connect()
	defer db.Close()

	result, err := db.Exec("DELETE FROM products WHERE id = ?", productID)
	checkError(err)
	affected, err := result.RowsAffected()
	if  affected == 0 {
		res.WriteHeader(http.StatusNotFound)
		jsonapi.MarshalErrors(res, []*jsonapi.ErrorObject{{
			Title: "NotFound",
			Status: strconv.Itoa(http.StatusNotFound),
			Detail: fmt.Sprintf("Product with id %s not found", productID),
		}})
	}

	res.WriteHeader(http.StatusNoContent)
}

func updateProduct(res http.ResponseWriter, req *http.Request) {
	productID := mux.Vars(req)["id"]
	var product Product
	err := jsonapi.UnmarshalPayload(req.Body, &product)
	if err != nil {
		res.Header().Set("Content-Type", jsonapi.MediaType)
		res.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(res, []*jsonapi.ErrorObject{{
			Title: "ValidationError",
			Detail: "Given request is invalid",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
		}})
		return
	}

	db := connect()
	defer db.Close()

	query, err := db.Prepare("UPDATE products SET name = ?, price = ? WHERE id = ?")
	query.Exec(product.Name, product.Price, productID)
	checkError(err)

	product.ID, _ = strconv.ParseInt(productID, 10, 64)
	renderJson(res, &product)
}

func showProduct(res http.ResponseWriter, req *http.Request) {
	productID := mux.Vars(req)["id"]

	db := connect()
	defer db.Close()

	query, err := db.Query("SELECT id, name, price FROM products WHERE id = " + productID)
	checkError(err)
	var product Product
	for query.Next() {
		if err := query.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			log.Print(err)
		}
	}

	renderJson(res, &product)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/products", listProduct).Methods("GET")
	router.HandleFunc("/api/products", createProduct).Methods("POST")
	router.HandleFunc("/api/products/{id}", deleteProduct).Methods("DELETE")
	router.HandleFunc("/api/products/{id}", updateProduct).Methods("PATCH")
	router.HandleFunc("/api/products/{id}", showProduct).Methods("GET")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env("PORT", "8000")), router))
}