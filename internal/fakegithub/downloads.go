package fakegithub

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Checksums returns a checksums.txt asset that lists the SHA-256 checksum
// of each of assets, in the format sha256sum writes and Educates releases
// publish.
func Checksums(assets ...Asset) Asset {
	var b strings.Builder
	for _, a := range assets {
		sum := sha256.Sum256(a.Content)
		b.WriteString(hex.EncodeToString(sum[:]) + "  " + a.Name + "\n")
	}
	return Asset{Name: "checksums.txt", Content: []byte(b.String())}
}

// CutNextDownload makes the next download of the asset name of the release
// tagged tag fail mid-transfer: the response announces the asset's full
// length and sends the first half of it, then the server drops the
// connection. Later downloads of the asset are served whole. Calling it
// twice for the same asset panics, as Handle does.
func (s *Server) CutNextDownload(tag, name string) {
	s.interruptNextDownload(tag, name, func(w http.ResponseWriter, _ *http.Request) {
		conn, _, err := http.NewResponseController(w).Hijack()
		if err != nil {
			panic("fakegithub: cannot drop the connection: " + err.Error())
		}
		_ = conn.Close()
	})
}

// stallLimit is how long a stalled download sends nothing before the
// server drops the connection, so that a client that never gives up fails
// instead of hanging the test.
const stallLimit = 10 * time.Second

// StallNextDownload makes the next download of the asset name of the
// release tagged tag stall mid-transfer: the response announces the
// asset's full length and sends the first half of it, then sends nothing
// more and keeps the connection open until the client gives up. After 10
// seconds the server drops the connection. Later downloads of the asset
// are served whole. Calling it twice for the same asset panics, as Handle
// does.
func (s *Server) StallNextDownload(tag, name string) {
	s.interruptNextDownload(tag, name, func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(stallLimit):
		}
	})
}

// TrickleDownloads makes every download of the asset name of the release
// tagged tag arrive slowly: the response announces the asset's full length
// and sends it in chunks pieces of about the same size, waiting interval
// before each one, so the transfer takes chunks times interval in all.
// Calling it twice for the same asset panics, as Handle does.
func (s *Server) TrickleDownloads(tag, name string, chunks int, interval time.Duration) {
	s.Handle("GET /download/"+tag+"/"+name, func(w http.ResponseWriter, r *http.Request) {
		content, ok := s.assetContent(tag, name)
		if !ok {
			notFound(w)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.WriteHeader(http.StatusOK)
		rc := http.NewResponseController(w)
		_ = rc.Flush()
		for i := range chunks {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(interval):
			}
			_, _ = w.Write(content[i*len(content)/chunks : (i+1)*len(content)/chunks])
			_ = rc.Flush()
		}
	})
}

// interruptNextDownload makes the next download of the asset name of the
// release tagged tag announce the asset's full length and send the first
// half of it, then calls interrupt to end the transfer. Later downloads of
// the asset are served whole.
func (s *Server) interruptNextDownload(tag, name string, interrupt http.HandlerFunc) {
	var interrupted atomic.Bool
	s.Handle("GET /download/"+tag+"/"+name, func(w http.ResponseWriter, r *http.Request) {
		if interrupted.Swap(true) {
			s.routes.ServeHTTP(w, r)
			return
		}

		content, ok := s.assetContent(tag, name)
		if !ok {
			notFound(w)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content[:len(content)/2])
		_ = http.NewResponseController(w).Flush()
		interrupt(w, r)
	})
}

// assetContent returns the content of the asset name of the release tagged
// tag.
func (s *Server) assetContent(tag, name string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stored := range s.releases {
		if stored.release.Tag != tag {
			continue
		}
		for _, a := range stored.assets {
			if a.asset.Name == name {
				return a.asset.Content, true
			}
		}
	}
	return nil, false
}
