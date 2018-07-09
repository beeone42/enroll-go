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
	"strings"
)

var tac *Tac
var ld *Ldap
var conf Configuration

type Configuration struct {
	CaUrl        string
	CaUser       string
	CaPass       string
	PhotoUrl     string
	LdapServer   string
	LdapBind     string
	LdapPassword string
	LdapBaseDn   string
}

type Page struct {
	Conf    Configuration
	Title   string
	Section string
	Rfid    string
	Login   string
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

func dashboard(w http.ResponseWriter, r *http.Request) {
	p := Page{conf, "Enroll", "", "", ""}
	t := template.New("Enroll")
	t = template.Must(t.ParseFiles("tmpl/layout.tmpl", "tmpl/dashboard.tmpl"))
	err := t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}

}

func searchProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := Page{conf, "Profile", "profile", vars["rfid"], vars["login"]}
	t := template.New("User Profile")
	t = template.Must(t.ParseFiles("tmpl/layout.tmpl", "tmpl/profile.tmpl"))
	err := t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}

}

func apiGetUserByRfid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rfid := vars["rfid"]
	tac.Login()
	_, body := tac.GetUserByTag(rfid)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetUserById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	tac.Login()
	_, body := tac.GetUserById(id)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetProfileById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	tac.Login()
	_, body := tac.GetProfileById(id)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetTagsById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	tac.Login()
	_, body := tac.GetTagsById(id)
	fmt.Fprintf(w, "%s", body)
	return
}

func ldapSearchByLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	login := vars["login"]
	search := strings.Replace("(uid={login})", "{login}", login, -1)
	fmt.Println("search: ", search)
	entries, err := ld.Search(search)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}
	fmt.Fprintf(w, "%s", ld.JsonEntries(entries))
	return
}

func ldapSearchByRfid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rfid := vars["rfid"]
	search := strings.Replace("(badgeRfid={rfid})", "{rfid}", rfid, -1)
	fmt.Println("search: ", search)
	entries, err := ld.Search(search)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}
	fmt.Fprintf(w, "%s", ld.JsonEntries(entries))
	return
}

func main() {
	r := mux.NewRouter()
	tac = &Tac{}
	ld = &Ldap{}
	conf = Configuration{}
	err := gonfig.GetConf("config.json", &conf)
	if err != nil {
		panic(err)
	}
	fmt.Println("%#v", conf)

	tac.SetCredentials(conf.CaUrl, conf.CaUser, conf.CaPass)

	ld.Connect(conf)
	if ld.conn != nil {
		defer ld.conn.Close()
	}

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	r.HandleFunc("/", dashboard)
	r.HandleFunc("/profile", searchProfile)
	r.HandleFunc("/profile/rfid/{rfid}", searchProfile)
	r.HandleFunc("/profile/login/{login}", searchProfile)
	r.HandleFunc("/api/ldap/bylogin/{login}", ldapSearchByLogin)
	r.HandleFunc("/api/ldap/byrfid/{rfid}", ldapSearchByRfid)
	r.HandleFunc("/api/tac/user/byrfid/{rfid}", apiGetUserByRfid)
	r.HandleFunc("/api/tac/user/byid/{id}", apiGetUserById)
	r.HandleFunc("/api/tac/profile/byid/{id}", apiGetProfileById)
	r.HandleFunc("/api/tac/tags/byid/{id}", apiGetTagsById)

	fmt.Printf("Listening http://localhost:8080/\n")
	http.ListenAndServe(":8080", r)
}
