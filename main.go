package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/prometheus/promql"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(validateHandler)))
}

func validateHandler(resp http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		panic(err)
	}

	_, err := promql.ParseExpr(req.Form["query"][0])

	if err == nil {
		fmt.Fprint(resp, "valid query\n")
		return
	}

	fmt.Fprint(resp, err)
}
