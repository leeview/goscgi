goscgi
======

SimpleCGI protocol implementation for Go lang. Allows creation of a basic HTTP server if used with Nginx or other SCGI capable web server.

Usage
-----

~~~
package main

import (
	"bytes"
	scgi "github.com/leeview/goscgi"
	"log"
	"runtime"
)

func main() {
	log.Println("listening for nginx connections at localhost:8080")
	log.Println("press ctrl + c to close...")
	runtime.GOMAXPROCS(runtime.NumCPU())
	serv := scgi.NewServer(scgi.NewSettings())
	serv.AddHandler("/", handlerIndex)

	err := serv.ListenTcp(":8080")
	//err := serv.ListenUnix("/tmp/goscgi.socket")
	if err != nil {
		log.Println(err.Error())
	}
}

func handlerIndex(req *scgi.Request) *scgi.Response {
	var body bytes.Buffer
	body.WriteString(header)
	writeHeaders(req, &body)
	body.WriteString(footer)
	return scgi.NewResponse(scgi.RespCodeOK, scgi.RespTypeHtml, body.Bytes())
}

func writeHeaders(req *scgi.Request, buff *bytes.Buffer) {
	for k, v := range req.Header {
		buff.WriteString(k + " = " + v + "<br/>\r\n")
	}
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
~~~

Nginx configuration
-------------------
Locate Nginx configuration file. In Ubuntu it may be located at `/etc/nginx/sites-enabled/default`.
Add scgi_pass & include scgi_params directives in the root location.
~~~
location / {
	scgi_pass 127.0.0.1:8080;
	#scgi_pass unix:/tmp/goscgi.socket;
	include scgi_params;
}
~~~
If you use unix sockets, (it's slightly faster than tcp) don't forget to give write permission
to www-data (default nginx user) on the socket file (created at runtime).
The example above, uses tcp sockets and doesn't need any special treatment.
Save the config file & restart the Nginx service. In Ubuntu: `sudo service nginx restart`.
Acccess `http://localhost/anyurl`.
