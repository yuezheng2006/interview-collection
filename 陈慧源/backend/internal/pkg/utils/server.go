package utils

import (
	"log"
	"net/http"
)

func MustStartServer(handler http.Handler) {
	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
