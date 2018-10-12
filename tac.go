package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"
)

type Tac struct {
	url      string
	login    string
	passwd   string
	cookie   http.Cookie
	jar      *cookiejar.Jar
	loggedOn bool
	last     time.Time
}

type TacUserProfil struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Company  string      `json:"company"`
	Valid    interface{} `json:"valid"`
	Expired  interface{} `json:"expired"`
	Disabled string      `json:"disabled"`
	Apbypass string      `json:"apbypass"`
}

type TacUserInfos struct {
	Firstname  string      `json:"firstname"`
	Lastname   string      `json:"lastname"`
	Prettyname string      `json:"prettyname"`
	Company    string      `json:"company"`
	Phone      string      `json:"phone"`
	Email      string      `json:"email"`
	Validity   interface{} `json:"validity"`
	Expiry     interface{} `json:"expiry"`
	Valid      interface{} `json:"valid"`
	Expired    interface{} `json:"expired"`
	Disabled   string      `json:"disabled"`
	Apbypass   string      `json:"apbypass"`
}

type TacLastTagRead struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Tag    string `json:"tag"`
	Rfid   string
	Pin    string
	PID    string
	UID    string
	UserID struct {
		UserID string `json:"user_id"`
	} `json:"user_id"`
	Tac TacUserInfos
	Ldap []map[string]string
}

func (t *Tac) GetJar() (jar *cookiejar.Jar) {
	return t.jar
}

func (t *Tac) SetCredentials(tac_url, login, passwd string) {
	fmt.Printf("set creds...\n")
	t.url = tac_url
	t.login = login
	t.passwd = passwd
	t.loggedOn = false
	t.jar, _ = cookiejar.New(nil)
	t.last = time.Now()
	fmt.Printf("url: %s...\n", t.url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = 
		&tls.Config{InsecureSkipVerify: true}
}

func (t *Tac) RequestEx(action string, params []string,
		paramsEx map[string]string) (code int, body string) {
	fmt.Println("%s: %#v", action, params)
	client := &http.Client{ Jar: t.jar }
	v := url.Values{}
	v.Set("rpc[func]", action)
	i := 0
	for i < len(params) {
		v.Add(fmt.Sprintf("rpc[params][%d]", i), params[i])
		i++
	}
	for k := range paramsEx { v.Add(k, paramsEx[k])	}
	fmt.Println("v: %#v", v)
	resp, err := client.PostForm(t.url+"action.php", v)
	if err != nil {
		fmt.Printf("%s failed: %s\n", action, err.Error())
		return -1, ""
	}
	fmt.Println(resp.Status)
	defer resp.Body.Close()
	contents, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println("%s", err)
		return -2, ""
	}
	fmt.Printf("%s\n", string(contents))
	return resp.StatusCode, string(contents)
}

func (t *Tac) Request(action string, params []string) (code int, body string) {
	return tac.RequestEx(action, params, map[string]string{})
}

func (t *Tac) ReverseTag(tag string) string {
	res := ""
	i := 0
	for i < len(tag) {
		res = tag[i:i+2] + res
		i += 2
	}
	return res
}

func (t *Tac) Login() {
	d := time.Since(t.last).Seconds()
	if d > 30 {
		t.loggedOn = false
	}
	if t.loggedOn {
		return
	}

	fmt.Printf("TAC login in...\n")
	fmt.Printf("get %s\n", t.url)

	client := &http.Client{
		Jar: t.jar,
	}
	resp, err := client.Get(t.url)
	if err != nil {
		fmt.Printf("Login 1 failed: %s\n", err.Error())
		return
	}

	fmt.Printf("login on %s with %s:%s creds: %s\n", t.url, t.login, t.passwd, resp.Status)

	v := url.Values{}
	v.Set("rpc[func]", "taction_login")
	v.Add("rpc[params][0]", t.login)
	v.Add("rpc[params][1]", t.passwd)
	v.Add("rpc[params][2]", "db")

	fmt.Println("%#v", v)

	resp2, err2 := client.PostForm(t.url+"action.php", v)
	if err2 != nil {
		fmt.Printf("Login 2 failed: %s\n", err2.Error())
		return
	}
	fmt.Println(resp2.Status)
	t.last = time.Now()
	t.loggedOn = true
	return
}

