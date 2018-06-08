package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Tac struct {
	url      string
	login    string
	passwd   string
	cookie   http.Cookie
	loggedOn bool
}

func (t *Tac) SetCredentials(tac_url, login, passwd string) {
	fmt.Printf("set creds...\n")
	t.url = tac_url
	t.login = login
	t.passwd = passwd
	t.loggedOn = false
	fmt.Printf("url: %s...\n", t.url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (t *Tac) Request(action string, params []string) (code int, body string) {
	fmt.Println("%s: %#v", action, params)

	v := url.Values{}
	v.Set("rpc[func]", action)

	i := 0
	for i < len(params) {
		v.Add(fmt.Sprintf("rpc[params][%d]", i), params[i])
		i++
	}
	fmt.Println("v: %#v", v)
	resp, err := http.PostForm(t.url+"action.php", v)
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
	resp, err := http.Get(t.url)
	if err != nil {
		fmt.Printf("Login 1 failed: %s\n", err.Error())
		return
	}
	if len(resp.Cookies()) < 1 {
		fmt.Printf("No cookie\n")
		return
	}
	t.cookie = *resp.Cookies()[0]
	fmt.Printf("%s\n", t.cookie.String())

	fmt.Printf("login on %s with %s:%s creds\n", t.url, t.login, t.passwd)

	v := url.Values{}
	v.Set("rpc[func]", "taction_login")
	v.Add("rpc[params][0]", t.login)
	v.Add("rpc[params][1]", t.passwd)
	v.Add("rpc[params][2]", "db")

	fmt.Println("%#v", v)

	resp2, err2 := http.PostForm(t.url+"action.php", v)
	if err2 != nil {
		fmt.Printf("Login 2 failed: %s\n", err2.Error())
		return
	}
	fmt.Println(resp2.Status)
	t.loggedOn = true
	return
}

func (t *Tac) GetUserByTag(tag string) (code int, body string) {
	return t.Request("taction_get_user_bytag", []string{t.ReverseTag(tag)})
}
