package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ldap.v2"
	"time"
	"crypto/tls"
)

type LdapStaff struct {
	server   string
	bindUser string
	bindPass string
	baseDn   string
	conn     *ldap.Conn
	last     time.Time
}

func (l *LdapStaff) Init(conf Configuration) {
	l.server = conf.LdapStaffServer
	l.bindUser = conf.LdapStaffBind
	l.bindPass = conf.LdapStaffPassword
	l.baseDn = conf.LdapStaffBaseDn
	l.last = time.Now()
}

func (l *LdapStaff) Auth(login, passwd string) (bool, error) {
	l.Close()
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	conn, err := ldap.DialTLS("tcp", l.server, tlsConfig)
	if err != nil {
		return false, fmt.Errorf("Failed to connect. %s", err)
	}
	if err := conn.Bind(login, passwd); err != nil {
		return false, fmt.Errorf("Failed to bind. %s", err)
	}
	return true, nil
}

func (l *LdapStaff) Connect() (*ldap.Conn, error) {
	d := time.Since(l.last).Seconds()
	if d > 30 {
		l.Close()
	}
	if l.conn != nil {
		return l.conn, nil
	}
	l.conn = nil
	fmt.Println("LDAP STAFF Connect")
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	conn, err := ldap.DialTLS("tcp", l.server, tlsConfig)
	if err != nil {
		fmt.Println("LDAP STAFF Connect FAIL")
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}
	if err := conn.Bind(l.bindUser, l.bindPass); err != nil {
		fmt.Println("LDAP STAFF Bind FAIL")
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}
	l.last = time.Now()
	l.conn = conn
	return conn, nil
}

func (l *LdapStaff) Close() {
	if l.conn == nil {
		return
	}
	fmt.Println("LDAP STAFF Close")
	l.conn.Close()
	l.conn = nil
}

func (l *LdapStaff) MapEntry(entry *ldap.Entry) map[string]string {
	var res map[string]string
	res = make(map[string]string)
	res["uid"] = entry.GetAttributeValue("sAMAccountName")
	res["dn"] = entry.DN
	res["cn"] = entry.GetAttributeValue("cn")
	res["sn"] = entry.GetAttributeValue("sn")
	res["givenname"] = entry.GetAttributeValue("givenName")
	return res
}

func (l *LdapStaff) JsonEntry(entry *ldap.Entry) string {
	res, err := json.Marshal(l.MapEntry(entry))
	if err != nil {
		return "json encoding error"
	}
	return string(res)
}

func (l *LdapStaff) JsonEntries(entries []*ldap.Entry) string {
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

func (l *LdapStaff) Search(query string) ([]*ldap.Entry, error) {
	l.Connect()
	if l.conn == nil {
		return nil, nil
	}
	searchRequest := ldap.NewSearchRequest(
		l.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		query, []string{"sAMAccountName", "dn", "cn", "sn", "givenname"}, nil,
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

func (l *LdapStaff) GetDn(query string) (string, error) {
	entries, err := l.Search(query)
	if err != nil {
		return "error", err
	}
	if len(entries) > 0 {
		res := l.MapEntry(entries[0])
		return res["dn"], nil
	}
	return "", nil
}
