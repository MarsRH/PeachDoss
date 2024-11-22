package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MarsRH/PeachDoss/objects"
)

const HOST = "127.0.0.1"
const PORT = "8096"
const ADDR = HOST + ":" + PORT

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Default().Println(r.Header)
		w.Write([]byte("Hello PeachDoss"))
	})

	http.HandleFunc("/handleObjcts", objects.ObjectHandle)

	fmt.Println("Running On " + ADDR)
	http.ListenAndServe(ADDR, nil)
}
