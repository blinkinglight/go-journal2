package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/ini.v1"

	"server/ba"
)

var (
	flagConfig = flag.String("config", "/etc/journal2srv.ini", "-config /etc/journal2srv.ini")
)

var (
	db *bolt.DB
)

func main() {

	flag.Parse()

	cfg, err := ini.Load(*flagConfig)
	if err != nil {
		panic(err)
	}

	db, err = bolt.Open(cfg.Section("").Key("db").String(), 0755, nil)
	if err != nil {
		panic(err)
	}

	initDB()

	ba.AuthFunc = func(user, password string) bool {
		if cfg.Section("users").Haskey(user) {
			return cfg.Section("users").Key(user).String() == password
		}
		return false
	}

	http.HandleFunc("/", ba.HandlerFuncCB(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_limit := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(_limit)
		if err != nil {
			limitDefault, err := cfg.Section("").Key("limit").Int()
			if err != nil {
				limitDefault = 1000
			}
			limit = limitDefault
		}
		query := r.URL.Query().Get("query")
		var recs []JournalRecord
		if query == "" {
			recs = getRecords(limit)
		} else {
			recs = getUserRecords(query, limit)
		}
		day := -1
		for _, rec := range recs {
			ts := time.Unix(0, rec.ID)
			if day != ts.Hour() {
				day = ts.Hour()
				fmt.Fprintf(w, "<h2>%d-%02d-%02d %02d:00</h2>\n", ts.Year(), ts.Month(), ts.Day(), ts.Hour())
			}
			fmt.Fprintf(w, "%02d:%02d - <strong>%s</strong> - %s<br/>\n", ts.Hour(), ts.Minute(), rec.Name, rec.Content)
		}
	}))

	http.HandleFunc("/post", ba.HandlerFuncCB(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		content := r.URL.Query().Get("content")
		if name == "" || content == "" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		createRecord(name, content)
		w.Write([]byte("OK"))
	}))

	http.ListenAndServe(cfg.Section("").Key("bind").String(), nil)
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
