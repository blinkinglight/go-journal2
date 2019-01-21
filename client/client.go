package main

import (
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
)

var (
	flagQ = flag.String("q", "", "-q query string")
	flagN = flag.Int("n", -1, "-n 10 // shows last 10 records")
)

func main() {
	flag.Parse()

	cfg, err := ini.Load(userConfig())
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	_url := cfg.Section("").Key("url").String()

	name := cfg.Section("").Key("name").String()
	content := strings.Join(os.Args[1:], " ")

	if content == "" {
		*flagQ = "today"
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

	url := fmt.Sprintf("%s/post", _url)

	r, e := http.PostForm(url, neturl.Values{
		"name":    {name},
		"content": {content},
	})

	if e != nil {
		panic(e)
	}
	defer r.Body.Close()
	println(r.StatusCode)
}

type JournalRecord struct {
	ID      int64
	Name    string
	Content string
}

func userConfig() string {
	home, _ := user.Current()
	_, err := os.Stat(home.HomeDir + "/.journal2cli.ini")
	if err != nil {
		return "/etc/journal2cli.ini"
	}
	return home.HomeDir + "/.journal2cli.ini"
}
