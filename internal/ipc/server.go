package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct { Action string `json:"action"` }
type Response struct { OK bool `json:"ok"`; Status string `json:"status"` }
func Serve(ctx context.Context, socketPath string) error { network, address := endpoint(socketPath); if network == "unix" { _ = os.Remove(address) }; listener, err := net.Listen(network, address); if err != nil { return err }; defer listener.Close(); done := make(chan struct{}); go func() { <-ctx.Done(); listener.Close(); if network == "unix" { _ = os.Remove(address) }; close(done) }(); for { conn, err := listener.Accept(); if err != nil { select { case <-ctx.Done(): <-done; return nil; default: return err } }; go handle(conn) } }
func Ping(ctx context.Context, socketPath string) (Response, error) { var d net.Dialer; network, address := endpoint(socketPath); conn, err := d.DialContext(ctx, network, address); if err != nil { return Response{}, err }; defer conn.Close(); if err := json.NewEncoder(conn).Encode(Request{Action:"ping"}); err != nil { return Response{}, err }; var resp Response; if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil { return Response{}, err }; return resp, nil }
func endpoint(raw string) (string, string) { switch { case strings.HasPrefix(raw, "tcp:"): return "tcp", strings.TrimPrefix(raw, "tcp:"); case strings.HasPrefix(raw, "unix:"): return "unix", strings.TrimPrefix(raw, "unix:"); case strings.Contains(raw, "/"): return "unix", raw; default: return "tcp", raw } }
func FormatEndpoint(raw string) string { network, address := endpoint(raw); return fmt.Sprintf("%s:%s", network, address) }
func handle(conn net.Conn) { defer conn.Close(); var req Request; if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil { _ = json.NewEncoder(conn).Encode(Response{OK:false, Status:"decode-error"}); return }; switch req.Action { case "ping": _ = json.NewEncoder(conn).Encode(Response{OK:true, Status:"pong"}); default: _ = json.NewEncoder(conn).Encode(Response{OK:false, Status:"unknown-action"}) } }
