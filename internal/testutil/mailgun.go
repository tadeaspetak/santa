package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// MailgunServer fakes the parts of the Mailgun API the app uses: getting a domain
// (to verify the credentials) and sending messages.
type MailgunServer struct {
	// APIBase is what the Mailgun client should use instead of e.g. https://api.eu.mailgun.net/v3.
	APIBase string

	domain string
	apiKey string

	mu       sync.Mutex
	messages []Message
}

// NewMailgunServer starts a fake Mailgun API for the given domain and API key.
func NewMailgunServer(t *testing.T, domain, apiKey string) *MailgunServer {
	t.Helper()

	s := &MailgunServer{domain: domain, apiKey: apiKey}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v3/domains/{domain}", s.authenticated(func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("domain") != s.domain {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Domain not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"domain": map[string]string{"name": s.domain}})
	}))
	mux.HandleFunc("POST /v3/{domain}/messages", s.authenticated(func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("domain") != s.domain {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Domain not found"})
			return
		}
		s.mu.Lock()
		s.messages = append(s.messages, Message{
			From:    r.FormValue("from"),
			To:      r.FormValue("to"),
			ReplyTo: r.FormValue("h:Reply-To"),
			Subject: r.FormValue("subject"),
			HTML:    r.FormValue("html"),
		})
		s.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]string{"id": "<test@" + s.domain + ">", "message": "Queued. Thank you."})
	}))

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	s.APIBase = server.URL + "/v3"

	return s
}

// Messages returns a copy of all the messages received so far.
func (s *MailgunServer) Messages() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Message(nil), s.messages...)
}

func (s *MailgunServer) authenticated(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "api" || pass != s.apiKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Invalid private key"})
			return
		}
		handler(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
