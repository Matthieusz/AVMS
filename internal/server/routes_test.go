package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/Matthieusz/AVMS/internal/database"
	"github.com/Matthieusz/AVMS/internal/pqc"
)

type stubDatabaseService struct {
	createItemCalls int
	deleteItemCalls int

	registerVehicleFunc  func(ctx context.Context, input database.RegisterVehicleInput) (database.VehicleRecord, error)
	issueCredentialFunc  func(ctx context.Context, input database.IssueCredentialInput) (database.CredentialRecord, error)
	revokeCredentialFunc func(ctx context.Context, input database.RevokeCredentialInput) (database.CredentialRecord, error)
	rotateKeysFunc       func(ctx context.Context, input database.RotateKeyInput) (database.KeyRotationResult, error)
	credentialStatusFunc func(ctx context.Context, credentialID string) (database.CredentialRecord, error)
	currentPolicyFunc    func(ctx context.Context) (database.SecurityPolicy, error)
	joinVehicleFunc      func(ctx context.Context, input database.JoinVehicleInput) (database.JoinSessionRecord, error)
	reportIncidentFunc   func(ctx context.Context, input database.IncidentReportInput) (database.IncidentRecord, error)

	createItemFunc func(ctx context.Context, value string) (database.Item, error)
	deleteItemFunc func(ctx context.Context, id int64) (bool, error)
}

func (s *stubDatabaseService) Health() map[string]string {
	return map[string]string{
		"status":    "up",
		"message":   "healthy",
		"service":   "api",
		"timestamp": "2026-01-01T00:00:00Z",
	}
}

