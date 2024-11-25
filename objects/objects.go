package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const BASIC_FILE_URL = "./testfiles/"

func ObjectHandle(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method + r.URL.EscapedPath())

	switch r.Method {
	//Get Method
	case http.MethodGet:
		getObjects(w, r)
	//Put Method
	case http.MethodPut:
		saveObjects(w, r)
	//Other Method
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func saveObjects(w http.ResponseWriter, r *http.Request) {

	filename := BASIC_FILE_URL + strings.Split(r.URL.EscapedPath(), "/")[2]

	f, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	io.Copy(f, r.Body)
}

func getObjects(w http.ResponseWriter, r *http.Request) {

	filename := BASIC_FILE_URL + strings.Split(r.URL.EscapedPath(), "/")[2]

	f, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()

	io.Copy(w, f)
}
