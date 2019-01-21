package main

import (
	"fmt"
	// "html"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("/etc/journal2cli.ini")
	if err != nil {
		panic(err)
	}

	name := cfg.Section("").Key("name").String()
	content := strings.Join(os.Args[1:], " ")
	url := fmt.Sprintf(
		"%s/post?name=%s&content=%s",
		cfg.Section("").Key("url").String(),
		url.QueryEscape(name), url.QueryEscape(content),
	)
	request, _ := http.NewRequest("POST", url, nil)
	client := &http.Client{}
	r, e := client.Do(request)
	if e != nil {
		panic(e)
	}
	defer r.Body.Close()
	println(r.StatusCode)
}