func (t *Tac) ParseResponse(body string) string {
	r, _ := regexp.Compile(`\[(.*),\s+\]`)
	res := r.FindStringSubmatch(body)
	if res != nil {
		for index, match := range res {
			fmt.Printf("[%d] %s\n", index, match)
		}
		return res[1]
	}
	return "{}"
}

func (t *Tac) GetUserByTag(tag string) (code int, body string) {
	var p TacUserProfil

	if (tag == "") {
		return -1, "tag empty"
	}
	code, body = t.Request("taction_get_user_bytag", []string{t.ReverseTag(tag)})
	res := t.ParseResponse(body)
	err := json.Unmarshal([]byte(res), &p)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
	}
	fmt.Println("%#v", p)
	return code, res
}

func (t *Tac) GetUserById(id string) (code int, body string) {
	var i TacUserInfos
	code, body = t.Request("taction_get_user_info", []string{id})
	res := t.ParseResponse(body)
	err := json.Unmarshal([]byte(res), &i)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
	}
	fmt.Println("%#v", i)
	return code, res
}

func (t *Tac) GetProfileById(id string) (code int, body string) {
	code, body = t.RequestEx("taction_get_profile_byuser", []string{id},
		map[string]string{"sort": "name", "dir": "ASC"})
	res := fmt.Sprintf("[%s]", t.ParseResponse(body))
	return code, res
}

func (t *Tac) GetTagsById(id string) (code int, body string) {
	code, body = t.Request("taction_get_tag_byuser", []string{id})
	res := fmt.Sprintf("[%s]", t.ParseResponse(body))
	return code, res
}

func (t *Tac) GetUsersByProfile(id string) (code int, body string) {
	code, body = t.RequestEx("taction_get_user_byprofile", []string{id},
		map[string]string{"sort": "name", "dir": "ASC"})
	res := t.ParseResponse(body)
	fmt.Println("%#v", body)
	return code, res
}

func (t *Tac) GetUsersByEmail(email string) (code int, body string) {
	code, body = t.RequestEx("taction_get_user_list", []string{},
		map[string]string{
			"start":                  "0",
			"stop":                   "1",
			"sort":                   "name",
			"dir":                    "ASC",
			"filter[0][field]":       "email",
			"filter[0][data][type]":  "string",
			"filter[0][data][value]": email,
		})
	res := fmt.Sprintf("[%s]", t.ParseResponse(body))
	return code, res
}

func (t *Tac) GetLastTagRead(porte_id, event_id string) (code int, lt TacLastTagRead) {
	var body string
	code, body = t.Request("taction_get_last_tag_read",
		[]string{porte_id, event_id})
	res := t.ParseResponse(body)
	err := json.Unmarshal([]byte(res), &lt)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
	}
	if len(lt.Tag) > 10 {
		lt.Rfid = t.ReverseTag(lt.Tag[0:10])
	} else {
		lt.Rfid = t.ReverseTag(lt.Tag)
	}
	if len(lt.Tag) == 14 {
		lt.Pin = t.ReverseTag(lt.Tag[10:14])
	}
	lt.PID = porte_id
	return code, lt
}

func (t *Tac) GetLastTagReadEx(porte_id1, porte_id2, event_id string) (code int, lt TacLastTagRead) {
	code1, lt1 := tac.GetLastTagRead(porte_id1, event_id)
	code2, lt2 := tac.GetLastTagRead(porte_id2, event_id)
	if lt1.ID > lt2.ID {
		return code1, lt1
	} else {
		return code2, lt2
	}
}

func (t *Tac) GetCtrlList() (interface{}) {
	var l []interface{}

	_, body := t.RequestEx("taction_get_ctrl_list", []string{},
		map[string]string{
			"start":                  "0",
			"limit":                  "99",
			"sort":                   "name",
			"dir":                    "ASC",
		})
	res := fmt.Sprintf("[%s]", t.ParseResponse(body))
	err := json.Unmarshal([]byte(res), &l)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
		return nil
	}

	return l
}
