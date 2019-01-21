package main

import (
	"bytes"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
	"os"
)

func initDB() {
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("index"))
		tx.CreateBucketIfNotExists([]byte("records"))
		return nil
	})

	_, err := os.Stat(indexdb)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		idx, err := bleve.New(indexdb, mapping)
		if err != nil {
			panic(err)
		}
		db.View(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte("records")).Cursor()
			for _, v := c.First(); v != nil; _, v = c.Next() {
				var jr JournalRecord
				jr.Decode(v)
				idx.Index(string(jr.Id()), jr)
			}
			return nil
		})
		idx.Close()
	}
}

func createRecord(name, content string) {
	db.Update(func(tx *bolt.Tx) error {
		jr := NewJR(name, content)
		tx.Bucket([]byte("records")).Put(jr.Id(), jr.Encode())
		tx.Bucket([]byte("index")).Put(jr.Index(), jr.Id())

		idx, err := bleve.Open(indexdb)
		defer idx.Close()
		if err != nil {
			panic(err)
		}
		return idx.Index(string(jr.Id()), jr)
	})
}

func getUserRecords(user string, limit int) []JournalRecord {
	num := 0
	var records []JournalRecord
	prefix := []byte(user)
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("index")).Cursor()
		jrb := tx.Bucket([]byte("records"))
		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			num++
			if num >= limit {
				return nil
			}
			var jr JournalRecord
			jr.Decode(jrb.Get(v))
			records = append(records, jr)
		}
		return nil
	})
	return records
}

func getRecords(limit int) []JournalRecord {
	num := 0
	var records []JournalRecord
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("records")).Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			num++
			if num >= limit {
				return nil
			}
			var jr JournalRecord
			jr.Decode(v)
			records = append(records, jr)
		}
		return nil
	})
	return records
}
