package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Matthieusz/AVMS/internal/database"
	"github.com/Matthieusz/AVMS/internal/pqc"
)

func (s *Server) rsuBeaconHandler(c *gin.Context) {
	rsuID := strings.TrimSpace(c.Param("rsuID"))
	if rsuID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rsuId is required"})
		return
	}

	kemAlgorithm := strings.TrimSpace(c.Query("kemAlgorithm"))
	if kemAlgorithm == "" {
		kemAlgorithm = database.RecommendedKEMAlgorithm
	}

	beacon, err := pqc.GetRSUBeacon(rsuID, kemAlgorithm)
	if err != nil {
		respondWithDomainError(c, "get RSU beacon", err)
		return
	}

	c.JSON(http.StatusOK, beacon)
}

type registerVehicleRequest struct {
	VehicleID          string `json:"vehicleId"`
	Manufacturer       string `json:"manufacturer"`
	HardwareProfile    string `json:"hardwareProfile"`
	PublicKey          string `json:"publicKey"`
	SignatureAlgorithm string `json:"signatureAlgorithm"`
}

type issueCredentialRequest struct {
	SubjectType  string `json:"subjectType"`
	SubjectID    string `json:"subjectId"`
	Algorithm    string `json:"algorithm"`
	Purpose      string `json:"purpose"`
	PublicKey    string `json:"publicKey"`
	ValidityDays int    `json:"validityDays"`
}

type revokeCredentialRequest struct {
	CredentialID string `json:"credentialId"`
	Reason       string `json:"reason"`
}

type rotateKeyRequest struct {
	VehicleID    string `json:"vehicleId"`
	Algorithm    string `json:"algorithm"`
	PublicKey    string `json:"publicKey"`
	Reason       string `json:"reason"`
	ValidityDays int    `json:"validityDays"`
}

type joinVehicleRequest struct {
	VehicleID          string `json:"vehicleId"`
	RSUID              string `json:"rsuId"`
	CredentialID       string `json:"credentialId"`
	KEMAlgorithm       string `json:"kemAlgorithm"`
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	Ciphertext         string `json:"ciphertext"`
	Signature          string `json:"signature"`
	VehicleCertificate string `json:"vehicleCertificate"`
}

type reportIncidentRequest struct {
	SubjectType  string `json:"subjectType"`
	SubjectID    string `json:"subjectId"`
	CredentialID string `json:"credentialId"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
}

func (s *Server) registerVehicleHandler(c *gin.Context) {
	var payload registerVehicleRequest
	if !bindJSONPayload(c, &payload) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.RegisterVehicle(ctx, database.RegisterVehicleInput{
		VehicleID:          payload.VehicleID,
		Manufacturer:       payload.Manufacturer,
		HardwareProfile:    payload.HardwareProfile,
		PublicKey:          payload.PublicKey,
		SignatureAlgorithm: payload.SignatureAlgorithm,
	})
	if err != nil {
		respondWithDomainError(c, "register vehicle", err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (s *Server) issueCredentialHandler(c *gin.Context) {
	var payload issueCredentialRequest
	if !bindJSONPayload(c, &payload) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.IssueCredential(ctx, database.IssueCredentialInput{
		SubjectType:  payload.SubjectType,
		SubjectID:    payload.SubjectID,
		Algorithm:    payload.Algorithm,
		Purpose:      payload.Purpose,
		PublicKey:    payload.PublicKey,
		ValidityDays: payload.ValidityDays,
	})
	if err != nil {
		respondWithDomainError(c, "issue credential", err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (s *Server) revokeCredentialHandler(c *gin.Context) {
	var payload revokeCredentialRequest
	if !bindJSONPayload(c, &payload) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.RevokeCredential(ctx, database.RevokeCredentialInput{
		CredentialID: payload.CredentialID,
		Reason:       payload.Reason,
	})
	if err != nil {
		respondWithDomainError(c, "revoke credential", err)
		return
	}

	c.JSON(http.StatusOK, record)
}

func (s *Server) rotateKeysHandler(c *gin.Context) {
	var payload rotateKeyRequest
	if !bindJSONPayload(c, &payload) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.RotateKeys(ctx, database.RotateKeyInput{
		VehicleID:    payload.VehicleID,
		Algorithm:    payload.Algorithm,
		PublicKey:    payload.PublicKey,
		Reason:       payload.Reason,
		ValidityDays: payload.ValidityDays,
	})
	if err != nil {
		respondWithDomainError(c, "rotate keys", err)
		return
	}

	c.JSON(http.StatusOK, record)
}

func (s *Server) credentialStatusHandler(c *gin.Context) {
	credentialID := strings.TrimSpace(c.Param("credentialID"))
	if credentialID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credentialId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.GetCredentialStatus(ctx, credentialID)
	if err != nil {
		respondWithDomainError(c, "get credential status", err)
		return
	}

	c.JSON(http.StatusOK, record)
}

func (s *Server) currentPolicyHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	policy, err := s.db.GetCurrentPolicy(ctx)
	if err != nil {
		respondWithDomainError(c, "get current policy", err)
		return
	}

	c.JSON(http.StatusOK, policy)
}

func (s *Server) joinVehicleHandler(c *gin.Context) {
	var payload joinVehicleRequest
	if !bindJSONPayload(c, &payload) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.JoinVehicle(ctx, database.JoinVehicleInput{
		VehicleID:          payload.VehicleID,
		RSUID:              payload.RSUID,
		CredentialID:       payload.CredentialID,
		KEMAlgorithm:       payload.KEMAlgorithm,
		SignatureAlgorithm: payload.SignatureAlgorithm,
		Ciphertext:         payload.Ciphertext,
		Signature:          payload.Signature,
		VehicleCertificate: payload.VehicleCertificate,
	})
	if err != nil {
		respondWithDomainError(c, "join vehicle", err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (s *Server) reportIncidentHandler(c *gin.Context) {
	var payload reportIncidentRequest
	if !bindJSONPayload(c, &payload) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	record, err := s.db.ReportIncident(ctx, database.IncidentReportInput{
		SubjectType:  payload.SubjectType,
		SubjectID:    payload.SubjectID,
		CredentialID: payload.CredentialID,
		Severity:     payload.Severity,
		Description:  payload.Description,
	})
	if err != nil {
		respondWithDomainError(c, "report incident", err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

func bindJSONPayload(c *gin.Context, payload any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateItemBodySize)
	if err := c.ShouldBindJSON(payload); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body is too large"})
			return false
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return false
	}
	return true
}

func respondWithDomainError(c *gin.Context, operation string, err error) {
	status := http.StatusBadRequest
	message := err.Error()

	switch {
	case errors.Is(err, database.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, database.ErrConflict):
		status = http.StatusConflict
	case strings.HasPrefix(message, "register vehicle:") || strings.HasPrefix(message, "issue credential:") || strings.HasPrefix(message, "lookup credential:") || strings.HasPrefix(message, "lookup vehicle:") || strings.HasPrefix(message, "record join session:") || strings.HasPrefix(message, "report incident:"):
		status = http.StatusInternalServerError
	}

	if status >= http.StatusInternalServerError {
		logServerError(c, operation, err)
	}

	c.JSON(status, gin.H{"error": message})
}