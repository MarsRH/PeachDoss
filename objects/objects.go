package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var BASIC_FILE_URL, _ = os.Getwd()

func ObjectHandle(w http.ResponseWriter, r *http.Request) {

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

	log.Default().Panicln(r.URL.EscapedPath())
	filename := BASIC_FILE_URL + "/files/" + strings.Split(r.URL.EscapedPath(), "/")[2]

	f, err := os.Create(filename)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	io.Copy(f, r.Body)
}

func getObjects(w http.ResponseWriter, r *http.Request) {

	log.Default().Println(r.URL.EscapedPath())
	filename := BASIC_FILE_URL + "/files/" + strings.Split(r.URL.EscapedPath(), "/")[2]

	f, err := os.Open(filename)
	if err != nil {
		log.Default().Println(err)
		return
	}
	defer f.Close()

	io.Copy(w, f)
}
