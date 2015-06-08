package acme

type registrationMessage struct {
	Contact []string `json:"contact"`
}

// Registration is returned by the ACME server after the registration
// The client implementation should save this registration somewhere.
type Registration struct {
	ID  int `json:"id"`
	Key struct {
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"key"`
	Recoverytoken string   `json:"recoveryToken"`
	Contact       []string `json:"contact"`
}

// RegistrationResource represents all important informations about a registration
// of which the client needs to keep track itself.
type RegistrationResource struct {
	Body        Registration
	URI         string
	NewAuthzURL string
	TosURL      string
}
