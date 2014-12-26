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
	"bytes"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type ClosingBuffer struct {
	bytes.Buffer
}

func (buf ClosingBuffer) Close() error {
	return nil
}

type THttpTransport struct {
	conn       net.Conn
	serverConn *httputil.ServerConn
	addr       net.Addr
	timeout    time.Duration

	req   *http.Request
	rw    http.ResponseWriter // rw != nil, use req and rw; rw == nil, use conn, serverConn
	buf   ClosingBuffer
	peek  bool
	flush bool
}

func NewTHttpTransport(conn net.Conn, timeout time.Duration) *THttpTransport {
	return &THttpTransport{
		conn:       conn,
		serverConn: httputil.NewServerConn(conn, nil),
		addr:       conn.RemoteAddr(),
		timeout:    timeout,
		peek:       true,
	}
}

func NewTHttpTransportByRequest(req *http.Request, rw http.ResponseWriter) *THttpTransport {
	return &THttpTransport{
		req:  req,
		rw:   rw,
		peek: true,
	}
}

func (p *THttpTransport) SetTimeout(timeout time.Duration) error {
	p.timeout = timeout
	return nil
}

func (p *THttpTransport) pushDeadline(read, write bool) {
	var t time.Time
	if p.timeout > 0 {
		t = time.Now().Add(time.Duration(p.timeout))
	}
	if read && write {
		p.conn.SetDeadline(t)
	} else if read {
		p.conn.SetReadDeadline(t)
	} else if write {
		p.conn.SetWriteDeadline(t)
	}
}

func (p *THttpTransport) Open() error {
	return nil
}

func (p *THttpTransport) IsOpen() bool {
	if p.serverConn == nil && p.req == nil {
		return false
	}
	return true
}

func (p *THttpTransport) Close() error {
	if !p.flush {
		err := p.Flush()
		if err != nil {
			return err
		}
	}
	if p.serverConn != nil {
		err := p.serverConn.Close()
		if err != nil {
			return err
		}
		p.serverConn = nil
	}
	return nil
}

func (p *THttpTransport) Read(buf []byte) (int, error) {
	if !p.IsOpen() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	if p.rw == nil {
		p.pushDeadline(true, false)
		if p.req == nil {
			var err error
			p.req, err = p.serverConn.Read()
			if err != nil {
				return 0, NewTTransportExceptionFromError(err)
			}
		}
	}
	n, err := p.req.Body.Read(buf)
	if err != nil {
		p.peek = false
	}
	if n > 0 {
		return n, nil
	}
	return n, NewTTransportExceptionFromError(err)
}

func (p *THttpTransport) Write(buf []byte) (int, error) {
	if !p.IsOpen() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	if p.rw == nil {
		p.pushDeadline(false, true)
		return p.buf.Write(buf)
	} else {
		return p.rw.Write(buf)
	}
}

func (p *THttpTransport) Peek() bool {
	return p.peek
}

func (p *THttpTransport) Flush() error {
	if p.rw != nil {
		return nil
	}

	header := http.Header{}
	header.Add("Content-Type", "application/x-thrift")
	rsp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          &(p.buf),
		ContentLength: int64(p.buf.Len()),
		Request:       p.req,
		Close:         true,
	}
	p.flush = true
	return p.serverConn.Write(p.req, rsp)
}
