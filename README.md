goscgi
======

SimpleCGI protocol implementation for Go lang. Allows creation of a basic HTTP server if used with Nginx or other SCGI capable web server.

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
If you use unix sockets, don't forget to give write permission
to www-data (default nginx user) on the socket file (created at runtime).
The examples below, use tcp sockets and don't need any special treatment.
Save the config file & restart the Nginx service. In Ubuntu: `sudo service nginx restart`.

Usage
-----
Create a classic Go folder structure `goscgi/src`. Set $GOPATH to `path/to/goscgi`.
Run `go get github.com/leeview/goscgi` inside `goscgi` folder. Create folder `goscgi/src/main`
and copy the example below in `goscgi/src/main/main.go`. Build and run the application. Acccess `http://localhost/`.
You should see a list of HTTP header fields passed from Nginx to goscgi.
~~~
package main

import (
	"bytes"
	scgi "github.com/leeview/goscgi"
	"html"
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
	writeHeaders(req, &body)
	body.WriteString(footer)
	return scgi.NewResponse(scgi.RespCodeOK, scgi.RespTypeHtml, body.Bytes())
}

func writeHeaders(req *scgi.Request, buff *bytes.Buffer) {
	for k, values := range req.Header {
		for _, value := range values {
			buff.WriteString(k + " = " + html.EscapeString(value) + "<br/>\r\n")
		}
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

One more example
----------------
~~~
package main

import (
	"bytes"
	scgi "github.com/leeview/goscgi"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

func main() {
	log.Println("listening for nginx connections at localhost:8080")
	log.Println("press ctrl + c to close...")
	runtime.GOMAXPROCS(runtime.NumCPU())
	serv := scgi.NewServer(scgi.NewSettings())
	serv.AddHandler("/ajax", ajax)
	serv.AddHandler("/", index)

	err := serv.ListenTcp(":8080")
	//err := serv.ListenUnix("/tmp/goscgi.socket")
	if err != nil {
		log.Println(err.Error())
	}
}

func ajax(req *scgi.Request) *scgi.Response {
	if !req.IsAJAX {
		return index(req)
	}
	var body bytes.Buffer
	writeRequestInfo(req, &body)
	return scgi.NewResponse(scgi.RespCodeOK, scgi.RespTypeHtml, body.Bytes())
}

func index(req *scgi.Request) *scgi.Response {
	var body bytes.Buffer
	body.WriteString(header)
	writeRequestInfo(req, &body)
	body.WriteString(footer)
	cookie := &http.Cookie{Name: "goscgi_cookie", Value: "some cookie", MaxAge: 3600}
	return scgi.NewResponse(scgi.RespCodeOK, scgi.RespTypeHtml, body.Bytes(), cookie)
}

func writeRequestInfo(req *scgi.Request, buff *bytes.Buffer) {
	buff.WriteString("Path = " + html.EscapeString(req.URL.Path) + "<br/>\r\n")
	buff.WriteString("<h3> Query values: </h3>")
	for k, v := range req.Query {
		if len(v) > 0 {
			buff.WriteString(k + " = " + html.EscapeString(v[0]) + "<br/>\r\n")
		}
	}
	buff.WriteString("<hr/><h3> Form values: </h3>")
	for k, v := range req.Form {
		if len(v) > 0 {
			buff.WriteString(k + " = " + html.EscapeString(v[0]) + "<br/>\r\n")
		}
	}
	buff.WriteString("<hr/><h3> Cookies: </h3>")
	for _, v := range req.Cookies {
		buff.WriteString(html.EscapeString(v.String()) + "<br/>\r\n")
	}
	buff.WriteString("<hr/><h3> Files: </h3>")
	for k, v := range req.Files {
		for _, header := range v {
			file, err := header.Open()
			fileName := html.EscapeString(header.Filename)
			buff.WriteString(k + " = '" + fileName)
			if err != nil {
				buff.WriteString("' error: '" + html.EscapeString(err.Error()))
			} else {
				content, err := ioutil.ReadAll(file)
				if err != nil {
					buff.WriteString("' error: '" + html.EscapeString(err.Error()))
				} else {
					buff.WriteString("' content: '" + html.EscapeString(string(content)))
				}
			}
			buff.WriteString("'<br/>\r\n")
		}
	}
	buff.WriteString("<hr/><h3> Header values: </h3>")
	for k, values := range req.Header {
		for _, value := range values {
			buff.WriteString(k + " = " + html.EscapeString(value) + "<br/>\r\n")
		}
	}
	if req.ContentSize > 0 && len(req.Content) > 0 {
		buff.WriteString("<hr/><h3> Raw content: </h3>")
		buff.WriteString(html.EscapeString(string(req.Content)))
	}
}

var header = `
<!DOCTYPE html>
<html>
<head>
	<title> GO + SCGI </title>
	<script src="http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js"></script>
	<style>	
		.centered {margin-left:auto;margin-right:auto;width:90%;}
		.bordered {border:1px solid black;padding:5px;}
	</style>
</head>
<body>
	<h3 class="centered"> GO + SCGI </h3>
	<div class="centered">
	Simple POST:
	<form method="POST" action="/post1?arg=val&param">
		Only text here: <input type="text" name="sometext" /><br/>
		<input type="submit" value="Submit"/>
	</form><br/>
	Multipart POST:
	<form method="POST" action="/post2?arg=val&param" enctype="multipart/form-data">
		Some text here: <input type="text" name="sometext" /><br/>
		Some file here: <input type="file" name="somefile" /><br/>
		<input type="submit" value="Submit"/>
	</form><br/>
	AJAX POST: <br/>
	Some text here: <input id="ajaxText" type="text" name="sometext" value="from browser for server" /><br/>
	<button id="btnAjax">AJAX</button>
	</div>
	<br/>
	<div id="container" class="centered bordered">
`

var footer = `
	</div>
	<script>
		$(function (){
			$("#btnAjax").click(ajaxCall);
		});
		function ajaxCall(){
			$.ajax({
				type: "POST",
				url: "/ajax",
				data: {sometext:$("#ajaxText").val()}
			})
			.done(function(resp){$("#container").html(resp);})
			.fail(function(){$("#container").html("Sorry, there was an error");})
		}
	</script>
</body>
</html>`
~~~
