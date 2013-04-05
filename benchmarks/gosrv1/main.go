package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"runtime"
)

func main() {
	log.Println("listening for nginx connections at localhost:8082")
	log.Println("press ctrl + c to close...")
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/", index)

	err := http.ListenAndServe(":8082", nil)
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
	<title> GO + HTTP </title>
	<style>	.centered {margin-left:auto;margin-right:auto;width:90%;} </style>
</head>
<body>
	<h3 class="centered"> GO + HTTP </h3>
	<div class="centered">`

var footer = `
	</div>
</body>
</html>`
