package pqc

type KEMCheckResult struct {
	LiboqsVersion         string   `json:"liboqsVersion"`
	EnabledKEMs           []string `json:"enabledKEMs"`
	KEMName               string   `json:"kemName"`
	Details               string   `json:"details"`
	SharedSecretsCoincide bool     `json:"sharedSecretsCoincide"`
}

type RSUBeacon struct {
	RSUID        string `json:"rsuId"`
	KEMAlgorithm string `json:"kemAlgorithm"`
	KEMPublicKey string `json:"kemPublicKey"`
	KeyVersion   string `json:"keyVersion"`
	Details      string `json:"details"`
}
