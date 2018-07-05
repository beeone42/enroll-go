package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ldap.v2"
)

type Ldap struct {
	server   string
	bindUser string
	bindPass string
	baseDn   string
	conn     *ldap.Conn
}

func (l *Ldap) Connect(conf Configuration) (*ldap.Conn, error) {

	l.server = conf.LdapServer
	l.bindUser = conf.LdapBind
	l.bindPass = conf.LdapPassword
	l.baseDn = conf.LdapBaseDn

	conn, err := ldap.Dial("tcp", l.server)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}
	if err := conn.Bind(l.bindUser, l.bindPass); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}
	l.conn = conn
	return conn, nil
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
	searchRequest := ldap.NewSearchRequest(
		l.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		query, []string{"dn", "cn", "uid",
			"uidnumber", "gidnumber",
			"sn", "givenname", "mobile",
			"badgerfid", "badgepin", "loginshell"}, nil,
	)
	sr, err := l.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	return sr.Entries, nil
}
