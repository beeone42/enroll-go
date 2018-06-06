package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Openfile, err := os.Open("index.html")
		defer Openfile.Close()
		if err != nil {
			http.Error(w, "File WTF not found.", 401)
			return
		}
		io.Copy(w, Openfile)
		return
	})

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

	fmt.Printf("Listening http://localhost:8080/\n")
	http.ListenAndServe(":8080", r)
}
