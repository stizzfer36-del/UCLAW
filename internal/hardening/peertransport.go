package hardening

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/stizzfer36-del/uclaw/internal/config"
)

func servePeerTransport(ctx context.Context, cfg config.Config, listen string, log io.Writer) error {
	if strings.TrimSpace(listen) == "" {
		listen = "127.0.0.1:44144"
	}
	server := &http.Server{
		Addr:    listen,
		Handler: peerTransportHandler(cfg),
	}
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	go func() {
		errCh <- server.ListenAndServe()
	}()
	if log != nil {
		_, _ = fmt.Fprintf(log, "sync-transport listening on http://%s\n", listen)
	}
	err := <-errCh
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func pullPeerPackage(ctx context.Context, cfg config.Config, name, endpoint string) error {
	target, err := syncURL(endpoint, "/v1/sync/export")
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) != "" {
		q := target.Query()
		q.Set("peer", name)
		target.RawQuery = q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sync pull failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return importPeerPackageBytes(cfg, body)
}

func pushPeerPackage(ctx context.Context, cfg config.Config, name, endpoint string) error {
	path, err := exportPeerPackage(cfg, name)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	target, err := syncURL(endpoint, "/v1/sync/import")
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		reply, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sync push failed: %s: %s", resp.Status, strings.TrimSpace(string(reply)))
	}
	return nil
}

func peerTransportHandler(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/v1/sync/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		peer := strings.TrimSpace(r.URL.Query().Get("peer"))
		if peer == "" {
			peer = "peer"
		}
		path, err := exportPeerPackage(cfg, peer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		body, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	mux.HandleFunc("/v1/sync/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := importPeerPackageBytes(cfg, body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}` + "\n"))
	})
	return mux
}

func syncURL(raw, path string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("missing sync endpoint")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	target, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	target.Path = path
	return target, nil
}
