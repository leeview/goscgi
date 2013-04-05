package main

import (
	"bytes"
	//scgi "github.com/leeview/goscgi"
	scgi "goscgi"
	"log"
	"runtime"
)

func main() {
	log.Println("listening for nginx connections at localhost:8080")
	log.Println("press ctrl + c to close...")
	runtime.GOMAXPROCS(runtime.NumCPU())
	serv := scgi.NewServer(scgi.NewSettings())
	serv.AddHandler("/", index)

	err := serv.ListenTcp(":8080")
	//err := serv.ListenUnix("/tmp/goscgi.socket")
	if err != nil {
		log.Println(err.Error())
	}
}

func index(req *scgi.Request) *scgi.Response {
	var body bytes.Buffer
	body.WriteString(header)
	body.WriteString("Hello world, " + req.URL.Path)
	body.WriteString(footer)
	return scgi.NewResponse(scgi.RespCodeOK, scgi.RespTypeHtml, body.Bytes())
}

var header = `
<!DOCTYPE html>
<html>
<head>
	<title> GO + SCGI </title>
	<style>	.centered {margin-left:auto;margin-right:auto;width:90%;} </style>
</head>
<body>
	<h3 class="centered"> GO + SCGI </h3>
	<div class="centered">`

var footer = `
	</div>
</body>
</html>`
