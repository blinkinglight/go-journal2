package main

import (
	"encoding/json"
	"time"
)

func NewJR(name, content string) *JournalRecord {
	jr := &JournalRecord{
		ID:      time.Now().UTC().UnixNano(),
		Name:    name,
		Content: content,
	}
	return jr
}

type JournalRecord struct {
	ID      int64
	Name    string
	Content string
}

func (jr *JournalRecord) Id() []byte {
	return itob(int(jr.ID))
}

func (jr *JournalRecord) Decode(b []byte) {
	json.Unmarshal(b, jr)
}

func (jr *JournalRecord) Encode() []byte {
	b, _ := json.Marshal(jr)
	return b
}

func (jr *JournalRecord) Index() []byte {
	tsb := itob(int(jr.ID))
	prefix := []byte(jr.Name)
	prefix = append(prefix, tsb...)
	return prefix
}
