package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/Matthieusz/AVMS/internal/config"
	"github.com/Matthieusz/AVMS/internal/entry"
)

type stubEntryService struct {
	createEntryCalls int
	deleteEntryCalls int

	createEntryFunc func(ctx context.Context, value string) (entry.Entry, error)
	listEntriesFunc func(ctx context.Context) ([]entry.Entry, error)
	deleteEntryFunc func(ctx context.Context, id int64) (bool, error)
}

func (s *stubEntryService) Health() map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (s *stubEntryService) CreateEntry(ctx context.Context, value string) (entry.Entry, error) {
	s.createEntryCalls++

	if s.createEntryFunc != nil {
		return s.createEntryFunc(ctx, value)
	}

	if strings.TrimSpace(value) == "" {
		return entry.Entry{}, entry.ValidationError{Field: "value", Message: entry.ErrBlankValue.Error()}
	}

	return entry.Entry{
		ID:        1,
		Value:     value,
		CreatedAt: "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubEntryService) ListEntries(_ context.Context) ([]entry.Entry, error) {
	if s.listEntriesFunc != nil {
		return s.listEntriesFunc(context.Background())
	}
	return []entry.Entry{}, nil
}

func (s *stubEntryService) DeleteEntry(ctx context.Context, id int64) (bool, error) {
	s.deleteEntryCalls++

	if s.deleteEntryFunc != nil {
		return s.deleteEntryFunc(ctx, id)
	}

	return true, nil
}

func (s *stubEntryService) Close() error {
	return nil
}

func makeRequest(t *testing.T, handler http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func decodeErrorMessage(t *testing.T, body *bytes.Buffer) string {
	t.Helper()

	var payload struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}

	return payload.Error
}

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{cfg: config.Default().Server}
	r := gin.New()
	r.GET("/api/", s.HelloWorldHandler)
	rr := makeRequest(t, r, http.MethodGet, "/api/", nil)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body
	expected := "{\"message\":\"Hello World\"}"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestKEMCheckHandler(t *testing.T) {
	s := &Server{cfg: config.Default().Server}
	r := gin.New()
	r.GET("/api/pqc/kem-check", s.kemCheckHandler)

	rr := makeRequest(t, r, http.MethodGet, "/api/pqc/kem-check", nil)

	if rr.Code != http.StatusOK {
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("KEM check failed with status %d and body %s", rr.Code, rr.Body.String())
		}

		if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "failed to run KEM check" {
			t.Fatalf("unexpected error message: got %q", errMessage)
		}

		return
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if _, exists := body["clientSharedSecretPreview"]; exists {
		t.Fatal("response must not expose client shared secret preview")
	}

	if _, exists := body["serverSharedSecretPreview"]; exists {
		t.Fatal("response must not expose server shared secret preview")
	}

	liboqsVersion, ok := body["liboqsVersion"].(string)
	if !ok || liboqsVersion == "" {
		t.Fatal("expected liboqsVersion to be present")
	}

	enabledKEMs, ok := body["enabledKEMs"].([]any)
	if !ok || len(enabledKEMs) == 0 {
		t.Fatal("expected at least one enabled KEM")
	}

	kemName, ok := body["kemName"].(string)
	if !ok || kemName != "ML-KEM-512" {
		t.Fatalf("unexpected KEM name: got %#v", body["kemName"])
	}

	sharedSecretsCoincide, ok := body["sharedSecretsCoincide"].(bool)
	if !ok {
		t.Fatalf("sharedSecretsCoincide should be a bool, got %#v", body["sharedSecretsCoincide"])
	}

	if !sharedSecretsCoincide {
		t.Fatal("expected shared secrets to coincide")
	}
}

func TestCreateEntryHandlerRejectsInvalidJSON(t *testing.T) {
	srv := &stubEntryService{}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/entries", []byte(`{"value":}`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "invalid request body" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}

	if srv.createEntryCalls != 0 {
		t.Fatalf("expected create entry not to be called, got %d calls", srv.createEntryCalls)
	}
}

func TestCreateEntryHandlerRejectsBlankValue(t *testing.T) {
	srv := &stubEntryService{}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/entries", []byte(`{"value":"   "}`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "value is required" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}

	if srv.createEntryCalls != 1 {
		t.Fatalf("expected create entry to be called once, got %d calls", srv.createEntryCalls)
	}
}

func TestCreateEntryHandlerRejectsOversizedBody(t *testing.T) {
	srv := &stubEntryService{}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	largeValue := strings.Repeat("a", maxCreateEntryBodySize)
	body := []byte(`{"value":"` + largeValue + `"}`)
	rr := makeRequest(t, handler, http.MethodPost, "/api/entries", body)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "request body is too large" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}

	if srv.createEntryCalls != 0 {
		t.Fatalf("expected create entry not to be called, got %d calls", srv.createEntryCalls)
	}
}

func TestDeleteEntryHandlerRejectsInvalidIDs(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "non-numeric id", path: "/api/entries/not-a-number"},
		{name: "zero id", path: "/api/entries/0"},
		{name: "negative id", path: "/api/entries/-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &stubEntryService{}
			s := &Server{entries: srv, cfg: config.Default().Server}
			handler := s.RegisterRoutes()

			rr := makeRequest(t, handler, http.MethodDelete, tt.path, nil)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}

			if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "invalid entry id" {
				t.Fatalf("unexpected error message: got %q", errMessage)
			}

			if srv.deleteEntryCalls != 0 {
				t.Fatalf("expected delete entry not to be called, got %d calls", srv.deleteEntryCalls)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	srv := &stubEntryService{}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodGet, "/api/health", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "up" {
		t.Fatalf("unexpected status: got %v", body["status"])
	}
}

func TestCreateEntryHandlerSuccess(t *testing.T) {
	srv := &stubEntryService{}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/entries", []byte(`{"value":"hello"}`))

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var ent entry.Entry
	if err := json.Unmarshal(rr.Body.Bytes(), &ent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if ent.Value != "hello" {
		t.Fatalf("unexpected value: got %q, want %q", ent.Value, "hello")
	}

	if srv.createEntryCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", srv.createEntryCalls)
	}
}

func TestListEntriesHandlerWithItems(t *testing.T) {
	srv := &stubEntryService{
		listEntriesFunc: func(_ context.Context) ([]entry.Entry, error) {
			return []entry.Entry{
				{ID: 1, Value: "first", CreatedAt: "2026-01-01T00:00:00Z"},
				{ID: 2, Value: "second", CreatedAt: "2026-01-02T00:00:00Z"},
			}, nil
		},
	}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodGet, "/api/entries", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body struct {
		Entries []entry.Entry `json:"entries"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(body.Entries))
	}
}

func TestDeleteEntryHandlerSuccess(t *testing.T) {
	srv := &stubEntryService{}
	s := &Server{entries: srv, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodDelete, "/api/entries/1", nil)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	if srv.deleteEntryCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", srv.deleteEntryCalls)
	}
}
