package main

import (
	"fmt"
	"net/http"
)

var dnslogRecode []string

func dnslog(w http.ResponseWriter, r *http.Request) {

	if len(dnslogRecode) > 1000 {
		dnslogRecode = []string{}
	}
	fmt.Fprintln(w, dnslogRecode)
}
