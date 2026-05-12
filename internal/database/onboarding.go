package database

import (
	"context"
	"crypto/sha256"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Matthieusz/AVMS/internal/pqc"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

const (
	RecommendedKEMAlgorithm       = "ML-KEM-768"
	RecommendedSignatureAlgorithm = "ML-DSA"
	DefaultSessionCipher          = "AES-256-GCM"
	DefaultCredentialValidityDays = 90
	DefaultRotationWindowDays     = 7
)

var allowedKEMAlgorithms = map[string]struct{}{
	"ML-KEM-512":  {},
	"ML-KEM-768":  {},
	"ML-KEM-1024": {},
}

var allowedSignatureAlgorithms = map[string]struct{}{
	"ML-DSA":  {},
	"SLH-DSA": {},
}

type RegisterVehicleInput struct {
	VehicleID          string
	Manufacturer       string
	HardwareProfile    string
	PublicKey          string
	SignatureAlgorithm string
}

type VehicleRecord struct {
	VehicleID           string `json:"vehicleId"`
	Manufacturer        string `json:"manufacturer"`
	HardwareProfile     string `json:"hardwareProfile"`
	PublicKey           string `json:"publicKey"`
	SignatureAlgorithm  string `json:"signatureAlgorithm"`
	Status              string `json:"status"`
	CurrentCredentialID string `json:"currentCredentialId,omitempty"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type IssueCredentialInput struct {
	SubjectType  string
	SubjectID    string
	Algorithm    string
	Purpose      string
	PublicKey    string
	ValidityDays int
}

type CredentialRecord struct {
	CredentialID string `json:"credentialId"`
	SubjectType  string `json:"subjectType"`
	SubjectID    string `json:"subjectId"`
	Algorithm    string `json:"algorithm"`
	Purpose      string `json:"purpose"`
	PublicKey    string `json:"publicKey"`
	Status       string `json:"status"`
	Version      int    `json:"version"`
	ValidFrom    string `json:"validFrom"`
	ValidTo      string `json:"validTo"`
	IssuedAt     string `json:"issuedAt"`
	RevokedAt    string `json:"revokedAt,omitempty"`
	RevokeReason string `json:"revokeReason,omitempty"`
	LastUpdated  string `json:"lastUpdated"`
}

type RevokeCredentialInput struct {
	CredentialID string
	Reason       string
}

type RotateKeyInput struct {
	VehicleID    string
	Algorithm    string
	PublicKey    string
	Reason       string
	ValidityDays int
}

type KeyRotationResult struct {
	VehicleID             string           `json:"vehicleId"`
	PreviousCredentialID  string           `json:"previousCredentialId,omitempty"`
	RotationReason        string           `json:"rotationReason"`
	EffectiveAt           string           `json:"effectiveAt"`
	RecommendedGraceDays  int              `json:"recommendedGraceDays"`
	NewCredential         CredentialRecord `json:"newCredential"`
}

type JoinVehicleInput struct {
	VehicleID          string
	RSUID              string
	CredentialID       string
	KEMAlgorithm       string
	SignatureAlgorithm string
	Ciphertext         string
	Signature          string
	VehicleCertificate string
}

type JoinSessionRecord struct {
	SessionID           string `json:"sessionId"`
	VehicleID           string `json:"vehicleId"`
	RSUID               string `json:"rsuId"`
	CredentialID        string `json:"credentialId"`
	KEMAlgorithm        string `json:"kemAlgorithm"`
	SignatureAlgorithm  string `json:"signatureAlgorithm"`
	SessionCipher       string `json:"sessionCipher"`
	SessionKeyRef       string `json:"sessionKeyRef"`
	Status              string `json:"status"`
	VerificationNotes   string `json:"verificationNotes"`
	SimulationMode      bool   `json:"simulationMode"`
	CreatedAt           string `json:"createdAt"`
	AcceptedAt          string `json:"acceptedAt,omitempty"`
	VehicleCertificate  string `json:"vehicleCertificate,omitempty"`
}

type IncidentReportInput struct {
	SubjectType   string
	SubjectID     string
	CredentialID  string
	Severity      string
	Description   string
}

type IncidentRecord struct {
	IncidentID         string `json:"incidentId"`
	SubjectType        string `json:"subjectType"`
	SubjectID          string `json:"subjectId"`
	CredentialID       string `json:"credentialId,omitempty"`
	Severity           string `json:"severity"`
	Description        string `json:"description"`
	RecommendedAction  string `json:"recommendedAction"`
	Status             string `json:"status"`
	ReportedAt         string `json:"reportedAt"`
}

type SecurityPolicy struct {
	RecommendedKEMAlgorithm       string   `json:"recommendedKemAlgorithm"`
	RecommendedSignatureAlgorithm string   `json:"recommendedSignatureAlgorithm"`
	AllowedKEMAlgorithms          []string `json:"allowedKemAlgorithms"`
	AllowedSignatureAlgorithms    []string `json:"allowedSignatureAlgorithms"`
	SessionCipher                 string   `json:"sessionCipher"`
	CredentialValidityDays        int      `json:"credentialValidityDays"`
	RotationWindowDays            int      `json:"rotationWindowDays"`
	HybridModeRecommended         bool     `json:"hybridModeRecommended"`
	PrivateKeyStorage             string   `json:"privateKeyStorage"`
	Notes                         []string `json:"notes"`
}

func CurrentSecurityPolicy() SecurityPolicy {
	return SecurityPolicy{
		RecommendedKEMAlgorithm:       RecommendedKEMAlgorithm,
		RecommendedSignatureAlgorithm: RecommendedSignatureAlgorithm,
		AllowedKEMAlgorithms:          []string{"ML-KEM-512", "ML-KEM-768", "ML-KEM-1024"},
		AllowedSignatureAlgorithms:    []string{"ML-DSA", "SLH-DSA"},
		SessionCipher:                 DefaultSessionCipher,
		CredentialValidityDays:        DefaultCredentialValidityDays,
		RotationWindowDays:            DefaultRotationWindowDays,
		HybridModeRecommended:         true,
		PrivateKeyStorage:             "Private keys must remain inside the vehicle secure element, TPM, or HSM.",
		Notes: []string{
			"ML-KEM-768 is the default onboarding KEM because it balances transport latency with security margin.",
			"ML-DSA is the primary signature algorithm for vehicle and infrastructure credentials.",
			"SLH-DSA remains available for supplemental long-term or infrastructure-focused signature flows.",
			"Post-quantum mechanisms should be concentrated on onboarding, session setup, OTA, and administrative channels.",
		},
	}
}

func (s *service) RegisterVehicle(ctx context.Context, input RegisterVehicleInput) (VehicleRecord, error) {
	vehicleID := strings.TrimSpace(input.VehicleID)
	manufacturer := strings.TrimSpace(input.Manufacturer)
	hardwareProfile := strings.TrimSpace(input.HardwareProfile)
	publicKey := strings.TrimSpace(input.PublicKey)
	signatureAlgorithm := strings.TrimSpace(input.SignatureAlgorithm)

	if vehicleID == "" || manufacturer == "" || hardwareProfile == "" || publicKey == "" {
		return VehicleRecord{}, fmt.Errorf("vehicleId, manufacturer, hardwareProfile, and publicKey are required")
	}

	if signatureAlgorithm == "" {
		signatureAlgorithm = RecommendedSignatureAlgorithm
	}
	if err := ensureAllowedSignature(signatureAlgorithm); err != nil {
		return VehicleRecord{}, err
	}
	if err := pqc.ValidateSignaturePublicKey(signatureAlgorithm, publicKey); err != nil {
		return VehicleRecord{}, fmt.Errorf("publicKey does not match %s: %w", signatureAlgorithm, err)
	}

	now := nowUTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO vehicles (vehicle_id, manufacturer, hardware_profile, public_key, signature_algorithm, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'registered', ?, ?)
		ON CONFLICT(vehicle_id) DO UPDATE SET
			manufacturer = excluded.manufacturer,
			hardware_profile = excluded.hardware_profile,
			public_key = excluded.public_key,
			signature_algorithm = excluded.signature_algorithm,
			status = 'registered',
			updated_at = excluded.updated_at
	`, vehicleID, manufacturer, hardwareProfile, publicKey, signatureAlgorithm, now, now)
	if err != nil {
		return VehicleRecord{}, fmt.Errorf("register vehicle: %w", err)
	}

	return s.lookupVehicle(ctx, vehicleID)
}

