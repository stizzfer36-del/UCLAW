package ipc

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
)

func TestSocketPing(t *testing.T) {
	serverConn, clientConn := net.Pipe(); defer serverConn.Close(); defer clientConn.Close(); go func() { handle(serverConn) }()
	if err := json.NewEncoder(clientConn).Encode(Request{Action:"ping"}); err != nil { t.Fatal(err) }
	var resp Response; if err := json.NewDecoder(bufio.NewReader(clientConn)).Decode(&resp); err != nil { t.Fatal(err) }
	if !resp.OK || resp.Status != "pong" { t.Fatalf("unexpected response: %+v", resp) }
}
