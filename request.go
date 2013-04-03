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
	UserAgent   string
	Content     []byte
	ContentType string
	ContentSize int64
}

const (
	GET byte = iota
	POST
	PUT
	DELETE
)

const (
	ContentTypeForm          = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm = "multipart/form-data"
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
	req := Request{}
	req.Connection = conn

	var err error
	if req.Header, err = ReadHeader(conn, settings); err != nil {
		return nil, err
	}

	var ok bool
	if contentSizeStr, ok := req.Header[ContentSizeKey]; ok && len(contentSizeStr) > 0 {
		req.ContentSize, err = strconv.ParseInt(contentSizeStr, 10, 0)
		if err != nil {
			return nil, err
		}
		if req.ContentSize > settings.MaxContentSize {
			return nil, InvalidContentErr
		}
	}

	// extract request method
	if methodStr, ok := req.Header[RequestMethodKey]; ok {
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
	if req.RawURI, ok = req.Header[RequestUriKey]; ok {
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
	req.ContentType = req.Header[ContentTypeKey]
	req.UserAgent = req.Header[HttpUserAgentKey]

	// HTTP_X_REQUESTED_WITH = XMLHttpRequest ?
	if requestedWith, ok := req.Header[RequestedWithKey]; ok {
		req.IsAJAX = (requestedWith == "XMLHttpRequest")
	}

	return &req, nil
}

func (req *Request) ReadContent(timeout time.Duration) error {
	if req.ContentSize > 0 && len(req.Content) == 0 {
		content := make([]byte, req.ContentSize)
		var alreadyRead int64
		for alreadyRead < req.ContentSize {
			req.Connection.SetReadDeadline(time.Now().Add(timeout))
			if readCnt, err := req.Connection.Read(content[alreadyRead:]); err != nil {
				return err
			} else {
				alreadyRead += int64(readCnt)
			}
		}
		req.Content = content
	}
	return nil
}

func (req *Request) ParseCookies() error {
	//TODO see go source
	return nil
}

func (req *Request) ParseForm() error {
	//TODO call ReadContent if len(req.Content) == 0
	return nil
}

func (req *Request) ParseMultipartForm() error {
	//TODO
	return nil
}
