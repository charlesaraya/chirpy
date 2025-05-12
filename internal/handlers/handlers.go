package handlers

import (
	"net/http"
)

const (
	HealthOK string = "OK"
)

func GetHealth(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte(HealthOK))
}
