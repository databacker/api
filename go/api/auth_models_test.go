package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEngineCredentialsJSONShape(t *testing.T) {
	retained := []uint64{1, 2}
	want := EngineCredentials{
		Version:                           DatabackerCredentialsV2,
		Seed:                              strings.Repeat("A", 43) + "=",
		AuthenticationGeneration:          3,
		ConfigurationEncryptionGeneration: 4,
		RetainedConfigurationEncryptionGenerations: &retained,
	}

	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal credentials: %v", err)
	}

	var got EngineCredentials
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal credentials: %v", err)
	}

	if got.Version != DatabackerCredentialsV2 {
		t.Fatalf("version = %q, want %q", got.Version, DatabackerCredentialsV2)
	}
	if got.Seed != want.Seed {
		t.Fatal("seed did not survive credential round trip")
	}
	if got.AuthenticationGeneration != 3 || got.ConfigurationEncryptionGeneration != 4 {
		t.Fatalf("unexpected generations: auth=%d config=%d", got.AuthenticationGeneration, got.ConfigurationEncryptionGeneration)
	}
	if got.RetainedConfigurationEncryptionGenerations == nil || len(*got.RetainedConfigurationEncryptionGenerations) != 2 {
		t.Fatalf("unexpected retained generations: %#v", got.RetainedConfigurationEncryptionGenerations)
	}
}

func TestEncryptedSpecJSONShape(t *testing.T) {
	want := EncryptedSpec{
		Version:              DatabackerEncryptedConfigV2,
		ConfigurationVersion: 7,
		RecipientKeyID:       "config:" + strings.Repeat("a", 64),
		RecipientGeneration:  2,
		KeyAgreement:         X25519,
		KeyDerivation:        HkdfSha256,
		Encryption:           EncryptedSpecEncryptionChacha20Poly1305,
		SenderPublicKey:      strings.Repeat("A", 43) + "=",
		Nonce:                strings.Repeat("A", 16),
		Ciphertext:           "AAAAAAAAAAAAAAAAAAAAAA==",
		PlaintextMediaType:   ApplicationJSON,
	}

	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal encrypted spec: %v", err)
	}

	var got EncryptedSpec
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal encrypted spec: %v", err)
	}

	if got != want {
		t.Fatalf("encrypted spec round trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}