func (s *service) IssueCredential(ctx context.Context, input IssueCredentialInput) (CredentialRecord, error) {
	subjectType := normalizeSubjectType(input.SubjectType)
	subjectID := strings.TrimSpace(input.SubjectID)
	algorithm := strings.TrimSpace(input.Algorithm)
	purpose := normalizePurpose(input.Purpose)
	publicKey := strings.TrimSpace(input.PublicKey)
	validityDays := input.ValidityDays

	if subjectType == "" || subjectID == "" || algorithm == "" || purpose == "" || publicKey == "" {
		return CredentialRecord{}, fmt.Errorf("subjectType, subjectId, algorithm, purpose, and publicKey are required")
	}

	if validityDays <= 0 {
		validityDays = DefaultCredentialValidityDays
	}

	if err := validateAlgorithmForPurpose(algorithm, purpose); err != nil {
		return CredentialRecord{}, err
	}
	if purpose == "kem" {
		if err := pqc.ValidateKEMPublicKey(algorithm, publicKey); err != nil {
			return CredentialRecord{}, fmt.Errorf("publicKey does not match %s: %w", algorithm, err)
		}
	} else {
		if err := pqc.ValidateSignaturePublicKey(algorithm, publicKey); err != nil {
			return CredentialRecord{}, fmt.Errorf("publicKey does not match %s: %w", algorithm, err)
		}
	}

	if subjectType == "vehicle" {
		if _, err := s.lookupVehicle(ctx, subjectID); err != nil {
			return CredentialRecord{}, err
		}
	}

	version, err := s.nextCredentialVersion(ctx, subjectType, subjectID, purpose)
	if err != nil {
		return CredentialRecord{}, err
	}

	credentialID, err := randomID("cred")
	if err != nil {
		return CredentialRecord{}, fmt.Errorf("generate credential id: %w", err)
	}

	now := nowUTC()
	validTo := time.Now().UTC().Add(time.Duration(validityDays) * 24 * time.Hour).Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO credentials (
			credential_id, subject_type, subject_id, algorithm, purpose, public_key, status, version, valid_from, valid_to, issued_at
		)
		VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, ?, ?)
	`, credentialID, subjectType, subjectID, algorithm, purpose, publicKey, version, now, validTo, now)
	if err != nil {
		return CredentialRecord{}, fmt.Errorf("issue credential: %w", err)
	}

	if subjectType == "vehicle" {
		if _, err := s.db.ExecContext(ctx, `
			UPDATE vehicles
			SET current_credential_id = ?, updated_at = ?
			WHERE vehicle_id = ?
		`, credentialID, now, subjectID); err != nil {
			return CredentialRecord{}, fmt.Errorf("link credential to vehicle: %w", err)
		}
	}

	return s.GetCredentialStatus(ctx, credentialID)
}

func (s *service) RevokeCredential(ctx context.Context, input RevokeCredentialInput) (CredentialRecord, error) {
	credentialID := strings.TrimSpace(input.CredentialID)
	reason := strings.TrimSpace(input.Reason)
	if credentialID == "" {
		return CredentialRecord{}, fmt.Errorf("credentialId is required")
	}
	if reason == "" {
		reason = "administrative decision"
	}

	now := nowUTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE credentials
		SET status = 'revoked', revoked_at = ?, revoke_reason = ?
		WHERE credential_id = ? AND status != 'revoked'
	`, now, reason, credentialID)
	if err != nil {
		return CredentialRecord{}, fmt.Errorf("revoke credential: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return CredentialRecord{}, fmt.Errorf("revoke credential rows affected: %w", err)
	}
	if rowsAffected == 0 {
		status, statusErr := s.GetCredentialStatus(ctx, credentialID)
		if statusErr != nil {
			return CredentialRecord{}, statusErr
		}
		return status, nil
	}

	if _, err := s.db.ExecContext(ctx, `
		UPDATE vehicles SET current_credential_id = NULL, updated_at = ? WHERE current_credential_id = ?
	`, now, credentialID); err != nil {
		return CredentialRecord{}, fmt.Errorf("unlink revoked credential from vehicle: %w", err)
	}

	return s.GetCredentialStatus(ctx, credentialID)
}

