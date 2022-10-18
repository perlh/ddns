package main

import (
	"fmt"
	"net/http"
)

func startHttp() {
	http.HandleFunc("/dns", handler1203)
	http.ListenAndServe(":8080", nil)
}

func handler1203(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, DomainTypeA)
}
