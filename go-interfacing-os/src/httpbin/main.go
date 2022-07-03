package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func printRequest(w http.ResponseWriter, r *http.Request) {
	req, err := httputil.DumpRequest(r, true)
	if err == nil {
		fmt.Printf(string(req))
		fmt.Println("")
	} else {
		fmt.Print(err)
	}

	http.Error(w, "OK", 200)
}

func main() {
	fmt.Printf("Listening on 8000\n")
	http.HandleFunc("/", printRequest)
	http.ListenAndServe(":8000", nil)
}
