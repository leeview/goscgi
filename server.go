// Copyright 2013 Liviu G. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goscgi

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Settings  *Settings
	Handlers  []Handler
	Close     chan bool
	WaitGroup sync.WaitGroup
}

type Handler struct {
	Path string
	Func HandlerFunc
}

type HandlerFunc func(*Request) *Response

var (
	RespNotFound      = NewResponse(RespCodeNotFound, RespTypeText, RespCodeNotFound)
	RespBadRequest    = NewResponse(RespCodeBadRequest, RespTypeText, RespCodeBadRequest)
	RespInternalError = NewResponse(RespCodeInternalError, RespTypeText, RespCodeInternalError)
)

func NewServer(s *Settings) *Server {
	srv := &Server{}
	srv.Settings = s
	srv.Close = make(chan bool)
	return srv
}

func (srv *Server) AddHandler(path string, handler HandlerFunc) {
	srv.Handlers = append(srv.Handlers, Handler{path, handler})
}

func (srv *Server) ListenTcp(port string) error {
	addr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	defer srv.WaitGroup.Wait()
	srv.listenLoop(listener)
	return nil
}

func (srv *Server) ListenUnix(addrStr string) error {
	addr, err := net.ResolveUnixAddr("unix", addrStr)
	if err != nil {
		return err
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	defer srv.WaitGroup.Wait()
	srv.listenLoop(listener)
	return nil
}

type anyListener interface {
	Accept() (net.Conn, error)
	SetDeadline(time.Time) error
}

func (srv *Server) listenLoop(listener anyListener) {
	osInterrupt := make(chan os.Signal, 1)
	signal.Notify(osInterrupt, os.Interrupt)
	listenTimeout := srv.Settings.ListenTimeout
	for {
		select {
		case <-srv.Close:
			log.Println("Server.listenLoop, <-srv.Close: closing...")
			return
		case <-osInterrupt:
			log.Println("Server.listenLoop, <-osInterrupt: terminating...")
			return
		default: // if default -> go and execute the rest of the loop
		}

		listener.SetDeadline(time.Now().Add(listenTimeout))
		conn, err := listener.Accept()
		if err == nil {
			srv.WaitGroup.Add(1)
			go srv.handleConn(conn)
		} /* else {
			log.Println("Server.listenLoop, listener.Accept():", err.Error())
		}*/
	}
}

func (srv *Server) handleConn(conn net.Conn) {
	defer srv.WaitGroup.Done()
	defer conn.Close()
	req, err := ReadRequest(conn, srv.Settings)
	if err != nil {
		log.Println("Server.handleConn, ReadRequest:", err.Error())
		err = RespBadRequest.Write(conn, srv.Settings.WriteTimeout)
		if err != nil {
			log.Println("Server.handleConn, RespBadRequest.Send:", err.Error())
		}
	} else {
		srv.handleReq(req)
	}
}

func (srv *Server) handleReq(req *Request) {
	var err error
	timeout := srv.Settings.WriteTimeout
	if handler := srv.getHandler(req.URL.Path); handler != nil {
		if resp := handler(req); resp != nil {
			err = resp.Write(req.Connection, timeout)
		} else {
			err = RespInternalError.Write(req.Connection, timeout)
		}
	} else {
		err = RespNotFound.Write(req.Connection, timeout)
	}
	if err != nil {
		log.Println("Server.handleReq:", err.Error())
	}
}

func (srv *Server) getHandler(path string) HandlerFunc {
	for _, handler := range srv.Handlers {
		if strings.HasPrefix(path, handler.Path) {
			return handler.Func
		}
	}
	return nil
}
