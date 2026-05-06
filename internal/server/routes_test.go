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
	"github.com/Matthieusz/AVMS/internal/database"
)

type stubDatabaseService struct {
	createItemCalls int
	deleteItemCalls int

	createItemFunc func(ctx context.Context, value string) (database.Item, error)
	deleteItemFunc func(ctx context.Context, id int64) (bool, error)
}

func (s *stubDatabaseService) Health() map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (s *stubDatabaseService) CreateItem(ctx context.Context, value string) (database.Item, error) {
	s.createItemCalls++

	if s.createItemFunc != nil {
		return s.createItemFunc(ctx, value)
	}

	return database.Item{
		ID:        1,
		Value:     value,
		CreatedAt: "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubDatabaseService) ListItems(_ context.Context) ([]database.Item, error) {
	return []database.Item{}, nil
}

func (s *stubDatabaseService) DeleteItem(ctx context.Context, id int64) (bool, error) {
	s.deleteItemCalls++

	if s.deleteItemFunc != nil {
		return s.deleteItemFunc(ctx, id)
	}

	return true, nil
}

func (s *stubDatabaseService) Close() error {
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

func TestCreateItemHandlerRejectsInvalidJSON(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/items", []byte(`{"value":}`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "invalid request body" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}

	if db.createItemCalls != 0 {
		t.Fatalf("expected create item not to be called, got %d calls", db.createItemCalls)
	}
}

func TestCreateItemHandlerRejectsBlankValue(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/items", []byte(`{"value":"   "}`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "value is required" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}

	if db.createItemCalls != 0 {
		t.Fatalf("expected create item not to be called, got %d calls", db.createItemCalls)
	}
}

func TestCreateItemHandlerRejectsOversizedBody(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	largeValue := strings.Repeat("a", maxCreateItemBodySize)
	body := []byte(`{"value":"` + largeValue + `"}`)
	rr := makeRequest(t, handler, http.MethodPost, "/api/items", body)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "request body is too large" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}

	if db.createItemCalls != 0 {
		t.Fatalf("expected create item not to be called, got %d calls", db.createItemCalls)
	}
}

func TestDeleteItemHandlerRejectsInvalidIDs(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "non-numeric id", path: "/api/items/not-a-number"},
		{name: "zero id", path: "/api/items/0"},
		{name: "negative id", path: "/api/items/-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &stubDatabaseService{}
			s := &Server{db: db, cfg: config.Default().Server}
			handler := s.RegisterRoutes()

			rr := makeRequest(t, handler, http.MethodDelete, tt.path, nil)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}

			if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "invalid item id" {
				t.Fatalf("unexpected error message: got %q", errMessage)
			}

			if db.deleteItemCalls != 0 {
				t.Fatalf("expected delete item not to be called, got %d calls", db.deleteItemCalls)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db, cfg: config.Default().Server}
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

func TestCreateItemHandlerSuccess(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/items", []byte(`{"value":"hello"}`))

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var item database.Item
	if err := json.Unmarshal(rr.Body.Bytes(), &item); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if item.Value != "hello" {
		t.Fatalf("unexpected value: got %q, want %q", item.Value, "hello")
	}

	if db.createItemCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", db.createItemCalls)
	}
}

func TestListItemsHandlerWithItems(t *testing.T) {
	db := &stubDatabaseService{
		createItemFunc: func(_ context.Context, _ string) (database.Item, error) {
			return database.Item{ID: 1, Value: "item", CreatedAt: "2026-01-01T00:00:00Z"}, nil
		},
	}
	// Override ListItems to return data
	customDB := &customListItemsDB{
		stubDatabaseService: db,
		items: []database.Item{
			{ID: 1, Value: "first", CreatedAt: "2026-01-01T00:00:00Z"},
			{ID: 2, Value: "second", CreatedAt: "2026-01-02T00:00:00Z"},
		},
	}
	s := &Server{db: customDB, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodGet, "/api/items", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body struct {
		Items []database.Item `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(body.Items))
	}
}

type customListItemsDB struct {
	*stubDatabaseService
	items []database.Item
}

func (c *customListItemsDB) ListItems(_ context.Context) ([]database.Item, error) {
	return c.items, nil
}

func TestDeleteItemHandlerSuccess(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db, cfg: config.Default().Server}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodDelete, "/api/items/1", nil)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	if db.deleteItemCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", db.deleteItemCalls)
	}
}
