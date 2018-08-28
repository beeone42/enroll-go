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
	"encoding/json"
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

func login(w http.ResponseWriter, r *http.Request) {
	p := Page{conf, "Enroll Login", "", "", ""}
	t := template.New("Enroll")
	t = template.Must(t.ParseFiles("tmpl/login.tmpl"))
	err := t.ExecuteTemplate(w, "login", p)
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

func apiLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "{\"res\": \"ok\"}")
	return
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

func apiGetUsersByEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]
	tac.Login()
	_, body := tac.GetUsersByEmail(email)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetLastTagRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	porte_id := vars["pid"]
	event_id := vars["eid"]
	tac.Login()
	_, lt := tac.GetLastTagRead(porte_id, event_id)
	res, err := json.Marshal(lt)
	if err != nil {
		fmt.Fprintf(w, "{\"result\":\"error\"}")
	} else {
		fmt.Fprintf(w, "%s", res)
	}
	return
}

func apiGetLastTagReadEx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	porte_id1 := vars["pid1"]
	porte_id2 := vars["pid2"]
	event_id := vars["eid"]
	tac.Login()
	_, lt := tac.GetLastTagReadEx(porte_id1, porte_id2, event_id)
	res, err := json.Marshal(lt)
	if err != nil {
		fmt.Fprintf(w, "{\"result\":\"error\"}")
	} else {
		fmt.Fprintf(w, "%s", res)
	}
	return
}

func apiGetLastTagReadInfos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	porte_id := vars["pid"]
	event_id := vars["eid"]

	tac.Login()
	_, lt := tac.GetLastTagRead(porte_id, event_id)
	_, infos := tac.GetUserById(lt.UserID.UserID)
	fmt.Println("%#v", infos)
	err := json.Unmarshal([]byte(infos), &lt.Infos)
	if err != nil {
		fmt.Fprintf(w, "{\"result\":\"error\"}")
	}
	res, _ := json.Marshal(lt)
	fmt.Fprintf(w, "%s", res)
	return
}

func apiGetLastTagReadInfosEx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	porte_id1 := vars["pid1"]
	porte_id2 := vars["pid2"]
	event_id := vars["eid"]

	tac.Login()
	_, lt := tac.GetLastTagReadEx(porte_id1, porte_id2, event_id)
	_, infos := tac.GetUserById(lt.UserID.UserID)
	fmt.Println("%#v", infos)
	err := json.Unmarshal([]byte(infos), &lt.Infos)
	if err != nil {
		fmt.Fprintf(w, "{\"result\":\"error\"}")
	}
	res, _ := json.Marshal(lt)
	fmt.Fprintf(w, "%s", res)
	return
}

func ldapSearchByLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	login := vars["login"]
	search := strings.Replace("(uid={login})", "{login}", login, -1)
	fmt.Println("search: ", search)
	entries, err := ld.Search(search)
	if err != nil {
		fmt.Println("%s", err)
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

func ldapAutocomplete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	search := strings.Replace("(|(uid=*{query}*)(cn=*{query}*))", "{query}", query, -1)
	fmt.Println("search: ", search)
	entries, err := ld.Search(search)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}
	fmt.Fprintf(w, "%s", ld.JsonEntries(entries))
	return
}

func ldapEnroll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	login := vars["login"]
	rfid := vars["rfid"]
	res, err := ld.Enroll(login, rfid)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}
	fmt.Fprintf(w, "%s", res)
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

	ld.Init(conf)
	ld.Connect()
	if ld.conn != nil {
		defer ld.Close()
	} else {
		fmt.Println("LDAP Connect failed !")
	}

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	r.HandleFunc("/", dashboard)
	r.HandleFunc("/login", login)
	r.HandleFunc("/profile", searchProfile)
	r.HandleFunc("/profile/rfid/{rfid}", searchProfile)
	r.HandleFunc("/profile/login/{login}", searchProfile)
	r.HandleFunc("/api/login", apiLogin)
	r.HandleFunc("/api/ldap/bylogin/{login}", ldapSearchByLogin)
	r.HandleFunc("/api/ldap/byrfid/{rfid}", ldapSearchByRfid)
	r.HandleFunc("/api/ldap/autocomplete/{query}", ldapAutocomplete)
	r.HandleFunc("/api/ldap/enroll/{login}/{rfid}", ldapEnroll)
	r.HandleFunc("/api/tac/user/byrfid/{rfid}", apiGetUserByRfid)
	r.HandleFunc("/api/tac/user/byid/{id}", apiGetUserById)
	r.HandleFunc("/api/tac/user/byemail/{email}", apiGetUsersByEmail)
	r.HandleFunc("/api/tac/profile/byid/{id}", apiGetProfileById)
	r.HandleFunc("/api/tac/tags/byid/{id}", apiGetTagsById)
	r.HandleFunc("/api/tac/tags/bypid/{pid}/{eid}", apiGetLastTagRead)
	r.HandleFunc("/api/tac/tags/bypids/{pid1}/{pid2}/{eid}", apiGetLastTagReadEx)
	r.HandleFunc("/api/tac/events/bypid/{pid}/{eid}", apiGetLastTagReadInfos)
	r.HandleFunc("/api/tac/events/bypids/{pid1}/{pid2}/{eid}", apiGetLastTagReadInfosEx)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	fmt.Printf("Listening http://localhost:" + port + "/\n")
	http.ListenAndServe(":"+port, r)
}
