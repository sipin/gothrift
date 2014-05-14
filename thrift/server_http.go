/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thrift

import (
	"net"
	"net/http"
	"time"
)

type TServerHttp struct {
	listener      net.Listener
	addr          net.Addr
	clientTimeout time.Duration
	interrupted   bool
}

func NewTServerHttp(listenAddr string) (*TServerHttp, error) {
	return NewTServerHttpTimeout(listenAddr, 0)
}

func NewTServerHttpTimeout(listenAddr string, clientTimeout time.Duration) (*TServerHttp, error) {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &TServerHttp{addr: addr, clientTimeout: clientTimeout}, nil
}

func (p *TServerHttp) Listen() error {
	if p.IsListening() {
		return nil
	}
	l, err := net.Listen(p.addr.Network(), p.addr.String())
	if err != nil {
		return err
	}
	p.listener = l
	return nil
}

func (p *TServerHttp) Accept() (TTransport, error) {
	if p.interrupted {
		return nil, errTransportInterrupted
	}
	if p.listener == nil {
		return nil, NewTTransportException(NOT_OPEN, "No underlying server socket")
	}
	conn, err := p.listener.Accept()
	if err != nil {
		return nil, NewTTransportExceptionFromError(err)
	}
	return NewTHttpTransport(conn, p.clientTimeout), nil
}

// Checks whether the socket is listening.
func (p *TServerHttp) IsListening() bool {
	return p.listener != nil
}

// Connects the socket, creating a new socket object if necessary.
func (p *TServerHttp) Open() error {
	if p.IsListening() {
		return NewTTransportException(ALREADY_OPEN, "Server socket already open")
	}
	if l, err := net.Listen(p.addr.Network(), p.addr.String()); err != nil {
		return err
	} else {
		p.listener = l
	}
	return nil
}

func (p *TServerHttp) Addr() net.Addr {
	return p.addr
}

func (p *TServerHttp) Close() error {
	defer func() {
		p.listener = nil
	}()
	if p.IsListening() {
		return p.listener.Close()
	}
	return nil
}

func (p *TServerHttp) Interrupt() error {
	p.interrupted = true
	return nil
}

type ErrorHandler interface {
	HandleError(error)
}

type HttpHandler struct {
	processorFactory      TProcessorFactory
	inputProtocolFactory  TProtocolFactory
	outputProtocolFactory TProtocolFactory
	errorHandler          ErrorHandler
}

func NewHttpHandler(processor TProcessor, inputProtocolFactory, outputProtocolFactory TProtocolFactory, errorHandler ErrorHandler) *HttpHandler {
	return &HttpHandler{NewTProcessorFactory(processor), inputProtocolFactory, outputProtocolFactory, errorHandler}
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	trans := NewTHttpTransportByRequest(r, w)
	inputTransport := trans
	outputTransport := trans
	processor := h.processorFactory.GetProcessor(trans)
	inputProtocol := h.inputProtocolFactory.GetProtocol(inputTransport)
	outputProtocol := h.outputProtocolFactory.GetProtocol(outputTransport)
	if inputTransport != nil {
		defer inputTransport.Close()
	}
	if outputTransport != nil {
		defer outputTransport.Close()
	}
	for {
		ok, err := processor.Process(inputProtocol, outputProtocol)
		if err, ok := err.(TTransportException); ok && err.TypeId() == END_OF_FILE {
			return
		} else if err != nil {
			if h.errorHandler != nil {
				h.errorHandler.HandleError(err)
			}
			return
		}
		if !ok {
			break
		}
	}
}
