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
	"strconv"
	"strings"
	"encoding/json"
	"gopkg.in/ldap.v2"
	"crypto/rand"
)

var bank *Bank
var tac *Tac
var ctrl *Ctrl
var ld *Ldap
var ls *LdapStaff
var conf Configuration
var sessions map[string]string

type SipassConf struct {
	Cam		string
	Pid1 	string
	Pid2 	string
}

type Configuration struct {
	CaUrl        string
	CaUser       string
	CaPass       string
	PhotoUrl     string
	LdapServer   string
	LdapBind     string
	LdapPassword string
	LdapBaseDn   string
	LdapStaffServer   string
	LdapStaffBind     string
	LdapStaffPassword string
	LdapStaffBaseDn   string
	Sipass		 map[string]SipassConf
	SipassDefault string
	BankUrl		string
	BankVendor	string
	BankKey		string
}

type Page struct {
	Conf    Configuration
	Title   string
	Section string
	Rfid    string
	Login   string
	Sipass  string
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

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/img/favicon.ico")
	return
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	p := Page{conf, "Enroll", "", "", "", ""}
	t := template.New("Enroll")
	t = template.Must(t.ParseFiles("tmpl/layout.tmpl", "tmpl/dashboard.tmpl"))
	err := t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}

}

func login(w http.ResponseWriter, r *http.Request) {
	p := Page{conf, "Enroll Login", "", "", "", ""}
	t := template.New("Enroll")
	t = template.Must(t.ParseFiles("tmpl/login.tmpl"))
	err := t.ExecuteTemplate(w, "login", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}

}

func logout(w http.ResponseWriter, r *http.Request) {
	p := Page{conf, "Enroll Logout", "", "", "", ""}
	t := template.New("Enroll")
	t = template.Must(t.ParseFiles("tmpl/logout.tmpl"))
	err := t.ExecuteTemplate(w, "logout", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}

}

func sipass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sipass := vars["sipass"]
	if sipass == "" {
		sipass = conf.SipassDefault
	}
	fmt.Println("sipass: ", sipass)
	p := Page{conf, "Enroll sipass", "sipass", "", "", sipass}
	t := template.New("Enroll")
	t = template.Must(t.ParseFiles("tmpl/layout.tmpl", "tmpl/sipass.tmpl"))
	err := t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}
}

func searchProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := Page{conf, "Profile", "profile", vars["rfid"], vars["login"], ""}
	t := template.New("User Profile")
	t = template.Must(t.ParseFiles("tmpl/layout.tmpl", "tmpl/profile.tmpl"))
	err := t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		log.Fatalf("Template execution: %s", err)
	}

}

func tokenGenerator() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func checkSession(w http.ResponseWriter, r *http.Request) bool {
	res := make(map[string]string)
	bearer, exists := r.Header["Authorization"]
	if exists {
		tmp := strings.SplitN(bearer[0], " ", 2)
		if len(tmp) == 2 {
			token := tmp[1]
			_, exists = sessions[token]
			if exists {
				return true
			}
		}
	}
	res["error"] = "true"
	res["authentified"] = "false"
	res["goto"] = "/login"
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(res)
	return false
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)
	login := r.FormValue("login")
	passwd := r.FormValue("passwd")
	auth, _ := ls.Auth(login, passwd)

	if auth {
		res["auth"] = "true"
		res["token"] = tokenGenerator()
		sessions[res["token"]] = login
	} else {
		res["auth"] = "false"
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(res)

	return
}

func apiLogout(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)
	login := r.FormValue("login")
	token := r.FormValue("token")
	res["result"] = "false"

	fmt.Println(sessions)

	value, exist := sessions[token];

	if exist {
		if value == login {
			res["result"] = "true"
			delete(sessions, token)
		} else {
			res["error"] = "login / token mismatch"
		}
	} else {
		res["error"] = "unknown token"
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(res)

	return
}

func apiCheck(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	res := make(map[string]string)
	res["authentified"] = "true"
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(res)
	return
}



