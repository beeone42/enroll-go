package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
)

type Tac struct {
	url    string
	login  string
	passwd string
	cookie http.Cookie
}

func (t *Tac) SetCredentials(tac_url, login, passwd string) {
	fmt.Printf("set creds...\n")
	t.url = tac_url
	t.login = login
	t.passwd = passwd
	fmt.Printf("url: %s...\n", t.url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (t Tac) Login() {
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
	v.Add("[0]", t.login)
	v.Add("[1]", t.passwd)

	resp2, err2 := http.PostForm(t.url, v)
	if err2 != nil {
		fmt.Printf("Login 2 failed: %s\n", err2.Error())
		return
	}
	fmt.Printf(resp2.Status)
	return
}
