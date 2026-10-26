// Package fakegithub serves the parts of the GitHub REST API that
// educatesenv uses, from releases a test sets up, on an httptest server.
//
// It serves one repository, educates/educates-training-platform unless
// SetRepository changes it:
//
//   - GET /repos/{owner}/{repo}/releases lists the releases in the order
//     they were added, paginated with page and per_page and a Link header,
//     as GitHub does.
//   - GET /repos/{owner}/{repo}/releases/tags/{tag} returns one release.
//   - GET /download/{tag}/{name} serves an asset's content. Each asset's
//     browser_download_url points there.
//
// Anything else, including another repository, gets GitHub's 404 body.
// Handle registers a handler that takes precedence over these, for
// responses a test needs that the defaults do not give.
package fakegithub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	// DefaultOwner is the owner of the repository the server serves by default.
	DefaultOwner = "educates"
	// DefaultRepository is the repository the server serves by default.
	DefaultRepository = "educates-training-platform"

	defaultPerPage = 30
	maxPerPage     = 100
)

// Release is a GitHub release the server serves.
type Release struct {
	Tag string
	// PublishedAt is served as published_at. The zero time is served as
	// null, as GitHub does for drafts.
	PublishedAt time.Time
	Prerelease  bool
	Draft       bool
	Assets      []Asset
}

// Asset is a file attached to a release.
type Asset struct {
	Name    string
	Content []byte
}

// Request is a request the server received.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
}

// Server is a fake GitHub API on an httptest server.
type Server struct {
	// URL is the server's base URL, with no trailing slash.
	URL string

	mu        sync.Mutex
	owner     string
	repo      string
	releases  []storedRelease
	nextID    int64
	requests  []Request
	overrides *http.ServeMux
	routes    *http.ServeMux
}

type storedRelease struct {
	id      int64
	release Release
	assets  []storedAsset
}

type storedAsset struct {
	id    int64
	asset Asset
}

// New starts a server that serves no releases yet. It is closed when the
// test ends.
func New(t testing.TB) *Server {
	t.Helper()

	s := &Server{
		owner:     DefaultOwner,
		repo:      DefaultRepository,
		overrides: http.NewServeMux(),
		routes:    http.NewServeMux(),
	}
	s.routes.HandleFunc("GET /repos/{owner}/{repo}/releases", s.listReleases)
	s.routes.HandleFunc("GET /repos/{owner}/{repo}/releases/tags/{tag}", s.releaseByTag)
	s.routes.HandleFunc("GET /download/{tag}/{name}", s.download)
	s.routes.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { notFound(w) })

	srv := httptest.NewServer(http.HandlerFunc(s.serveHTTP))
	t.Cleanup(srv.Close)
	s.URL = srv.URL
	return s
}

// SetRepository changes the repository the server serves.
func (s *Server) SetRepository(owner, repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner, s.repo = owner, repo
}

// AddReleases adds releases after the ones already added. Lists serve them
// in that order, so add them newest first, as GitHub lists them.
func (s *Server) AddReleases(releases ...Release) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range releases {
		s.nextID++
		stored := storedRelease{id: s.nextID, release: r}
		for _, a := range r.Assets {
			s.nextID++
			stored.assets = append(stored.assets, storedAsset{id: s.nextID, asset: a})
		}
		s.releases = append(s.releases, stored)
	}
}

// Handle registers handler for pattern, a net/http ServeMux pattern such
// as "GET /download/3.8.0/checksums.txt". A request that matches a pattern
// registered here gets that handler instead of the default response.
// Registering the same pattern twice panics, as ServeMux does.
func (s *Server) Handle(pattern string, handler http.HandlerFunc) {
	s.overrides.Handle(pattern, handler)
}

// DownloadURL returns the URL the server serves an asset at.
func (s *Server) DownloadURL(tag, name string) string {
	return s.URL + "/download/" + url.PathEscape(tag) + "/" + url.PathEscape(name)
}

// Requests returns the requests the server received, in order.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, Request{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: r.Header.Clone(),
	})
	s.mu.Unlock()

	if _, pattern := s.overrides.Handler(r); pattern != "" {
		s.overrides.ServeHTTP(w, r)
		return
	}
	s.routes.ServeHTTP(w, r)
}

func (s *Server) listReleases(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.servesRepository(r) {
		notFound(w)
		return
	}

	perPage := queryInt(r, "per_page", defaultPerPage)
	perPage = min(max(perPage, 1), maxPerPage)
	page := max(queryInt(r, "page", 1), 1)
	lastPage := max((len(s.releases)+perPage-1)/perPage, 1)

	start := min((page-1)*perPage, len(s.releases))
	end := min(start+perPage, len(s.releases))
	body := make([]releaseJSON, 0, end-start)
	for _, stored := range s.releases[start:end] {
		body = append(body, s.toJSON(stored))
	}

	if page < lastPage {
		pageURL := func(p int) string {
			q := url.Values{"per_page": {strconv.Itoa(perPage)}, "page": {strconv.Itoa(p)}}
			return fmt.Sprintf("%s%s?%s", s.URL, r.URL.Path, q.Encode())
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next", <%s>; rel="last"`, pageURL(page+1), pageURL(lastPage)))
	}
	writeJSON(w, http.StatusOK, body)
}

func (s *Server) releaseByTag(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.servesRepository(r) {
		notFound(w)
		return
	}

	tag := r.PathValue("tag")
	for _, stored := range s.releases {
		if stored.release.Tag == tag {
			writeJSON(w, http.StatusOK, s.toJSON(stored))
			return
		}
	}
	notFound(w)
}

func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag, name := r.PathValue("tag"), r.PathValue("name")
	for _, stored := range s.releases {
		if stored.release.Tag != tag {
			continue
		}
		for _, a := range stored.assets {
			if a.asset.Name == name {
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Length", strconv.Itoa(len(a.asset.Content)))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(a.asset.Content)
				return
			}
		}
	}
	notFound(w)
}

func (s *Server) servesRepository(r *http.Request) bool {
	return r.PathValue("owner") == s.owner && r.PathValue("repo") == s.repo
}

type releaseJSON struct {
	ID          int64       `json:"id"`
	TagName     string      `json:"tag_name"`
	Name        string      `json:"name"`
	Draft       bool        `json:"draft"`
	Prerelease  bool        `json:"prerelease"`
	PublishedAt *time.Time  `json:"published_at"`
	Assets      []assetJSON `json:"assets"`
}

type assetJSON struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Size               int    `json:"size"`
	ContentType        string `json:"content_type"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (s *Server) toJSON(stored storedRelease) releaseJSON {
	rel := stored.release
	out := releaseJSON{
		ID:         stored.id,
		TagName:    rel.Tag,
		Name:       rel.Tag,
		Draft:      rel.Draft,
		Prerelease: rel.Prerelease,
		Assets:     make([]assetJSON, 0, len(stored.assets)),
	}
	if !rel.PublishedAt.IsZero() {
		published := rel.PublishedAt.UTC()
		out.PublishedAt = &published
	}
	for _, a := range stored.assets {
		out.Assets = append(out.Assets, assetJSON{
			ID:                 a.id,
			Name:               a.asset.Name,
			Size:               len(a.asset.Content),
			ContentType:        "application/octet-stream",
			BrowserDownloadURL: s.DownloadURL(rel.Tag, a.asset.Name),
		})
	}
	return out
}

func queryInt(r *http.Request, key string, fallback int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return fallback
	}
	return n
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{
		"message":           "Not Found",
		"documentation_url": "https://docs.github.com/rest",
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
