package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"strconv"
)

var mysqlDB *sql.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

type Product struct {
	ID    int64  `jsonapi:"primary,products"`
	Name  string `jsonapi:"attr,name"`
	Price int    `jsonapi:"attr,price"`
}

func (product Product) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("http://localhost:8000/api/products/%d", product.ID),
	}
}

func BrowseProduct(res http.ResponseWriter, _ *http.Request) {
	defer mysqlDB.Close()

	rows, err := mysqlDB.Query("SELECT id, name, price FROM products")
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

func CreateProduct(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", jsonapi.MediaType)
	defer mysqlDB.Close()

	var product Product
	err := jsonapi.UnmarshalPayload(req.Body, &product)
	if err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(res, []*jsonapi.ErrorObject{{
			Title:  "ValidationError",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
			Detail: "Given request body was invalid",
		}})
		return
	}

	query, err := mysqlDB.Prepare("INSERT INTO products (name, price) values (?, ?)")
	checkError(err)
	result, err := query.Exec(product.Name, product.Price)
	checkError(err)
	productID, err := result.LastInsertId()
	checkError(err)

	product.ID = productID
	res.WriteHeader(http.StatusCreated)
	renderJson(res, &product)
}

func DeleteProduct(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	productID := mux.Vars(req)["id"]

	defer mysqlDB.Close()

	result, err := mysqlDB.Exec("DELETE FROM products WHERE id = ?", productID)
	checkError(err)
	affected, err := result.RowsAffected()
	if affected == 0 {
		res.WriteHeader(http.StatusNotFound)
		jsonapi.MarshalErrors(res, []*jsonapi.ErrorObject{{
			Title:  "NotFound",
			Status: strconv.Itoa(http.StatusNotFound),
			Detail: fmt.Sprintf("Product with id %s not found", productID),
		}})
	}

	res.WriteHeader(http.StatusNoContent)
}

func UpdateProduct(res http.ResponseWriter, req *http.Request) {
	productID := mux.Vars(req)["id"]
	var product Product
	err := jsonapi.UnmarshalPayload(req.Body, &product)
	if err != nil {
		res.Header().Set("Content-Type", jsonapi.MediaType)
		res.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(res, []*jsonapi.ErrorObject{{
			Title:  "ValidationError",
			Detail: "Given request is invalid",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
		}})
		return
	}

	defer mysqlDB.Close()

	query, err := mysqlDB.Prepare("UPDATE products SET name = ?, price = ? WHERE id = ?")
	query.Exec(product.Name, product.Price, productID)
	checkError(err)

	product.ID, _ = strconv.ParseInt(productID, 10, 64)
	renderJson(res, &product)
}

func ShowProduct(res http.ResponseWriter, req *http.Request) {
	productID := mux.Vars(req)["id"]

	defer mysqlDB.Close()

	query, err := mysqlDB.Query("SELECT id, name, price FROM products WHERE id = " + productID)
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
	mysqlDB = connect()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/products", BrowseProduct).Methods("GET")
	router.HandleFunc("/api/products", CreateProduct).Methods("POST")
	router.HandleFunc("/api/products/{id}", DeleteProduct).Methods("DELETE")
	router.HandleFunc("/api/products/{id}", UpdateProduct).Methods("PATCH")
	router.HandleFunc("/api/products/{id}", ShowProduct).Methods("GET")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env("PORT", "8000")), router))
}