func (s *service) RotateKeys(ctx context.Context, input RotateKeyInput) (KeyRotationResult, error) {
	vehicleID := strings.TrimSpace(input.VehicleID)
	algorithm := strings.TrimSpace(input.Algorithm)
	publicKey := strings.TrimSpace(input.PublicKey)
	reason := strings.TrimSpace(input.Reason)
	if vehicleID == "" || algorithm == "" || publicKey == "" {
		return KeyRotationResult{}, fmt.Errorf("vehicleId, algorithm, and publicKey are required")
	}
	if reason == "" {
		reason = "scheduled rotation"
	}

	vehicle, err := s.lookupVehicle(ctx, vehicleID)
	if err != nil {
		return KeyRotationResult{}, err
	}

	if err := ensureAllowedSignature(algorithm); err != nil {
		return KeyRotationResult{}, err
	}
	if err := pqc.ValidateSignaturePublicKey(algorithm, publicKey); err != nil {
		return KeyRotationResult{}, fmt.Errorf("publicKey does not match %s: %w", algorithm, err)
	}

	now := nowUTC()
	if vehicle.CurrentCredentialID != "" {
		if _, err := s.db.ExecContext(ctx, `
			UPDATE credentials SET status = 'superseded' WHERE credential_id = ? AND status = 'active'
		`, vehicle.CurrentCredentialID); err != nil {
			return KeyRotationResult{}, fmt.Errorf("supersede previous credential: %w", err)
		}
	}

	if _, err := s.db.ExecContext(ctx, `
		UPDATE vehicles
		SET public_key = ?, signature_algorithm = ?, updated_at = ?
		WHERE vehicle_id = ?
	`, publicKey, algorithm, now, vehicleID); err != nil {
		return KeyRotationResult{}, fmt.Errorf("update vehicle after rotation: %w", err)
	}

	credential, err := s.IssueCredential(ctx, IssueCredentialInput{
		SubjectType:  "vehicle",
		SubjectID:    vehicleID,
		Algorithm:    algorithm,
		Purpose:      "signing",
		PublicKey:    publicKey,
		ValidityDays: input.ValidityDays,
	})
	if err != nil {
		return KeyRotationResult{}, err
	}

	return KeyRotationResult{
		VehicleID:            vehicleID,
		PreviousCredentialID: vehicle.CurrentCredentialID,
		RotationReason:       reason,
		EffectiveAt:          now,
		RecommendedGraceDays: DefaultRotationWindowDays,
		NewCredential:        credential,
	}, nil
}

