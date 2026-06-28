package handler

import (
	"fmt"
	"io"
	"net/http"
)

func GetUrlHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	id := req.PathValue("id")
	res.Header().Set("location",id)
	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Write([]byte(id))
}

func PostUrlHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res,"Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	content := req.Header.Get("content-type")
	if (content != "text/plain")&&(content != "text/plain;charset=UTF-8") {
		http.Error(res, "Only content-type: text/plain are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
        http.Error(res, err.Error(), http.StatusBadRequest)
        return
  }
	if len(body)==0 {
		http.Error(res,"URL is required",http.StatusBadRequest)
		return
	}
	url:="http://localhost:8080/EwHXdJfB"
	res.Header().Set("content-type", "text/plain")
	res.Header().Set("content-length", fmt.Sprintf("%d",len(url)))
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(url))
}