package main

import (
	"fmt"
	"net/http"
)

func TemplateHandler(statusCode int, name string) http.Handler {
	data, _ := Asset(fmt.Sprintf("templates/%s", name))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(data)
	})
}
