package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"runtime"
)

func main() {
	log.Println("listening for nginx connections at localhost:8081")
	log.Println("press ctrl + c to close...")
	runtime.GOMAXPROCS(runtime.NumCPU())
	addr, err := net.ResolveTCPAddr("tcp", ":8081")
	if err != nil {
		return
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	defer listener.Close()

	http.HandleFunc("/", index)

	err = fcgi.Serve(listener, nil)
	if err != nil {
		log.Println("fcgi.Serve:", err.Error())
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, header)
	fmt.Fprintf(w, "Hello world, "+req.URL.Path)
	fmt.Fprintf(w, footer)
}

var header = `
<!DOCTYPE html>
<html>
<head>
	<title> GO + FCGI </title>
	<style>	.centered {margin-left:auto;margin-right:auto;width:90%;} </style>
</head>
<body>
	<h3 class="centered"> GO + FCGI </h3>
	<div class="centered">`

var footer = `
	</div>
</body>
</html>`
