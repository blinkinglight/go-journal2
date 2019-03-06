package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"gopkg.in/ini.v1"

	"server/ba"
)

var (
	flagConfig = flag.String("config", "/etc/journal2srv.ini", "-config /etc/journal2srv.ini")
)

var (
	db      *bolt.DB
	indexdb string
)

func main() {

	flag.Parse()

	// cfg, err := ini.Load(*flagConfig)
	cfg, err := ini.LoadSources(ini.LoadOptions{
		KeyValueDelimiters: "=",
	}, *flagConfig)
	if err != nil {
		panic(err)
	}

	db, err = bolt.Open(cfg.Section("").Key("db").String(), 0755, nil)
	if err != nil {
		panic(err)
	}

	indexdb = cfg.Section("").Key("indexdb").String()

	initDB()

	ba.AuthFunc = func(perm, user, password string) bool {
		pair := user + ":" + password
		if cfg.Section("users").Haskey(pair) {
			perms := cfg.Section("users").Key(pair).In("", []string{"r", "w", "rw"})
			return strings.Contains(perms, perm)
		}
		return false
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", ba.HandlerFuncCB("r", func(w http.ResponseWriter, r *http.Request) {
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
		tpl0 := `<script>document.write((new Date(Date.UTC(%d,%d,%d,%d,%d,%d))).toLocaleDateString('en-GB', {day: 'numeric', month: 'numeric', year: 'numeric', hour: 'numeric'}).replace(/\//g, '-')+":00");</script>` + "\n"
		tpl1 := `<script>document.write((new Date(Date.UTC(%d,%d,%d,%d,%d,%d))).toLocaleDateString('en-GB', {hour: 'numeric', minute: 'numeric', second:'numeric'}).split(" ")[1]);</script>` + "\n"
		// tpl1 := `<script>document.write((new Date(Date.UTC(%d,%d,%d,%d,%d,%d))).toISOString().slice(0,19).split("T")[1]);</script>` + "\n"
		fmt.Fprintln(w, "<html>")
		fmt.Fprintln(w, "<body>")
		for _, rec := range recs {
			ts := time.Unix(0, rec.ID).UTC()
			if day != ts.Hour() {
				day = ts.Hour()
				// fmt.Fprintf(w, "<h2>%d-%02d-%02d %02d:00</h2>\n", ts.Year(), ts.Month(), ts.Day(), ts.Hour())
				fmt.Fprintf(w, `<h2>%s</h2>`, fmt.Sprintf(tpl0, ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second()))
			}
			_time := fmt.Sprintf(tpl1, ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
			fmt.Fprintf(w, "%s - <strong>%s</strong> - %s<br/>\n", _time, html.EscapeString(rec.Name), html.EscapeString(rec.Content))
		}
	}))

	mux.HandleFunc("/post", ba.HandlerFuncCB("w", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		content := r.FormValue("content")
		if name == "" || content == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		createRecord(-1, name, content)
		w.Write([]byte("OK"))
	}))

	mux.HandleFunc("/query", ba.HandlerFuncCB("r", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		index, _ := bleve.Open(indexdb)
		if q == "today" {
			_end := Day(time.Now(), 1)
			_start := TruncateDate(_end)
			q = fmt.Sprintf("+ID:>%d +ID:<%d", _start.UnixNano(), _end.UnixNano())
		}
		if q == "yesterday" {
			_end := Day(time.Now(), 0)
			_start := TruncateDate(_end)
			q = fmt.Sprintf("+ID:>%d +ID:<%d", _start.UnixNano(), _end.UnixNano())
		}
		query := bleve.NewQueryStringQuery(q)
		searchRequest := bleve.NewSearchRequest(query)
		if q == "today" || q == "yesterday" {
			searchRequest.SortBy([]string{"-ID"})
		} else {
			searchRequest.SortBy([]string{"ID"})
		}
		searchRequest.Size = 1000
		searchResult, _ := index.Search(searchRequest)
		defer index.Close()
		// fmt.Println(searchResult)
		var jrs []JournalRecord
		db.View(func(tx *bolt.Tx) error {
			for _, v := range searchResult.Hits {
				var jr JournalRecord
				// fmt.Printf("%#v %#v\n", k, v.ID)
				b := tx.Bucket([]byte("records"))
				jr.Decode(b.Get([]byte(v.ID)))
				jrs = append(jrs, jr)
			}
			return nil
		})
		// fmt.Printf("%#v\n", searchResult)
		json.NewEncoder(w).Encode(jrs)
	}))

	mux.HandleFunc("/latest", ba.HandlerFuncCB("r", func(w http.ResponseWriter, r *http.Request) {
		_limit := r.URL.Query().Get("n")
		limit, err := strconv.Atoi(_limit)
		if err != nil {
			limit = 10
		}

		recs := getRecords(limit)

		json.NewEncoder(w).Encode(recs)

	}))

	mux.HandleFunc("/date", ba.HandlerFuncCB("r", func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("d")

		jrs := getByDate(date)

		json.NewEncoder(w).Encode(jrs)
	}))

	mux.HandleFunc("/raw", ba.HandlerFuncCB("w", func(w http.ResponseWriter, r *http.Request) {
		var jr JournalRecord
		json.NewDecoder(r.Body).Decode(&jr)
		createRecord(jr.ID, jr.Name, jr.Content)
	}))
	go startSlackBot()
	println("Starting server...")
	log.Fatal(http.ListenAndServe(cfg.Section("").Key("bind").String(), logger(mux)))
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := ClientIp(r)
		if client.XForIP != "" {
			log.Printf(`[%s](%s) %s %s`, client.IP, client.XForIP, r.URL.String(), r.Method)
		} else {
			log.Printf(`[%s] %s %s`, client.IP, r.URL.String(), r.Method)

		}
		next.ServeHTTP(w, r)
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func Day(t time.Time, d int) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day+d, 0, 0, 0, 0, t.Location())
}

func TruncateDate(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

type Addr struct {
	IP     string
	XForIP string
}

func ClientIp(r *http.Request) (ip *Addr) {
	ip = new(Addr)
	remoteAddr := r.RemoteAddr
	idx := strings.LastIndex(remoteAddr, ":")
	if idx != -1 {
		remoteAddr = remoteAddr[0:idx]
		if remoteAddr[0] == '[' && remoteAddr[len(remoteAddr)-1] == ']' {
			remoteAddr = remoteAddr[1 : len(remoteAddr)-1]
		}
	}
	ip.IP = remoteAddr
	ip.XForIP = r.Header.Get("X-Forwarded-For")
	return
}
