package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	_ "modernc.org/sqlite"
)

func checkError(db *sql.DB, err error) {
	if err != nil {
		message := err.Error()
		fmt.Fprintf(os.Stderr, "Error: %v\n", message)

		// Log the error in the database
		sqlStmt := `
			INSERT INTO speedtest (speed, error, date)
			VALUES (?, ?, ?)
		`
		stmt, err := db.Prepare(sqlStmt)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()

		// Insert the error message and current time into the database

		_, err = stmt.Exec(0.0, message, nowUTC())
		println("before")
		if err != nil {
			panic(err)
		}

		os.Exit(1)
	}
}

func nowUTC() string {
	currentTime := time.Now().UTC()

	// Format the time as requested
	return currentTime.Format("2006-01-02 15:04:05")
}

func getOutboundIP() (string, error) {
	url := "https://api.ipify.org?format=text" // we are using a pulib IP API, we're using ipify here, below are some others
	// https://www.ipify.org
	// http://myexternalip.com
	// http://api.ident.me
	// http://whatismyipaddress.com/api
	fmt.Printf("Getting IP address from  ipify ...\n")
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Printf("My IP is:%s\n", ip)
	return string(ip), nil
}

func main() {
	dbPath := os.Args[1]

	db, err := sql.Open("sqlite", dbPath)
	checkError(db, err)
	defer db.Close()

	sqlStmt := `
		create table if not exists speedtest (
			id INTEGER NOT NULL PRIMARY KEY,
			speed REAL,
			latency REAL,
			ip TEXT,
			remote_ip TEXT,
			remote_id TEXT,
			date TEXT,
			error TEXT
		);
	`
	_, err = db.Exec(sqlStmt)
	checkError(db, err)

	var speedtest = speedtest.New()

	serverList, _ := speedtest.FetchServers()

	targets, _ := serverList.FindServer([]int{})

	if targets.Len() < 1 {
		err = errors.New("NO_AVAILABLE_SERVERS")
		checkError(db, err)
		return

	}
	for _, s := range targets {
		checkError(db, s.PingTest(nil))
		checkError(db, s.DownloadTest())

		fmt.Printf("Latency: %s, Download: %f\n", s.Latency, s.DLSpeed)

		// Prepare the INSERT statement
		sqlStmt := `
			INSERT INTO speedtest (speed, latency, date, ip, remote_ip, remote_id)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		stmt, err := db.Prepare(sqlStmt)
		checkError(db, err)
		defer stmt.Close()

		// Define the record to insert
		speed := s.DLSpeed
		latency := s.Latency

		// Insert the record into the database
		ip, err := getOutboundIP()
		checkError(db, err)
		_, err = stmt.Exec(speed, latency, nowUTC(), ip, s.Host, s.ID)
		checkError(db, err)
	}
}
