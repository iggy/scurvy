package msgs

import (
	"encoding/json"
)

// NatsMsg - a msg that we send to the Nats MQ bus
type NatsMsg interface {
	serialize() []byte
}

// NewDownload - struct for Marshal'ing new download messages
type NewDownload struct {
	Name string
	Path string
}

func (s NewDownload) serialize() []byte {
	b, _ := json.Marshal(s)
	return b
}

// FailedDownload - struct for Marshal'ing new download messages
type FailedDownload struct {
	Name string
	Path string
}

func (s FailedDownload) serialize() []byte {
	b, _ := json.Marshal(s)
	return b
}

// DiskFull - struct for Marshal'ing disk full messages from SABNZBD
type DiskFull struct {
	Message string
}

func (s DiskFull) serialize() []byte {
	b, _ := json.Marshal(s)
	return b
}

// ReportFiles = struct for Marshal'ing the report files messages
type ReportFiles struct {
	Full    bool
	Changed bool
}

func (s ReportFiles) serialize() []byte {
	b, _ := json.Marshal(s)
	return b
}