func (s *stubDatabaseService) RegisterVehicle(ctx context.Context, input database.RegisterVehicleInput) (database.VehicleRecord, error) {
	if s.registerVehicleFunc != nil {
		return s.registerVehicleFunc(ctx, input)
	}

	return database.VehicleRecord{
		VehicleID:          input.VehicleID,
		Manufacturer:       input.Manufacturer,
		HardwareProfile:    input.HardwareProfile,
		PublicKey:          input.PublicKey,
		SignatureAlgorithm: firstNonEmptyTest(input.SignatureAlgorithm, database.RecommendedSignatureAlgorithm),
		Status:             "registered",
		CreatedAt:          "2026-01-01T00:00:00Z",
		UpdatedAt:          "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubDatabaseService) IssueCredential(ctx context.Context, input database.IssueCredentialInput) (database.CredentialRecord, error) {
	if s.issueCredentialFunc != nil {
		return s.issueCredentialFunc(ctx, input)
	}

	return database.CredentialRecord{
		CredentialID: "cred_test",
		SubjectType:  input.SubjectType,
		SubjectID:    input.SubjectID,
		Algorithm:    input.Algorithm,
		Purpose:      input.Purpose,
		PublicKey:    input.PublicKey,
		Status:       "active",
		Version:      1,
		ValidFrom:    "2026-01-01T00:00:00Z",
		ValidTo:      "2026-04-01T00:00:00Z",
		IssuedAt:     "2026-01-01T00:00:00Z",
		LastUpdated:  "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubDatabaseService) RevokeCredential(ctx context.Context, input database.RevokeCredentialInput) (database.CredentialRecord, error) {
	if s.revokeCredentialFunc != nil {
		return s.revokeCredentialFunc(ctx, input)
	}

	return database.CredentialRecord{
		CredentialID: input.CredentialID,
		Status:       "revoked",
		RevokedAt:    "2026-01-01T00:00:00Z",
		RevokeReason: firstNonEmptyTest(input.Reason, "administrative decision"),
		LastUpdated:  "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubDatabaseService) RotateKeys(ctx context.Context, input database.RotateKeyInput) (database.KeyRotationResult, error) {
	if s.rotateKeysFunc != nil {
		return s.rotateKeysFunc(ctx, input)
	}

	return database.KeyRotationResult{
		VehicleID:            input.VehicleID,
		RotationReason:       firstNonEmptyTest(input.Reason, "scheduled rotation"),
		EffectiveAt:          "2026-01-01T00:00:00Z",
		RecommendedGraceDays: database.DefaultRotationWindowDays,
		NewCredential: database.CredentialRecord{
			CredentialID: "cred_rotated",
			SubjectType:  "vehicle",
			SubjectID:    input.VehicleID,
			Algorithm:    input.Algorithm,
			Purpose:      "signing",
			PublicKey:    input.PublicKey,
			Status:       "active",
			Version:      2,
			ValidFrom:    "2026-01-01T00:00:00Z",
			ValidTo:      "2026-04-01T00:00:00Z",
			IssuedAt:     "2026-01-01T00:00:00Z",
			LastUpdated:  "2026-01-01T00:00:00Z",
		},
	}, nil
}

func (s *stubDatabaseService) GetCredentialStatus(ctx context.Context, credentialID string) (database.CredentialRecord, error) {
	if s.credentialStatusFunc != nil {
		return s.credentialStatusFunc(ctx, credentialID)
	}

	return database.CredentialRecord{
		CredentialID: credentialID,
		Status:       "active",
		Algorithm:    database.RecommendedSignatureAlgorithm,
		Purpose:      "signing",
		ValidFrom:    "2026-01-01T00:00:00Z",
		ValidTo:      "2026-04-01T00:00:00Z",
		IssuedAt:     "2026-01-01T00:00:00Z",
		LastUpdated:  "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubDatabaseService) GetCurrentPolicy(ctx context.Context) (database.SecurityPolicy, error) {
	if s.currentPolicyFunc != nil {
		return s.currentPolicyFunc(ctx)
	}

	return database.CurrentSecurityPolicy(), nil
}

func (s *stubDatabaseService) JoinVehicle(ctx context.Context, input database.JoinVehicleInput) (database.JoinSessionRecord, error) {
	if s.joinVehicleFunc != nil {
		return s.joinVehicleFunc(ctx, input)
	}

	return database.JoinSessionRecord{
		SessionID:          "join_test",
		VehicleID:          input.VehicleID,
		RSUID:              input.RSUID,
		CredentialID:       firstNonEmptyTest(input.CredentialID, "cred_test"),
		KEMAlgorithm:       firstNonEmptyTest(input.KEMAlgorithm, database.RecommendedKEMAlgorithm),
		SignatureAlgorithm: firstNonEmptyTest(input.SignatureAlgorithm, database.RecommendedSignatureAlgorithm),
		SessionCipher:      database.DefaultSessionCipher,
		SessionKeyRef:      "sess_test",
		Status:             "accepted",
		VerificationNotes:  "Simulated onboarding accepted.",
		SimulationMode:     true,
		CreatedAt:          "2026-01-01T00:00:00Z",
		AcceptedAt:         "2026-01-01T00:00:00Z",
	}, nil
}

func (s *stubDatabaseService) ReportIncident(ctx context.Context, input database.IncidentReportInput) (database.IncidentRecord, error) {
	if s.reportIncidentFunc != nil {
		return s.reportIncidentFunc(ctx, input)
	}

	return database.IncidentRecord{
		IncidentID:        "inc_test",
		SubjectType:       input.SubjectType,
		SubjectID:         input.SubjectID,
		CredentialID:      input.CredentialID,
		Severity:          input.Severity,
		Description:       input.Description,
		RecommendedAction: "revoke the active credential, rotate keys, and require a fresh onboarding sequence",
		Status:            "open",
		ReportedAt:        "2026-01-01T00:00:00Z",
	}, nil
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

func firstNonEmptyTest(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{}
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
	s := &Server{}
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
		if !ok || kemName != database.RecommendedKEMAlgorithm {
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
	s := &Server{db: db}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/items", []byte(`{"value":`))

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
	s := &Server{db: db}
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
	s := &Server{db: db}
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
			s := &Server{db: db}
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

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty uses default",
			input: "",
			want:  []string{defaultAllowedOrigin},
		},
		{
			name:  "parses and deduplicates origins",
			input: " http://localhost:3000, http://localhost:5173, http://localhost:3000, *, ",
			want:  []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:  "only invalid entries falls back",
			input: " , *, ",
			want:  []string{defaultAllowedOrigin},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAllowedOrigins(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected origins: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db}
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
	s := &Server{db: db}
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
	s := &Server{db: customDB}
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
	s := &Server{db: db}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodDelete, "/api/items/1", nil)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	if db.deleteItemCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", db.deleteItemCalls)
	}
}

func TestRegisterVehicleHandlerSuccess(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/vehicles/register", []byte(`{"vehicleId":"veh-001","manufacturer":"ACME","hardwareProfile":"TPM-2.0","publicKey":"pk","signatureAlgorithm":"ML-DSA"}`))

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var body database.VehicleRecord
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.VehicleID != "veh-001" || body.Status != "registered" {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

func TestPolicyHandlerSuccess(t *testing.T) {
	db := &stubDatabaseService{}
	s := &Server{db: db}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodGet, "/api/policies/current", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body database.SecurityPolicy
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.RecommendedKEMAlgorithm != database.RecommendedKEMAlgorithm {
		t.Fatalf("unexpected policy response: %+v", body)
	}
}

func TestRSUBeaconHandlerSuccess(t *testing.T) {
	s := &Server{db: &stubDatabaseService{}}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodGet, "/api/rsus/rsu-01/beacon", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body pqc.RSUBeacon
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.RSUID != "rsu-01" || body.KEMAlgorithm != database.RecommendedKEMAlgorithm || body.KEMPublicKey == "" {
		t.Fatalf("unexpected beacon response: %+v", body)
	}
}

func TestJoinVehicleHandlerRejectsMissingFields(t *testing.T) {
	db := &stubDatabaseService{
		joinVehicleFunc: func(_ context.Context, _ database.JoinVehicleInput) (database.JoinSessionRecord, error) {
			return database.JoinSessionRecord{}, errors.New("vehicleId, rsuId, ciphertext, signature, and vehicleCertificate are required")
		},
	}
	s := &Server{db: db}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodPost, "/api/vehicles/join", []byte(`{"vehicleId":"veh-001"}`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if errMessage := decodeErrorMessage(t, rr.Body); errMessage != "vehicleId, rsuId, ciphertext, signature, and vehicleCertificate are required" {
		t.Fatalf("unexpected error message: got %q", errMessage)
	}
}

func TestCredentialStatusHandlerNotFound(t *testing.T) {
	db := &stubDatabaseService{
		credentialStatusFunc: func(_ context.Context, _ string) (database.CredentialRecord, error) {
			return database.CredentialRecord{}, database.ErrNotFound
		},
	}
	s := &Server{db: db}
	handler := s.RegisterRoutes()

	rr := makeRequest(t, handler, http.MethodGet, "/api/credentials/missing/status", nil)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}
