package main

import (
	"fmt"
	"net/http"
	"pronghorn-app/apiv1"
)

func main() {
	fmt.Println("Starting http server....")

	mux := http.NewServeMux()

	apiv1.StartRoutes(mux)

	http.ListenAndServe(":8080", mux)
}
