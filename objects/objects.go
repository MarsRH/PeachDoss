package objects

import "net/http"

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
	// TODO:
}

func getObjects(w http.ResponseWriter, r *http.Request) {
	// TODO:
}
