package goscgi

import (
	"net"
	"strconv"
	"time"
)

type Request struct {
	Connection  net.Conn
	Header      Header
	URI         string
	DocumentURI string
	QueryString string
	Method      byte
	IsAJAX      bool
	IsWebSocket bool
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

	// extract content type
	if contentType, ok := header[ContentTypeKey]; ok {
		req.ContentType = contentType
	}

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

	// extract request uri
	if requestUri, ok := header[RequestUriKey]; ok {
		req.URI = requestUri
	}

	// extract document uri
	if documentUri, ok := header[DocumentUriKey]; ok {
		req.DocumentURI = documentUri
	}

	// extract query string
	if queryString, ok := header[QueryStringKey]; ok {
		req.QueryString = queryString
	}

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

/*
TODO
-decode the URI & split URI parts
-decode the content if it's form-encoded
-decode multipart content using mime/multipart package
*/
