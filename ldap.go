package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/ldap.v2"
	"strings"
	"time"
)

type Ldap struct {
	server   string
	bindUser string
	bindPass string
	baseDn   string
	conn     *ldap.Conn
	last     time.Time
}

func (l *Ldap) Init(conf Configuration) {
	l.server = conf.LdapServer
	l.bindUser = conf.LdapBind
	l.bindPass = conf.LdapPassword
	l.baseDn = conf.LdapBaseDn
	l.last = time.Now()
}

func (l *Ldap) Connect() (*ldap.Conn, error) {
	d := time.Since(l.last).Seconds()
	if d > 30 {
		l.Close()
	}
	if l.conn != nil {
		return l.conn, nil
	}
	l.conn = nil
	fmt.Println("LDAP Connect")
	conn, err := ldap.Dial("tcp", l.server)
	if err != nil {
		fmt.Println("LDAP Connect FAIL")
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}
	if err := conn.Bind(l.bindUser, l.bindPass); err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}
	l.last = time.Now()
	l.conn = conn
	return conn, nil
}

func (l *Ldap) Close() {
	if l.conn == nil {
		return
	}
	fmt.Println("LDAP Close")
	l.conn.Close()
	l.conn = nil
}

func (l *Ldap) MapEntry(entry *ldap.Entry) map[string]string {
	var res map[string]string
	res = make(map[string]string)
	res["dn"] = entry.DN
	res["cn"] = entry.GetAttributeValue("cn")
	res["sn"] = entry.GetAttributeValue("sn")
	res["givenname"] = entry.GetAttributeValue("givenName")
	res["badgerfid"] = entry.GetAttributeValue("badgeRfid")
	res["badgepin"] = entry.GetAttributeValue("badgePin")
	res["uidnumber"] = entry.GetAttributeValue("uidNumber")
	res["gidnumber"] = entry.GetAttributeValue("gidNumber")
	res["loginshell"] = entry.GetAttributeValue("loginShell")
	res["alias"] = entry.GetAttributeValue("alias")
	res["close"] = entry.GetAttributeValue("close")
	return res
}

func (l *Ldap) JsonEntry(entry *ldap.Entry) string {
	res, err := json.Marshal(l.MapEntry(entry))
	if err != nil {
		return "json encoding error"
	}
	return string(res)
}

func (l *Ldap) JsonEntries(entries []*ldap.Entry) string {
	var tab []map[string]string

	for _, entry := range entries {
		tab = append(tab, l.MapEntry(entry))
	}
	res, err := json.Marshal(tab)
	if err != nil {
		return "json encoding error"
	}
	return string(res)
}

func (l *Ldap) Search(query string) ([]*ldap.Entry, error) {
	l.Connect()
	if l.conn == nil {
		return nil, nil
	}
	searchRequest := ldap.NewSearchRequest(
		l.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		query, []string{"dn", "cn", "uid",
			"uidnumber", "gidnumber",
			"sn", "givenname", "mobile", "alias", "close",
			"badgerfid", "badgepin", "loginshell"}, nil,
	)
	sr, err := l.conn.Search(searchRequest)
	if err != nil {
		fmt.Println("%s", err)
		l.Close()
		return nil, err
	}
	l.last = time.Now()
	return sr.Entries, nil
}

func (l *Ldap) GetDn(query string) (string, error) {
	entries, err := l.Search(query)
	if err != nil {
		return "", err
	}
	if len(entries) > 0 {
		res := l.MapEntry(entries[0])
		return res["dn"], nil
	}
	return "", nil
}

func (l *Ldap) Enroll(login string, rfid string) (string, error) {
	l.Connect()
	if l.conn == nil {
		return "", errors.New("connect error")
	}
	search := strings.Replace("(uid={login})", "{login}", login, -1)
	dn, err := l.GetDn(search)
	if err != nil {
		return "", err
	}
	modify := ldap.NewModifyRequest(dn)
	modify.Replace("badgeRfid", []string{rfid})
	err = l.conn.Modify(modify)
	if err != nil {
		l.Close()
		return "", err
	}
	return "ok", nil
}
