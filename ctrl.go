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

type Ctrl struct {
	url      	string
	host     	string
	cookie   	http.Cookie
	jar      	*cookiejar.Jar
	loggedOn 	bool
	last     	time.Time
	smList 		map[string]CtrlSmItem
}

type CtrlSm struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Smclass         string        `json:"smclass"`
	WiringSchemaURL string        `json:"wiringSchemaUrl"`
	Label           string        `json:"label"`
	Desc            string        `json:"desc"`
	Metadata        interface{}   `json:"metadata"`
}

type CtrlSmAction struct {
	Label 	string
	Script 	string
}

type CtrlSmItem struct{
	Host 	string
	Name 	string
	Smclass string
	Label   string
	Actions map[string]CtrlSmAction
}

func (c *Ctrl) GetSmList() (code int, body string) {
	var sms []CtrlSm
	var meta map[string]interface {}
	var actions map[string]interface{}
	var action map[string]interface{}
	var ok bool

	code, body = c.Request("taction_get_sm_list", []string{"sm"})
	res := c.ParseResponse(body)
	err := json.Unmarshal([]byte(res), &sms)
	if err != nil {
		fmt.Println("json decode error: %s", err.Error())
		return code, ""
	}
	i := 0
	for i < len(sms) {

		if _, ok = c.smList[sms[i].ID]; !ok {
			c.smList[sms[i].ID] = CtrlSmItem{}
		}
		s := c.smList[sms[i].ID]
		s.Host = c.GetHost()
		s.Name = sms[i].Name
		s.Smclass = sms[i].Smclass
		s.Label = sms[i].Label
		meta, ok = sms[i].Metadata.(map[string]interface {})
		if ok {
			actions, ok = meta["actions"].(map[string]interface{})
			if ok {
				s.Actions = make(map[string]CtrlSmAction)
				for k := range actions {
					action, ok = actions[k].(map[string]interface{})
					a := s.Actions[k]
					a.Label = action["label"].(string)
					a.Script = action["script"].(string)
					s.Actions[k] = a
				}
			}
		}
		c.smList[sms[i].ID] = s
		i++
	}
	fmt.Printf("smList: %v\n", c.smList)
	return code, res
}

func (c *Ctrl) SetCredentials(ctrl_url string, jar *cookiejar.Jar) {
	c.url = ctrl_url
	c.loggedOn = false
	c.jar = jar
	c.last = time.Now()
	c.smList = make(map[string]CtrlSmItem)
	fmt.Printf("ctrl url: %s...\n", c.url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = 
		&tls.Config{InsecureSkipVerify: true}
}

func (c *Ctrl) SetHost(ctrl_host string) {
	c.host = ctrl_host
	c.loggedOn = false
}

func (c *Ctrl) GetHost() (host string) {
	return c.host
}

func (c *Ctrl) RequestEx(action string, params []string,
		paramsEx map[string]string) (code int, body string) {
	var fullurl string
	fmt.Println("%s: %#v", action, params)
	client := &http.Client{ Jar: c.jar }
	fmt.Println(c.jar)
	v := url.Values{}
	v.Set("rpc[func]", action)
	i := 0
	for i < len(params) {
		v.Add(fmt.Sprintf("rpc[params][%d]", i), params[i])
		i++
	}
	for k := range paramsEx { v.Add(k, paramsEx[k])	}
	fmt.Println("v: %#v", v)
	fullurl = c.url+"action.php?url=https%3A%2F%2F"+c.host+"%2FLMC%2Faction.php"
	fmt.Println(fullurl)
	resp, err := client.PostForm(fullurl, v)
	if err != nil {
		fmt.Printf("%s failed: %s\n", action, err.Error())
		return -1, ""
	}
	//fmt.Println(resp.Status)
	defer resp.Body.Close()
	contents, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println("%s", err)
		return -2, ""
	}
	//fmt.Printf("%s\n", string(contents))
	return resp.StatusCode, string(contents)
}

func (c *Ctrl) Request(action string, params []string) (code int, body string) {
	return ctrl.RequestEx(action, params, map[string]string{})
}

func (c *Ctrl) Login() (code int, body string) {
	d := time.Since(c.last).Seconds()
	if d > 30 {
		c.loggedOn = false
	}
	if c.loggedOn {
		return 200, ""
	}
	fmt.Printf("Ctrl login in...\n")
	code, body = c.Request("taction_login", []string{})
	c.last = time.Now()
	c.loggedOn = true
	return code, body
}

func (c *Ctrl) ParseResponse(body string) string {
	r, _ := regexp.Compile(`\[(.*),\s+\]`)
	res := r.FindStringSubmatch(body)
	if res != nil {
		//for index, match := range res {
			//fmt.Printf("[%d] %s\n", index, match)
		//}
		return "["+res[1]+"]"
	}
	return "{}"
}