func bankByLogin(w http.ResponseWriter, r *http.Request) {
//	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	login := vars["login"]
	_, body := bank.GetUserInfosByLogin(login)
	fmt.Fprintf(w, "%s", body)
	return
}

func bankByRfid(w http.ResponseWriter, r *http.Request) {
//	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	rfid := vars["rfid"]
	_, body := bank.GetUserInfosByRfid(rfid)
	fmt.Fprintf(w, "%s", body)
	return
}

func bankRefundByLogin(w http.ResponseWriter, r *http.Request) {
//	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	login := vars["login"]
	refund := vars["refund"]
	_, body := bank.SetRefundByLogin(login, refund)
	fmt.Fprintf(w, "%s", body)
	return
}



func apiGetUserByRfid(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	rfid := vars["rfid"]
	tac.Login()
	_, body := tac.GetUserByTag(rfid)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetUserById(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	id := vars["id"]
	tac.Login()
	_, body := tac.GetUserById(id)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetProfileById(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	id := vars["id"]
	tac.Login()
	_, body := tac.GetProfileById(id)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetTagsById(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	id := vars["id"]
	tac.Login()
	_, body := tac.GetTagsById(id)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetUsersByEmail(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	email := vars["email"]
	tac.Login()
	_, body := tac.GetUsersByEmail(email)
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetLastTagRead(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	porte_id := vars["pid"]
	event_id := vars["eid"]
	tac.Login()
	_, lt := tac.GetLastTagRead(porte_id, event_id)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(lt)
	return
}

func apiGetLastTagReadEx(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
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
	if checkSession(w, r) != true { return }
	var entries []*ldap.Entry
	var lt TacLastTagRead
	var infos, search, r_rfid string

	vars := mux.Vars(r)
	porte_id1 := vars["pid"]
	porte_id2 := vars["pid2"]
	event_id := vars["eid"]

	tac.Login()

	if porte_id2 != "" {
		_, lt = tac.GetLastTagReadEx(porte_id1, porte_id2, event_id)
	} else {
		_, lt = tac.GetLastTagRead(porte_id1, event_id)
	}

	if lt.ID != "" {
		if len(lt.Rfid) >= 10 {
			r_rfid = lt.Rfid[0:10]
		} else {
			r_rfid = lt.Rfid
		}

		userid, err := strconv.Atoi(lt.UserID.UserID)
		if (userid > 0) {
			_, infos = tac.GetUserById(lt.UserID.UserID)
		} else {
			_, infos = tac.GetUserByTag(r_rfid)
		}



		fmt.Println("%#v", infos)
		err = json.Unmarshal([]byte(infos), &lt.Tac)
		if err != nil {
			fmt.Fprintf(w, "{\"result\":\"error\"}")
		}

		search = strings.Replace("(badgeRfid={rfid})", "{rfid}", r_rfid, -1)
		fmt.Println("ldap search rfid: ", search)

		entries, err = ld.Search(search)
		for _, entry := range entries {
			lt.Ldap = append(lt.Ldap, ld.MapEntry(entry))
		}

		if lt.Ldap != nil {
			lt.UID = strings.SplitN(strings.SplitN(lt.Ldap[0]["dn"], "uid=", 2)[1], ",", 2)[0]
		}
	}
	res, _ := json.Marshal(lt)
	fmt.Fprintf(w, "%s", res)
	return
}

func apiGetCtrlList(w http.ResponseWriter, r *http.Request) {
//	if checkSession(w, r) != true { return }
	tac.Login()
	_, body := tac.GetCtrlList()
	fmt.Fprintf(w, "%s", body)
	return
}

func apiGetCtrlSmList(w http.ResponseWriter, r *http.Request) {
//	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	host := vars["host"]
	tac.Login()
	if host != ctrl.GetHost() {
		ctrl.SetHost(host)
	}
	ctrl.Login()
	ctrl.GetSmList()
	json.NewEncoder(w).Encode(ctrl.smList)
	return
}

func ldapStaffSearchByLogin(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
	vars := mux.Vars(r)
	login := vars["login"]
	search := strings.Replace("(sAMAccountName={login})", "{login}", login, -1)
	fmt.Println("search: ", search)
	entries, err := ls.Search(search)
	if err != nil {
		fmt.Println("%s", err)
		fmt.Fprintf(w, "%s", err)
		return
	}
	fmt.Fprintf(w, "%s", ls.JsonEntries(entries))
	return
}


func ldapSearchByLogin(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r) != true { return }
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
	if checkSession(w, r) != true { return }
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
	if checkSession(w, r) != true { return }
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
	if checkSession(w, r) != true { return }
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
	bank = &Bank{}
	tac = &Tac{}
	ctrl = &Ctrl{}
	ld = &Ldap{}
	ls = &LdapStaff{}
	conf = Configuration{}
	sessions = make(map[string]string)
	err := gonfig.GetConf("config.json", &conf)
	if err != nil {
		panic(err)
	}
	fmt.Println("%#v", conf)

	bank.SetCredentials(conf.BankUrl, conf.BankVendor, conf.BankKey)
	tac.SetCredentials(conf.CaUrl, conf.CaUser, conf.CaPass)
	ctrl.SetCredentials(conf.CaUrl, tac.GetJar())

	ld.Init(conf)
	ld.Connect()
	if ld.conn != nil {
		defer ld.Close()
	} else {
		fmt.Println("LDAP Connect failed !")
	}

	ls.Init(conf)
	ls.Connect()
	if ls.conn != nil {
		defer ls.Close()
	} else {
		fmt.Println("LDAP STAFF Connect failed !")
	}

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	r.HandleFunc("/favicon.ico", favicon)
	r.HandleFunc("/", dashboard)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/profile", searchProfile)
	r.HandleFunc("/profile/rfid/{rfid}", searchProfile)
	r.HandleFunc("/profile/login/{login}", searchProfile)

	r.HandleFunc("/api/login", apiLogin).Methods("POST")
	r.HandleFunc("/api/logout", apiLogout).Methods("POST")
	r.HandleFunc("/api/check", apiCheck)

	r.HandleFunc("/api/bank/bylogin/{login}", bankByLogin)
	r.HandleFunc("/api/bank/byrfid/{rfid}", bankByRfid)
	r.HandleFunc("/api/bank/refund/{login}/{refund}", bankRefundByLogin)

	r.HandleFunc("/api/ldap/bylogin/{login}", ldapSearchByLogin)
	r.HandleFunc("/api/ldap/byrfid/{rfid}", ldapSearchByRfid)
	r.HandleFunc("/api/ldap/autocomplete/{query}", ldapAutocomplete)
	r.HandleFunc("/api/ldap/enroll/{login}/{rfid}", ldapEnroll)

	r.HandleFunc("/api/ldapstaff/bylogin/{login}", ldapStaffSearchByLogin)

	r.HandleFunc("/api/tac/ctrl", apiGetCtrlList)
	r.HandleFunc("/api/tac/ctrl/{host}", apiGetCtrlSmList)
	r.HandleFunc("/api/tac/user/byrfid/{rfid}", apiGetUserByRfid)
	r.HandleFunc("/api/tac/user/byid/{id}", apiGetUserById)
	r.HandleFunc("/api/tac/user/byemail/{email}", apiGetUsersByEmail)
	r.HandleFunc("/api/tac/profile/byid/{id}", apiGetProfileById)
	r.HandleFunc("/api/tac/tags/byid/{id}", apiGetTagsById)
	r.HandleFunc("/api/tac/tags/bypid/{pid}/{eid}", apiGetLastTagRead)
	r.HandleFunc("/api/tac/tags/bypids/{pid1}/{pid2}/{eid}", apiGetLastTagReadEx)
	r.HandleFunc("/api/tac/events/bypids/{pid}/{eid}", apiGetLastTagReadInfos)
	r.HandleFunc("/api/tac/events/bypids/{pid}/{pid2}/{eid}", apiGetLastTagReadInfos)
	r.HandleFunc("/sipass", sipass)
	r.HandleFunc("/sipass/{sipass}", sipass)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	fmt.Printf("Listening http://localhost:" + port + "/\n")
	http.ListenAndServe(":"+port, r)
}
