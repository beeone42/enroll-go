package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tkanos/gonfig"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

var tac *Tac

type Configuration struct {
	CaUrl  string
	CaUser string
	CaPass string
}

type Page struct {
	Title string
}

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
	tac = &Tac{}
	conf := Configuration{}
	err := gonfig.GetConf("config.json", &conf)
	if err != nil {
		panic(err)
	}
	fmt.Println("%#v", conf)

	tac.SetCredentials(conf.CaUrl, conf.CaUser, conf.CaPass)

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { sendFile(w, "index.html") })
	r.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		p := Page{"Profile"}
		t := template.New("User Profile")
		t = template.Must(t.ParseFiles("tmpl/layout.tmpl", "tmpl/profile.tmpl"))
		err := t.ExecuteTemplate(w, "layout", p)
		if err != nil {
			log.Fatalf("Template execution: %s", err)
		}
	})
	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("api\n")
		tac.Login()
		_, body := tac.GetUserByTag("1234567890")
		// fmt.Fprintf(w, "ok %d %s", res, body)
		fmt.Fprintf(w, "%s", body)
		return
	})

	fmt.Printf("Listening http://localhost:8080/\n")
	http.ListenAndServe(":8080", r)
}
