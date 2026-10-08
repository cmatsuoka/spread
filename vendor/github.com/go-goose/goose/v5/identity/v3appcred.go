package identity

import (
	"fmt"

	goosehttp "github.com/go-goose/goose/v5/http"
)

// v3AuthApplicationCredential contains an application credential
// authentication request, as described in:
// https://docs.openstack.org/api-ref/identity/v3/#application-credentials
type v3AuthApplicationCredential struct {
	ID     string `json:"id,omitempty"`
	Secret string `json:"secret"`
}

// V3AppCred is an Authenticator that will perform application
// credential authentication using the v3 protocol.
type V3AppCred struct {
	client goosehttp.HttpClient
}

// Auth performs a v3 application credential authentication request
// using the values supplied in creds.
func (a *V3AppCred) Auth(creds *Credentials) (*AuthDetails, error) {
	if a.client == nil {
		a.client = goosehttp.New()
	}
	if creds.ApplicationCredentialID == "" {
		return nil, fmt.Errorf("application credential id not specified")
	}
	if creds.ApplicationCredentialSecret == "" {
		return nil, fmt.Errorf("application credential secret not specified")
	}
	auth := v3AuthWrapper{
		Auth: v3AuthRequest{
			Identity: v3AuthIdentity{
				Methods: []string{"application_credential"},
				ApplicationCredential: &v3AuthApplicationCredential{
					ID:     creds.ApplicationCredentialID,
					Secret: creds.ApplicationCredentialSecret,
				},
			},
		},
	}
	return v3KeystoneAuth(a.client, &auth, creds.URL)
}