func (s *service) GetCredentialStatus(ctx context.Context, credentialID string) (CredentialRecord, error) {
	credentialID = strings.TrimSpace(credentialID)
	if credentialID == "" {
		return CredentialRecord{}, fmt.Errorf("credentialId is required")
	}

	var record CredentialRecord
	var revokedAt sql.NullString
	var revokeReason sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT credential_id, subject_type, subject_id, algorithm, purpose, public_key, status, version, valid_from, valid_to, issued_at, revoked_at, revoke_reason
		FROM credentials
		WHERE credential_id = ?
	`, credentialID).Scan(
		&record.CredentialID,
		&record.SubjectType,
		&record.SubjectID,
		&record.Algorithm,
		&record.Purpose,
		&record.PublicKey,
		&record.Status,
		&record.Version,
		&record.ValidFrom,
		&record.ValidTo,
		&record.IssuedAt,
		&revokedAt,
		&revokeReason,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CredentialRecord{}, fmt.Errorf("%w: credential %q", ErrNotFound, credentialID)
		}
		return CredentialRecord{}, fmt.Errorf("lookup credential: %w", err)
	}

	if revokedAt.Valid {
		record.RevokedAt = revokedAt.String
	}
	if revokeReason.Valid {
		record.RevokeReason = revokeReason.String
	}
	record.Status = deriveCredentialStatus(record.Status, record.ValidTo, record.RevokedAt)
	record.LastUpdated = firstNonEmpty(record.RevokedAt, record.IssuedAt)

	return record, nil
}

func (s *service) GetCurrentPolicy(context.Context) (SecurityPolicy, error) {
	return CurrentSecurityPolicy(), nil
}

func (s *service) JoinVehicle(ctx context.Context, input JoinVehicleInput) (JoinSessionRecord, error) {
	vehicleID := strings.TrimSpace(input.VehicleID)
	rsuID := strings.TrimSpace(input.RSUID)
	credentialID := strings.TrimSpace(input.CredentialID)
	kemAlgorithm := strings.TrimSpace(input.KEMAlgorithm)
	signatureAlgorithm := strings.TrimSpace(input.SignatureAlgorithm)
	ciphertext := strings.TrimSpace(input.Ciphertext)
	signature := strings.TrimSpace(input.Signature)
	vehicleCertificate := strings.TrimSpace(input.VehicleCertificate)

	if vehicleID == "" || rsuID == "" || ciphertext == "" || signature == "" || vehicleCertificate == "" {
		return JoinSessionRecord{}, fmt.Errorf("vehicleId, rsuId, ciphertext, signature, and vehicleCertificate are required")
	}

	if kemAlgorithm == "" {
		kemAlgorithm = RecommendedKEMAlgorithm
	}
	if err := ensureAllowedKEM(kemAlgorithm); err != nil {
		return JoinSessionRecord{}, err
	}

	vehicle, err := s.lookupVehicle(ctx, vehicleID)
	if err != nil {
		return JoinSessionRecord{}, err
	}

	if signatureAlgorithm == "" {
		signatureAlgorithm = vehicle.SignatureAlgorithm
	}
	if err := ensureAllowedSignature(signatureAlgorithm); err != nil {
		return JoinSessionRecord{}, err
	}

	if credentialID == "" {
		credentialID = vehicle.CurrentCredentialID
	}
	if credentialID == "" {
		return JoinSessionRecord{}, fmt.Errorf("%w: vehicle %q has no active credential", ErrConflict, vehicleID)
	}

	credential, err := s.GetCredentialStatus(ctx, credentialID)
	if err != nil {
		return JoinSessionRecord{}, err
	}
	if credential.Status != "active" {
		return JoinSessionRecord{}, fmt.Errorf("%w: credential %q is %s", ErrConflict, credentialID, credential.Status)
	}
	if credential.Purpose != "signing" {
		return JoinSessionRecord{}, fmt.Errorf("%w: credential %q is not a signing credential", ErrConflict, credentialID)
	}
	if credential.Algorithm != signatureAlgorithm {
		return JoinSessionRecord{}, fmt.Errorf("%w: signature algorithm %q does not match credential algorithm %q", ErrConflict, signatureAlgorithm, credential.Algorithm)
	}
	if vehicle.SignatureAlgorithm != signatureAlgorithm {
		return JoinSessionRecord{}, fmt.Errorf("%w: vehicle %q is registered for %s, not %s", ErrConflict, vehicleID, vehicle.SignatureAlgorithm, signatureAlgorithm)
	}

	joinPayload := buildJoinPayload(vehicleID, rsuID, credentialID, kemAlgorithm, vehicleCertificate, ciphertext)
	if err := pqc.VerifySignature(signatureAlgorithm, credential.PublicKey, joinPayload, signature); err != nil {
		return JoinSessionRecord{}, fmt.Errorf("%w: vehicle signature verification failed: %v", ErrConflict, err)
	}

	sharedSecret, err := pqc.DecapsulateRSU(rsuID, kemAlgorithm, ciphertext)
	if err != nil {
		return JoinSessionRecord{}, fmt.Errorf("%w: KEM decapsulation failed: %v", ErrConflict, err)
	}

	sessionID, err := randomID("join")
	if err != nil {
		return JoinSessionRecord{}, fmt.Errorf("generate join session id: %w", err)
	}
	sessionKeyRef := pqc.SessionKeyReference(sharedSecret)

	now := nowUTC()
	notes := fmt.Sprintf("Verified %s signature for vehicle %s and decapsulated %s ciphertext for RSU %s.", signatureAlgorithm, vehicleID, kemAlgorithm, rsuID)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO join_sessions (
			session_id, vehicle_id, rsu_id, credential_id, kem_algorithm, signature_algorithm, ciphertext, signature, session_key_ref, status, verification_notes, created_at, accepted_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'accepted', ?, ?, ?)
	`, sessionID, vehicleID, rsuID, credentialID, kemAlgorithm, signatureAlgorithm, ciphertext, signature, sessionKeyRef, notes, now, now)
	if err != nil {
		return JoinSessionRecord{}, fmt.Errorf("record join session: %w", err)
	}

	return JoinSessionRecord{
		SessionID:          sessionID,
		VehicleID:          vehicleID,
		RSUID:              rsuID,
		CredentialID:       credentialID,
		KEMAlgorithm:       kemAlgorithm,
		SignatureAlgorithm: signatureAlgorithm,
		SessionCipher:      DefaultSessionCipher,
		SessionKeyRef:      sessionKeyRef,
		Status:             "accepted",
		VerificationNotes:  notes,
		SimulationMode:     false,
		CreatedAt:          now,
		AcceptedAt:         now,
		VehicleCertificate: vehicleCertificate,
	}, nil
}

