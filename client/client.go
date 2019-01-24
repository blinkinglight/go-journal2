package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"os/user"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

var (
	flagQ = flag.String("q", "", "-q query string")
	flagN = flag.Int("n", -1, "-n 10 // shows last 10 records")
	flagM = flag.Bool("m", false, "-m")
	flagS = flag.Bool("sync", false, "-sync")
)

var (
	db *bolt.DB
)

func main() {
	flag.Parse()

	cfg, err := ini.Load(userConfig())
	if err != nil {
		panic(err)
	}

	db, err = bolt.Open(journalDir()+"/que.db", 0755, nil)
	if err != nil {
		panic(err)
	}
	qInit()

	client := &http.Client{}
	_url := cfg.Section("").Key("url").String()

	name := cfg.Section("").Key("name").String()
	content := strings.Join(os.Args[1:], " ")

	if content == "" {
		*flagM = true
	}

	if *flagM {
		searchRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/query?q=Name:%s", _url, neturl.QueryEscape(name)), nil)
		response, err := client.Do(searchRequest)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()

		var jrs []JournalRecord
		json.NewDecoder(response.Body).Decode(&jrs)
		for _, jr := range jrs {
			ts := time.Unix(0, jr.ID)
			fmt.Printf("%s %s %s\n", ts.Format(time.Stamp), jr.Name, jr.Content)
		}
		os.Exit(0)
	}

	if *flagQ != "" {
		searchRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/query?q=%s", _url, neturl.QueryEscape(*flagQ)), nil)
		response, err := client.Do(searchRequest)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()

		var jrs []JournalRecord
		json.NewDecoder(response.Body).Decode(&jrs)
		for _, jr := range jrs {
			ts := time.Unix(0, jr.ID)
			fmt.Printf("%s %s %s\n", ts.Format(time.Stamp), jr.Name, jr.Content)
		}
		os.Exit(0)
	}

	if *flagN != -1 {
		searchRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/latest?n=%d", _url, *flagN), nil)
		response, err := client.Do(searchRequest)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()

		var jrs []JournalRecord
		json.NewDecoder(response.Body).Decode(&jrs)
		for _, jr := range jrs {
			ts := time.Unix(0, jr.ID)
			fmt.Printf("%s %s %s\n", ts.Format(time.Stamp), jr.Name, jr.Content)
		}
		os.Exit(0)
	}

	if *flagS {
		for jr := qPeek(); jr.ID != 0; jr = qPeek() {
			_url := fmt.Sprintf("%s/raw", _url)
			rdr := bytes.NewReader(jr.Encode())
			rsp, err := http.Post(_url, "application/json", rdr)
			if err != nil {
				fmt.Printf("Error. Sync failed.\n")
				break
			}
			rsp.Body.Close()
			if rsp.StatusCode >= 200 && rsp.StatusCode <= 299 {
				qDelete(jr.ID)
			}
			if rsp.StatusCode == 401 {
				fmt.Println("Unauthorized")
				break
			}
		}
		fmt.Println("Sync done.")
		os.Exit(0)
	}

	jr := JournalRecord{
		ID:      time.Now().UnixNano(),
		Name:    name,
		Content: content,
	}

	qAdd(jr)
	fmt.Println("Message added to local queue")

	var lastError error
	for jr := qPeek(); jr.ID != 0; jr = qPeek() {
		_url := fmt.Sprintf("%s/raw", _url)
		rdr := bytes.NewReader(jr.Encode())
		rsp, err := http.Post(_url, "application/json", rdr)
		if err != nil {
			lastError = err
			break
		}
		rsp.Body.Close()
		if rsp.StatusCode >= 200 && rsp.StatusCode <= 299 {
			qDelete(jr.ID)
		}
		if rsp.StatusCode == 401 {
			lastError = errors.New("Unauthorized")
			fmt.Println("Unauthorized")
			break
		}
	}
	if lastError != nil {
		fmt.Printf("Something went wrong. run '%s -sync' manually\n", os.Args[0])
	} else {
		fmt.Println("Sync done")
	}
}

type JournalRecord struct {
	ID      int64
	Name    string
	Content string
}

func (jr *JournalRecord) Encode() []byte {
	b, _ := json.Marshal(jr)
	return b
}

func userConfig() string {
	home, _ := user.Current()
	_, err := os.Stat(home.HomeDir + "/.journal2cli.ini")
	if err != nil {
		return "/etc/journal2cli.ini"
	}
	return home.HomeDir + "/.journal2cli.ini"
}

func journalDir() string {
	dir, _ := user.Current()
	jrnldir := dir.HomeDir + "/.journal2"
	if fi, err := os.Stat(jrnldir); err != nil || !fi.IsDir() {
		os.MkdirAll(jrnldir, 0755)
	}
	return jrnldir
}

func qInit() {
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("que"))
		return nil
	})
}

func qAdd(jr JournalRecord) {
	db.Update(func(tx *bolt.Tx) error {
		tx.Bucket([]byte("que")).Put(itob(int(jr.ID)), jr.Encode())
		return nil
	})
}

func qPeek() JournalRecord {
	var jr JournalRecord
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("que")).Cursor()
		_, v := c.First()
		json.Unmarshal(v, &jr)
		return nil
	})
	return jr
}

func qDelete(id int64) {
	db.Update(func(tx *bolt.Tx) error {
		tx.Bucket([]byte("que")).Delete(itob(int(id)))
		return nil
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
