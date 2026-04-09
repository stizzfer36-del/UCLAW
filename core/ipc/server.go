// Package ipc provides a Unix-socket IPC server so the CLI and
// external processes can call into the running UCLAW daemon.
package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

// Request is the JSON payload sent by callers.
type Request struct {
	Method string            `json:"method"`
	Params map[string]string `json:"params"`
}

// Response is the JSON payload returned.
type Response struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// Handler maps method names to Go handlers.
type Handler func(params map[string]string) (interface{}, error)

var handlers = map[string]Handler{}

// Register registers a method handler.
func Register(method string, h Handler) {
	handlers[method] = h
}

// ListenAndServe starts the IPC server on the given socket path.
func ListenAndServe(socketPath string) error {
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("ipc: listen %s: %w", socketPath, err)
	}
	log.Printf("[ipc] listening on %s", socketPath)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[ipc] accept error: %v", err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			writeResp(conn, Response{Error: "bad JSON"})
			continue
		}
		h, ok := handlers[req.Method]
		if !ok {
			writeResp(conn, Response{Error: fmt.Sprintf("unknown method %q", req.Method)})
			continue
		}
		data, err := h(req.Params)
		if err != nil {
			writeResp(conn, Response{Error: err.Error()})
		} else {
			writeResp(conn, Response{OK: true, Data: data})
		}
	}
}

func writeResp(conn net.Conn, r Response) {
	b, _ := json.Marshal(r)
	b = append(b, '\n')
	_, _ = conn.Write(b)
}
