package main

import (
	"fmt"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
)

type TacDb struct {
	db 		 *sql.DB
	url      string
	loggedOn bool
	last     time.Time
}

type TacDbEvent struct {
	Ts		  int
	Door	  string
	Uid		  string
	Lastname  string
	Firstname string
}

func (t *TacDb) SetCredentials(tacdb_url string) {
	fmt.Printf("TAC-DB set creds...\n")
	t.url = tacdb_url
	t.loggedOn = false
	t.last = time.Now()
	fmt.Printf("url: %s...\n", t.url)
}

func (t *TacDb) Login() {
	d := time.Since(t.last).Seconds()
	if d > 30 {
		t.loggedOn = false
	}
	if t.loggedOn {
		return
	}

	fmt.Printf("TAC-DB login in...\n")
	fmt.Printf("connStr %s\n", t.url)

	db, err := sql.Open("postgres", t.url)

	if err != nil {
		t.db = nil
		fmt.Printf("TAC-DB Login error: %s\n", err.Error())
		t.loggedOn = false
	} else {
		fmt.Printf("TAC-DB Logged On !\n")
		t.db = db
		t.last = time.Now()
		t.loggedOn = true
	}
	return
}

func (t *TacDb) Logout() {
	if t.loggedOn == false {
		return
	}
	if t.db == nil {
		return
	}
	fmt.Printf("TAC-DB login out...\n")
	t.db.Close()
	t.db = nil
	t.loggedOn = false
	return

}

func (t *TacDb) Query(q string) (*sql.Rows, error) {
	rows, err := t.db.Query(q)
	if err != nil {
		fmt.Printf("TAC-DB Query error: %s\n", err.Error())
		return nil, err
	}
	return rows, nil
}
