package main

import (
	"encoding/json"
	"github.com/google/jsonapi"
	"log"
	"net/http"
	"os"
	"reflect"
)

func env(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
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
				"total": val.Len(),
			}
			json.NewEncoder(w).Encode(payloads)
		} else {
			json.NewEncoder(w).Encode(payload)
		}
	}
}
