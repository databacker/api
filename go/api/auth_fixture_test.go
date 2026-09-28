package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func TestAuthenticationFixtureSignature(t *testing.T) {
	data, err := os.ReadFile("../../testdata/auth-v2.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture struct {
		AuthenticationPublicKeyBase64 string           `json:"authenticationPublicKeyBase64"`
		Signature                     signatureFixture `json:"signature"`
		BodySignature                 signatureFixture `json:"bodySignature"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	publicKey, err := base64.StdEncoding.DecodeString(fixture.AuthenticationPublicKeyBase64)
	if err != nil {
		t.Fatalf("decode public key: %v", err)
	}
	for name, signed := range map[string]signatureFixture{
		"bodyless request": fixture.Signature,
		"body request":     fixture.BodySignature,
	} {
		t.Run(name, func(t *testing.T) {
			signature, err := base64.StdEncoding.DecodeString(signed.SignatureBase64)
			if err != nil {
				t.Fatalf("decode signature: %v", err)
			}
			if !ed25519.Verify(publicKey, []byte(signed.Base), signature) {
				t.Fatal("fixture signature did not verify")
			}
		})
	}
}

type signatureFixture struct {
	Base            string `json:"base"`
	SignatureBase64 string `json:"signatureBase64"`
}
