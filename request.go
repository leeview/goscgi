// Copyright 2013 Liviu G. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goscgi

import (
	"bytes"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Connection    net.Conn
	Header        http.Header
	RawURI        string
	URL           *url.URL
	Query         url.Values
	Form          url.Values
	Files         map[string][]*multipart.FileHeader
	MultipartForm *multipart.Form
	Cookies       []*http.Cookie
	Method        byte
	IsAJAX        bool
	UserAgent     string
	Content       []byte
	ContentType   string
	ContentSize   int64
	Settings      *Settings // settings used while reading this request
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
	HttpCookieKey    = "HTTP_COOKIE"
	HttpUpgradeKey   = "HTTP_UPGRADE"
	HttpUserAgentKey = "HTTP_USER_AGENT"
)

func ReadRequest(conn net.Conn, settings *Settings) (*Request, error) {
	req := Request{}
	req.Connection = conn
	req.Settings = settings

	var err error
	if req.Header, err = ReadHeader(conn, settings); err != nil {
		return nil, err
	}

	if contentSizeStr := req.Header.Get(ContentSizeKey); len(contentSizeStr) > 0 {
		if req.ContentSize, err = strconv.ParseInt(contentSizeStr, 10, 0); err != nil {
			return nil, err
		}
		if req.ContentSize > settings.MaxContentSize {
			return nil, InvalidContentErr
		}
		if req.ContentSize > 0 {
			if contentType := req.Header.Get(ContentTypeKey); len(contentType) > 0 {
				if contentType, params, err := mime.ParseMediaType(contentType); err != nil {
					return nil, err
				} else {
					req.ContentType = contentType
					switch contentType {
					case ContentTypeForm:
						if err = req.parseForm(); err != nil {
							return nil, err
						}
					case ContentTypeMultipartForm:
						if boundary, ok := params["boundary"]; !ok {
							return nil, InvalidContentErr
						} else if err = req.parseMultipartForm(boundary); err != nil {
							return nil, err
						}
					default:
						if err = req.readContent(); err != nil {
							return nil, err
						}
					}
				}
			} else {
				return nil, InvalidHeaderErr // invalid contentType
			}
		}
	}

	// extract request method
	methodStr := req.Header.Get(RequestMethodKey)
	switch methodStr {
	case "GET":
		req.Method = GET
	case "POST":
		req.Method = POST
	case "PUT":
		req.Method = PUT
	case "DELETE":
		req.Method = DELETE
	default:
		return nil, InvalidHeaderErr // invalid method
	}

	// extract request uri & parse url + query string
	if req.RawURI = req.Header.Get(RequestUriKey); len(req.RawURI) > 0 {
		if req.URL, err = url.ParseRequestURI(req.RawURI); err != nil {
			return nil, err
		}
		if req.Query, err = url.ParseQuery(req.URL.RawQuery); err != nil {
			return nil, err
		}
	} else {
		return nil, InvalidHeaderErr
	}

	req.parseCookies()
	req.UserAgent = req.Header.Get(HttpUserAgentKey)
	req.IsAJAX = (req.Header.Get(RequestedWithKey) == "XMLHttpRequest")

	return &req, nil
}

func (req *Request) readContent() error {
	content := make([]byte, req.ContentSize)
	var alreadyRead int64
	for alreadyRead < req.ContentSize {
		req.Connection.SetReadDeadline(time.Now().Add(req.Settings.ReadTimeout))
		if readCnt, err := req.Connection.Read(content[alreadyRead:]); err != nil {
			return err
		} else {
			alreadyRead += int64(readCnt)
		}
	}
	req.Content = content
	return nil
}

func (req *Request) parseForm() error {
	var err error
	if err = req.readContent(); err != nil {
		return err
	}
	if req.Form, err = url.ParseQuery(string(req.Content)); err != nil {
		return err
	}
	return nil
}

func (req *Request) parseMultipartForm(boundary string) error {
	// can't make it work without prebuffering in memory all the multipart content !!!
	// gives error: 'multipart: Part Read: read tcp 127.0.0.1:38904: i/o timeout' everytime
	//reader := multipart.NewReader(req.Connection, boundary)
	var err error
	if err = req.readContent(); err != nil {
		return err
	}
	// we pass the content as input stream to multipart.NewReader()
	reader := multipart.NewReader(bytes.NewBuffer(req.Content), boundary)
	if multipartForm, err := reader.ReadForm(req.Settings.MaxContentSize); err != nil {
		return err
	} else {
		req.MultipartForm = multipartForm
		req.Form = multipartForm.Value
		req.Files = multipartForm.File
	}
	return nil
}

func (req *Request) parseCookies() {
	if cookies := req.Header.Get(HttpCookieKey); len(cookies) > 0 {
		parts := strings.Split(cookies, ";")
		for idx := 0; idx < len(parts); idx++ {
			if part := strings.TrimSpace(parts[idx]); len(part) > 0 {
				var cookie *http.Cookie
				if eqIdx := strings.Index(part, "="); eqIdx > 0 {
					cookie = &http.Cookie{Name: part[:eqIdx], Value: unquoteStr(part[eqIdx+1:])}
				} else {
					cookie = &http.Cookie{Name: part}
				}
				req.Cookies = append(req.Cookies, cookie)
			}
		}
	}
}

func unquoteStr(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return str
}