func (s *service) ReportIncident(ctx context.Context, input IncidentReportInput) (IncidentRecord, error) {
	subjectType := normalizeSubjectType(input.SubjectType)
	subjectID := strings.TrimSpace(input.SubjectID)
	credentialID := strings.TrimSpace(input.CredentialID)
	severity := normalizeSeverity(input.Severity)
	description := strings.TrimSpace(input.Description)

	if subjectType == "" || subjectID == "" || severity == "" || description == "" {
		return IncidentRecord{}, fmt.Errorf("subjectType, subjectId, severity, and description are required")
	}

	incidentID, err := randomID("inc")
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("generate incident id: %w", err)
	}

	recommendedAction := "review telemetry and keep the credential under observation"
	if severity == "high" || severity == "critical" {
		recommendedAction = "revoke the active credential, rotate keys, and require a fresh onboarding sequence"
	}

	now := nowUTC()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO incidents (
			incident_id, subject_type, subject_id, credential_id, severity, description, recommended_action, status, reported_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'open', ?)
	`, incidentID, subjectType, subjectID, nullableString(credentialID), severity, description, recommendedAction, now)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("report incident: %w", err)
	}

	return IncidentRecord{
		IncidentID:        incidentID,
		SubjectType:       subjectType,
		SubjectID:         subjectID,
		CredentialID:      credentialID,
		Severity:          severity,
		Description:       description,
		RecommendedAction: recommendedAction,
		Status:            "open",
		ReportedAt:        now,
	}, nil
}

func (s *service) lookupVehicle(ctx context.Context, vehicleID string) (VehicleRecord, error) {
	var record VehicleRecord
	var currentCredentialID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT vehicle_id, manufacturer, hardware_profile, public_key, signature_algorithm, status, current_credential_id, created_at, updated_at
		FROM vehicles
		WHERE vehicle_id = ?
	`, vehicleID).Scan(
		&record.VehicleID,
		&record.Manufacturer,
		&record.HardwareProfile,
		&record.PublicKey,
		&record.SignatureAlgorithm,
		&record.Status,
		&currentCredentialID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return VehicleRecord{}, fmt.Errorf("%w: vehicle %q", ErrNotFound, vehicleID)
		}
		return VehicleRecord{}, fmt.Errorf("lookup vehicle: %w", err)
	}
	if currentCredentialID.Valid {
		record.CurrentCredentialID = currentCredentialID.String
	}
	return record, nil
}

