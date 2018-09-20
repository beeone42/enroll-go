package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Bank struct {
	url      string
	vendor   string
	key  	 string
}

type BankUserInfos interface{}

func (b *Bank) SetCredentials(bank_url, vendor, key string) {
	b.url = bank_url
	b.vendor = vendor
	b.key = key
	http.DefaultTransport.(*http.Transport).TLSClientConfig = 
		&tls.Config{InsecureSkipVerify: true}
}

func (b *Bank) Request(command, login, rfid, param, value string) (code int, body string) {

	client := &http.Client{}

	fmt.Println(b.url)

	v := url.Values{}
	v.Set("vendor", b.vendor)
	v.Set("key", b.key)
	v.Set("login", login)
	v.Set("rfid", rfid)

	if param != "" {
		v.Set(param, value)
	}

	fmt.Println(v)

	resp, err := client.PostForm(b.url + command, v)
	if err != nil {
		fmt.Printf("%s failed: %s\n", command, err.Error())
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

func (b *Bank) GetUserInfosByRfid(rfid string) (code int, body string) {
	var i BankUserInfos
	code, body = b.Request("balance.php", "", rfid, "", "")
	err := json.Unmarshal([]byte(body), &i)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
	}
	fmt.Println(i)
	return code, body
}

func (b *Bank) GetUserInfosByLogin(login string) (code int, body string) {
	var i BankUserInfos
	code, body = b.Request("balance.php", login, "", "", "")
	err := json.Unmarshal([]byte(body), &i)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
	}
	fmt.Println(i)
	return code, body
}

func (b *Bank) SetRefundByLogin(login string, refund string) (code int, body string) {
	code, body = b.Request("refund.php", login, "", "refund", refund)
	fmt.Println(body)
	return code, body
}
