package ipc

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type requestHandler func(Request) Response

type Server struct {
	statePath string
	token     string
	listener  net.Listener
	closeOnce sync.Once
	done      chan struct{}
}

func StartLoopbackServer(statePath string, handler func(command string) error) (*Server, error) {
	return startLoopbackServer(statePath, func(req Request) Response {
		if err := handler(req.Command); err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{OK: true}
	})
}

func StartLoopbackPayloadServer(statePath string, handler func(json.RawMessage) (json.RawMessage, error)) (*Server, error) {
	return startLoopbackServer(statePath, func(req Request) Response {
		payload, err := handler(req.Payload)
		if err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{OK: true, Payload: payload}
	})
}

func startLoopbackServer(statePath string, handler requestHandler) (*Server, error) {
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return nil, err
	}
	if err := PruneStaleState(statePath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	token, err := randomToken()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}

	server := &Server{
		statePath: statePath,
		token:     token,
		listener:  listener,
		done:      make(chan struct{}),
	}

	if err := writeEndpoint(statePath, Endpoint{
		Address: listener.Addr().String(),
		Token:   token,
		PID:     os.Getpid(),
	}); err != nil {
		_ = listener.Close()
		return nil, err
	}

	go server.serve(handler)
	return server, nil
}

func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.listener.Close()
		_ = os.Remove(s.statePath)
		<-s.done
	})
	return err
}

func (s *Server) serve(handler requestHandler) {
	defer close(s.done)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		go s.handleConn(conn, handler)
	}
}

func (s *Server) handleConn(conn net.Conn, handler requestHandler) {
	defer conn.Close()

	var req Request
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
		_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: err.Error()})
		return
	}
	if strings.TrimSpace(req.Token) != strings.TrimSpace(s.token) {
		_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: "unauthorized"})
		return
	}
	_ = json.NewEncoder(conn).Encode(handler(req))
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
