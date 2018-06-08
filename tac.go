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
)

type Tac struct {
	url      string
	login    string
	passwd   string
	cookie   http.Cookie
	jar      *cookiejar.Jar
	loggedOn bool
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

func (t *Tac) SetCredentials(tac_url, login, passwd string) {
	fmt.Printf("set creds...\n")
	t.url = tac_url
	t.login = login
	t.passwd = passwd
	t.loggedOn = false
	t.jar, _ = cookiejar.New(nil)
	fmt.Printf("url: %s...\n", t.url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (t *Tac) Request(action string, params []string) (code int, body string) {
	fmt.Println("%s: %#v", action, params)

	client := &http.Client{
		Jar: t.jar,
	}

	v := url.Values{}
	v.Set("rpc[func]", action)

	i := 0
	for i < len(params) {
		v.Add(fmt.Sprintf("rpc[params][%d]", i), params[i])
		i++
	}
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
	if t.loggedOn {
		return
	}

	fmt.Printf("login in...\n")
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
	t.loggedOn = true
	return
}

func (t *Tac) ParseResponse(body string) string {
	r, _ := regexp.Compile(`\[(.*),\s+\]`)
	res := r.FindStringSubmatch(body)
	for index, match := range res {
		fmt.Printf("[%d] %s\n", index, match)
	}
	return res[1]
}

func (t *Tac) GetUserByTag(tag string) (code int, body string) {
	var p TacUserProfil
	code, body = t.Request("taction_get_user_bytag", []string{t.ReverseTag(tag)})
	res := t.ParseResponse(body)
	err := json.Unmarshal([]byte(res), &p)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
	}
	fmt.Println("%#v", p)
	return code, res
}
