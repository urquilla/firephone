package firephone

type startVerificationRequestBody struct {
	PhoneNumber    string `json:"phoneNumber"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type startVerificationResponse struct {
	SessionInfo string `json:"sessionInfo"`
}

type completeVerificationRequestBody struct {
	SessionInfo string `json:"sessionInfo"`
	Code        string `json:"code"`
}

type completeVerificationResponse struct {
	IDToken     string `json:"idToken"`
	IsNewUser   bool   `json:"isNewUser"`
	PhoneNumber string `json:"phoneNumber"`
	// Library is only intended for phone number verification
	// Leaving session handling out of scope
	// RefreshToken string `json:"refreshToken"`
	// ExpiresIn   int    `json:"expiresIn"`
	// LocalId     int    `json:"localId"`
}

type VerificationInfo struct {
	IDToken     string
	IsNewUser   bool
	PhoneNumber string
}
