// Copyright 2013 Liviu G. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goscgi

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Request struct {
	Connection  net.Conn
	Header      Header
	RawURI      string
	URL         *url.URL
	Query       url.Values
	Cookies     []*http.Cookie
	Method      byte
	IsAJAX      bool
	IsWebSocket bool
	UserAgent   string
	Content     []byte
	ContentType string
}

const (
	GET byte = iota
	POST
	PUT
	DELETE
)

const (
	ContentSizeKey   = "CONTENT_LENGTH"
	ContentTypeKey   = "CONTENT_TYPE"
	RequestMethodKey = "REQUEST_METHOD"
	RequestUriKey    = "REQUEST_URI"
	DocumentUriKey   = "DOCUMENT_URI"
	DocumentRootKey  = "DOCUMENT_ROOT"
	QueryStringKey   = "QUERY_STRING"
	RemoteAddrKey    = "REMOTE_ADDR"
	RemotePortKey    = "REMOTE_PORT"
	RequestedWithKey = "HTTP_X_REQUESTED_WITH"
	HttpUpgradeKey   = "HTTP_UPGRADE"
	HttpUserAgentKey = "HTTP_USER_AGENT"
)

func ReadRequest(conn net.Conn, settings *Settings) (*Request, error) {
	header, err := ReadHeader(conn, settings)
	if err != nil {
		return nil, err
	}

	// parse content size & read content
	var content []byte
	if contentSizeStr, ok := header[ContentSizeKey]; ok && len(contentSizeStr) > 0 {
		contentSize, err := strconv.ParseInt(contentSizeStr, 10, 0)
		if err != nil {
			return nil, err
		}
		if contentSize > settings.MaxContentSize {
			return nil, InvalidContentErr
		}
		if contentSize > 0 {
			content = make([]byte, contentSize)
			var alreadyRead int64
			for alreadyRead < contentSize {
				conn.SetReadDeadline(time.Now().Add(settings.ReadTimeout))
				if readCnt, err := conn.Read(content[alreadyRead:]); err != nil {
					return nil, err
				} else {
					alreadyRead += int64(readCnt)
				}
			}
		}
	}

	req := Request{}
	req.Connection = conn
	req.Header = header
	req.Content = content

	var ok bool
	// extract request method
	if methodStr, ok := header[RequestMethodKey]; ok {
		switch methodStr {
		case "GET":
			req.Method = GET
		case "POST":
			req.Method = POST
		case "PUT":
			req.Method = PUT
		case "DELETE":
			req.Method = DELETE
		}
	}

	// extract request uri & parse url
	if req.RawURI, ok = header[RequestUriKey]; ok {
		if req.URL, err = url.ParseRequestURI(req.RawURI); err != nil {
			return nil, err
		}
		if req.Query, err = url.ParseQuery(req.URL.RawQuery); err != nil {
			return nil, err
		}
	} else {
		return nil, InvalidHeaderErr
	}

	// extract content type & user agent
	req.ContentType = header[ContentTypeKey]
	req.UserAgent = header[HttpUserAgentKey]

	//TODO parse Cookies

	// HTTP_X_REQUESTED_WITH = XMLHttpRequest ?
	if requestedWith, ok := header[RequestedWithKey]; ok {
		req.IsAJAX = (requestedWith == "XMLHttpRequest")
	}

	// HTTP_UPGRADE = WebSocket ?
	if upgrade, ok := header[HttpUpgradeKey]; ok {
		req.IsWebSocket = (upgrade == "websocket" || upgrade == "WebSocket")
	}
	return &req, nil
}

func (req *Request) ParseForm() error {
	//TODO
	return nil
}

func (req *Request) ParseMultipartForm() error {
	//TODO
	return nil
}
