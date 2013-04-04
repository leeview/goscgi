// Copyright 2013 Liviu G. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goscgi

import (
	"net"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	ResponseCode []byte
	ContentType  []byte
	Content      []byte
	Cookies      []*http.Cookie
	Header       http.Header
}

var (
	status        = []byte("Status: ")
	contentType   = []byte("Content-Type: ")
	contentLength = []byte("Content-Length: ")
	crlf          = []byte{0x0d, 0x0a}
	sep           = []byte{':'}

	RespTypeHtml = []byte("text/html")
	RespTypeText = []byte("text/plain")
	RespTypeJson = []byte("text/json")

	RespCodeOK            = []byte("200 OK")
	RespCodeNotFound      = []byte("404 Not found")
	RespCodeBadRequest    = []byte("400 Bad request")
	RespCodeInternalError = []byte("500 Internal error")
)

func NewResponse(respCode, contentType []byte, content []byte, cookies ...*http.Cookie) *Response {
	resp := Response{}
	resp.ResponseCode = respCode
	resp.ContentType = contentType
	resp.Content = content
	resp.Header = http.Header{}
	for _, cookie := range cookies {
		resp.SetCookie(cookie)
	}
	return &resp
}

func (resp *Response) SetCookie(cookie *http.Cookie) {
	if cookie != nil {
		resp.Header.Add("Set-Cookie", cookie.String())
	}
}

func (resp *Response) Write(conn net.Conn, timeout time.Duration) error {
	var err error
	// set a timeout for the first write
	// just in case the user closed the connection
	conn.SetWriteDeadline(time.Now().Add(timeout))
	if _, err = conn.Write(status); err != nil {
		return err
	}
	conn.Write(resp.ResponseCode)
	conn.Write(crlf)
	conn.Write(contentType)
	conn.Write(resp.ContentType)
	conn.Write(crlf)
	contentSize := int64(len(resp.Content))
	if contentSize > 0 {
		conn.Write(contentLength)
		conn.Write([]byte(strconv.FormatInt(contentSize, 10)))
		conn.Write(crlf)
	}

	/*for k, v := range resp.Header {
		conn.Write([]byte(k))
		conn.Write(sep)
		conn.Write([]byte(v))
		conn.Write(crlf)
	}*/
	if err = resp.Header.Write(conn); err != nil {
		return err
	}

	conn.Write(crlf)
	if contentSize > 0 {
		// set a timeout for sending (large?) content
		conn.SetWriteDeadline(time.Now().Add(timeout))
		if _, err = conn.Write(resp.Content); err != nil {
			return err
		}
	}
	return nil
}
