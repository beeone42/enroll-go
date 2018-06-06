package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

var tac Tac

func sendFile(w http.ResponseWriter, f string) {
	Openfile, err := os.Open(f)
	defer Openfile.Close()
	if err != nil {
		http.Error(w, "File not found.", 404)
		return
	}
	io.Copy(w, Openfile)
	return
}

func main() {
	r := mux.NewRouter()

	(&tac).SetCredentials("https://tac.domain.com/GMC/action.php", "tac_login", "tac_passwd")

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { sendFile(w, "index.html") })

	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("api\n")
		tac.Login()
		fmt.Fprintf(w, "ok")
		return
	})

	fmt.Printf("Listening http://localhost:8080/\n")
	http.ListenAndServe(":8080", r)
}
