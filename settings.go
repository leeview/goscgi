package goscgi

import "time"

type Settings struct {
	MaxHeaderSize  int
	MaxContentSize int64
	ListenTimeout  time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

func NewSettings() *Settings {
	return &Settings{
		42 * 1024,       //	MaxHeaderSize 42 KB = (max 4KB/cookie) * (max 10 cookies) + 2KB headers
		4 * 1024 * 1024, //	MaxContentSize 4 MB/request
		3 * time.Second, // ListenTimeout = the max duration listener.Accept() stays blocked waiting for a connection
		5 * time.Second, // ReadTimeout 5sec * 1MB/sec -> we can receive max 5MB on a 1MB downlink before timeout !!!
		5 * time.Second, // WriteTimeout 5sec * 1MB/sec -> we can deliver max 5MB on a 1MB uplink before timeout !!!
	}
}
