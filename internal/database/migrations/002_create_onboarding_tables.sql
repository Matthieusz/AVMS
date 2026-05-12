CREATE TABLE IF NOT EXISTS vehicles (
    vehicle_id TEXT PRIMARY KEY,
    manufacturer TEXT NOT NULL,
    hardware_profile TEXT NOT NULL,
    public_key TEXT NOT NULL,
    signature_algorithm TEXT NOT NULL,
    status TEXT NOT NULL,
    current_credential_id TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS credentials (
    credential_id TEXT PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    purpose TEXT NOT NULL,
    public_key TEXT NOT NULL,
    status TEXT NOT NULL,
    version INTEGER NOT NULL,
    valid_from TEXT NOT NULL,
    valid_to TEXT NOT NULL,
    issued_at TEXT NOT NULL,
    revoked_at TEXT,
    revoke_reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_credentials_subject ON credentials(subject_type, subject_id, purpose);
CREATE INDEX IF NOT EXISTS idx_credentials_status ON credentials(status);

CREATE TABLE IF NOT EXISTS join_sessions (
    session_id TEXT PRIMARY KEY,
    vehicle_id TEXT NOT NULL,
    rsu_id TEXT NOT NULL,
    credential_id TEXT NOT NULL,
    kem_algorithm TEXT NOT NULL,
    signature_algorithm TEXT NOT NULL,
    ciphertext TEXT NOT NULL,
    signature TEXT NOT NULL,
    session_key_ref TEXT NOT NULL,
    status TEXT NOT NULL,
    verification_notes TEXT NOT NULL,
    created_at TEXT NOT NULL,
    accepted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_join_sessions_vehicle ON join_sessions(vehicle_id, created_at DESC);

CREATE TABLE IF NOT EXISTS incidents (
    incident_id TEXT PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    credential_id TEXT,
    severity TEXT NOT NULL,
    description TEXT NOT NULL,
    recommended_action TEXT NOT NULL,
    status TEXT NOT NULL,
    reported_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_incidents_subject ON incidents(subject_type, subject_id, reported_at DESC);