func (s *service) nextCredentialVersion(ctx context.Context, subjectType, subjectID, purpose string) (int, error) {
	var version int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1 FROM credentials WHERE subject_type = ? AND subject_id = ? AND purpose = ?
	`, subjectType, subjectID, purpose).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("get next credential version: %w", err)
	}
	return version, nil
}

func ensureAllowedKEM(algorithm string) error {
	if _, ok := allowedKEMAlgorithms[algorithm]; ok {
		return nil
	}
	return fmt.Errorf("unsupported KEM algorithm %q", algorithm)
}

func ensureAllowedSignature(algorithm string) error {
	if _, ok := allowedSignatureAlgorithms[algorithm]; ok {
		return nil
	}
	return fmt.Errorf("unsupported signature algorithm %q", algorithm)
}

func validateAlgorithmForPurpose(algorithm, purpose string) error {
	if purpose == "kem" {
		return ensureAllowedKEM(algorithm)
	}
	return ensureAllowedSignature(algorithm)
}

func randomID(prefix string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buf), nil
}

func deriveCredentialStatus(current, validTo, revokedAt string) string {
	if revokedAt != "" || current == "revoked" {
		return "revoked"
	}
	if current == "superseded" {
		return "superseded"
	}
	if validTo == "" {
		return current
	}
	parsed, err := time.Parse(time.RFC3339, validTo)
	if err == nil && parsed.Before(time.Now().UTC()) {
		return "expired"
	}
	return current
}

func normalizeSubjectType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "vehicle", "rsu", "infrastructure":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizePurpose(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "signing", "signature":
		return "signing"
	case "kem":
		return "kem"
	default:
		return ""
	}
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "medium", "high", "critical":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func buildJoinPayload(vehicleID, rsuID, credentialID, kemAlgorithm, vehicleCertificate, ciphertext string) []byte {
	joined := strings.Join([]string{
		strings.TrimSpace(vehicleID),
		strings.TrimSpace(rsuID),
		strings.TrimSpace(credentialID),
		strings.TrimSpace(kemAlgorithm),
		strings.TrimSpace(vehicleCertificate),
		strings.TrimSpace(ciphertext),
	}, "\n")
	hash := sha256.Sum256([]byte(joined))
	return hash[:]
}