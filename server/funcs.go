package main

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
	"time"
)

func getByDate(date string) []JournalRecord {
	parsedDate, _ := time.Parse("2006-01-02", date)

	_start := parsedDate.UnixNano()
	_end := parsedDate.UnixNano() + 24*time.Hour.Nanoseconds()

	q := fmt.Sprintf("+ID:>%d +ID:<%d", _start, _end)

	index, _ := bleve.Open(indexdb)
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.SortBy([]string{"ID"})

	searchRequest.Size = 1000
	searchResult, _ := index.Search(searchRequest)
	defer index.Close()

	var jrs []JournalRecord
	db.View(func(tx *bolt.Tx) error {
		for _, v := range searchResult.Hits {
			var jr JournalRecord
			b := tx.Bucket([]byte("records"))
			jr.Decode(b.Get([]byte(v.ID)))
			jrs = append(jrs, jr)
		}
		return nil
	})
	return jrs

}
