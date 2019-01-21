#GoLang http basic auth middleware

```go
func main() {
	ba.SetUserPassword("user", "password")

	http.HandleFunc("/", ba.Auth(indexfn))

	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func indexfn(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi!"))
}
```

```
curl localhost:8080
```