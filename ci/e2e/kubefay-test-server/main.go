package main

import (
	"fmt"
	"net/http"
)

func echo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "kubefay")
}

func main() {
	http.HandleFunc("/", echo)
	http.ListenAndServe(":9000", nil)
}
