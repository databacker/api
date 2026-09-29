package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
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
		Envelope                      map[string]any   `json:"envelope"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if _, exists := fixture.Envelope["instance"]; exists {
		t.Fatal("fixture envelope contains obsolete engine instance")
	}
	if !strings.Contains(fixture.Signature.Base, `"@path": /engines/config`+"\n") {
		t.Fatal("bodyless fixture does not sign the self-only config route")
	}
	if !strings.Contains(fixture.BodySignature.Base, `"@path": /engines/telemetry/traces`+"\n") {
		t.Fatal("body fixture does not sign the self-only telemetry route")
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
