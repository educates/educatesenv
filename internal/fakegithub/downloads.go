package fakegithub

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
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
	var cut atomic.Bool
	s.Handle("GET /download/"+tag+"/"+name, func(w http.ResponseWriter, r *http.Request) {
		if cut.Swap(true) {
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
		rc := http.NewResponseController(w)
		_ = rc.Flush()
		conn, _, err := rc.Hijack()
		if err != nil {
			panic("fakegithub: cannot drop the connection: " + err.Error())
		}
		_ = conn.Close()
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
