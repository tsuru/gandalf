// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gandalftest provides a fake implementation of the Gandalf API.
package gandalftest

import (
	"fmt"
	"net"
)

// GandalfServer is a fake gandalf server. An instance of the client can be
// pointed to the address generated for this server
type GandalfServer struct {
	listener net.Listener
}

// NewServer returns an instance of the test server, bound to the specified
// address. To get a random port, users can specify the :0 port.
//
// Examples:
//
//     server, err := NewServer("127.0.0.1:8080") // will bind on port 8080
//     server, err := NewServer("127.0.0.1:0") // will get a random available port
func NewServer(bind string) (*GandalfServer, error) {
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	server := GandalfServer{listener: listener}
	return &server, nil
}

// Stop stops the server, cleaning the internal listener and freeing the
// allocated port.
func (s *GandalfServer) Stop() error {
	return s.listener.Close()
}

// URL returns the URL of the server, in the format "http://<host>:<port>/".
func (s *GandalfServer) URL() string {
	return fmt.Sprintf("http://%s/", s.listener.Addr())
}
