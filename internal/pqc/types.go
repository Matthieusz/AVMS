package pqc

type KEMCheckResult struct {
	LiboqsVersion         string   `json:"liboqsVersion"`
	EnabledKEMs           []string `json:"enabledKEMs"`
	KEMName               string   `json:"kemName"`
	Details               string   `json:"details"`
	SharedSecretsCoincide bool     `json:"sharedSecretsCoincide"`
}
