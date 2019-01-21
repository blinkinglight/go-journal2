package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

var (
	flagQ = flag.String("q", "", "-q query string")
)

func main() {
	flag.Parse()
	cfg, err := ini.Load("/etc/journal2cli.ini")
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	_url := cfg.Section("").Key("url").String()

	if *flagQ != "" {
		searchRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/query?q=%s", _url, *flagQ), nil)
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

	name := cfg.Section("").Key("name").String()
	content := strings.Join(os.Args[1:], " ")

	url := fmt.Sprintf(
		"%s/post?name=%s&content=%s",
		_url,
		url.QueryEscape(name), url.QueryEscape(content),
	)
	request, _ := http.NewRequest("POST", url, nil)

	r, e := client.Do(request)
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
