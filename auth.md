# Engine authentication and configuration encryption profile

This document is normative for the Databacker engine protocol. All byte
strings shown in quotes are their exact UTF-8 bytes. Integers are unsigned
64-bit big-endian unless stated otherwise.

[`testdata/auth-v2.json`](testdata/auth-v2.json) contains deterministic
cross-implementation fixtures. Every private value in that file is test data
and must never be used as a real credential.

## Local credential and derived keys

An engine credential seed is exactly 32 uniformly random bytes. Configuration
uses strict standard padded base64. New credentials start both generations at
1; generation 0 is invalid.

Derive a pseudorandom key with HKDF-Extract-SHA-256:

```text
PRK = HKDF-Extract-SHA256(
    salt = "databacker engine credential v2",
    IKM  = seed
)
```

Derive the Ed25519 seed and X25519 private input independently:

```text
authentication_seed = HKDF-Expand-SHA256(
    PRK,
    "databacker/authentication/ed25519/v1/generation/" || uint64_be(authentication_generation),
    32
)

configuration_private_input = HKDF-Expand-SHA256(
    PRK,
    "databacker/configuration-encryption/x25519/v1/generation/" || uint64_be(configuration_generation),
    32
)
```

Use `authentication_seed` with Ed25519's seed constructor. Use
`configuration_private_input` with the platform X25519 private-key constructor;
do not clamp it separately in application code.

## Deterministic key IDs (fingerprints)

Key IDs are deterministic, domain-separated fingerprints, not arbitrary IDs
allocated by a controller or database. Public keys are raw 32-byte values. A
key-ID digest is SHA-256 over exactly one of these byte strings:

```text
"databacker/engine-key-id/authentication/ed25519/v1\x00"
|| uint64_be(generation)
|| raw_ed25519_public_key

"databacker/engine-key-id/configuration-encryption/x25519/v1\x00"
|| uint64_be(generation)
|| raw_x25519_public_key
```

The display forms are `auth:` or `config:` followed by the complete digest as
64 lowercase hexadecimal characters. Anyone with the public key and generation
computes the same ID. The engine derives the public key from its local seed
and generation, so it does not store a server-issued key ID. During
registration, the controller computes the ID from the submitted public key and
generation and stores that fingerprint with the key and engine record. It MUST
recompute and reject any supplied ID that differs rather than trusting it.

The generation and domain separator make the fingerprint purpose- and
version-specific. Consequently, an authentication key and a configuration key
cannot share an ID even if their raw public-key bytes happen to be identical.

## HTTP Message Signatures

Every request uses HTTPS and RFC 9421 HTTP Message Signatures with Ed25519.
The only accepted signature label is `databacker-engine`. Exactly one member
with that label must occur in both `Signature-Input` and `Signature`.

Required signature parameters are:

- `created`: integer Unix time;
- `expires`: integer Unix time, later than `created` and no more than 300
  seconds after it;
- `keyid`: canonical deterministic `auth:` key fingerprint;
- `nonce`: exactly 16 random bytes encoded as unpadded base64url;
- `tag`: exactly `databacker-engine-v1`.

At verification time, `created` may be no more than 300 seconds old and no
more than 30 seconds in the future. The verifier allows at most 30 seconds of
negative clock skew after `expires`. Authentication failures return a uniform
401 response. Time failures include the ordinary HTTP `Date` response header.

Requests without a body cover these components in this order:

```text
"@method" "@authority" "@path" "@query"
```

Requests with a body cover these components in this order:

```text
"@method" "@authority" "@path" "@query"
"content-digest";sf "content-type" "idempotency-key"
```

RFC 9421 includes `@signature-params` as the final signature-base line. The
profile does not use a separate engine-ID header and does not carry the public
key on every request. The `keyid` fingerprint indexes the already-registered
authentication public key and its engine record. The verifier uses that stored
public key to check the request signature. Engine routes are self-only: the
resulting authenticated engine is the sole target, and callers cannot select
another engine through the path, query, or body.

`@path` and `@query` are constructed from the request target actually sent or
received. Percent-encoded octets are not decoded and re-encoded. Query order,
repeated keys, `+`, and `%20` remain distinct. An absent query has the RFC 9421
value `?`. Redirects are not followed for signed engine requests.

For a body, `Content-Digest` is an RFC 9530 Dictionary containing exactly a
SHA-256 digest of the transmitted representation:

```text
Content-Digest: sha-256=:STANDARD_BASE64_DIGEST:
```

The signature covers its strictly serialized Structured Fields value using the
`sf` component parameter. Profile v1 permits only identity content encoding.
`Idempotency-Key` is a UUID generated once for the logical operation and reused
for every retry; the nonce and signature timestamps are new for every HTTP
attempt.

The verifier rejects extra labels, duplicate or ambiguous fields, unsupported
components or parameters, and signatures whose request representation cannot
be reconstructed exactly. It resolves `keyid`, enforces time and key state,
verifies the body digest, verifies the Ed25519 signature, enforces idempotency,
and only then dispatches to the handler. Handlers obtain the engine exclusively
from authenticated request context.

The OpenAPI `apiKey` security schemes only declare the two RFC 9421 headers for
clients and generated tooling. Generated OpenAPI wrappers do not implement this
profile. A server MUST install verification middleware in front of every
generated handler and MUST NOT treat the generated security-scope context as
proof of authentication.

## Encrypted configuration v2

`GET /engines/config` returns a complete `Config` for the authenticated engine.
When `kind` is `encrypted`, `spec` is the v2 `EncryptedSpec`. Successful
decryption yields another complete `Config` encoded as UTF-8 JSON; it may be
`local`, `remote`, or `encrypted`.

The mandatory suite is X25519, HKDF-SHA-256, and ChaCha20-Poly1305. Public keys,
the 12-byte nonce, and ciphertext with its 16-byte tag use strict standard
padded base64.

Let `opaque(x)` be a four-byte unsigned big-endian length followed by `x`.
Canonical additional authenticated data is:

```text
opaque("databacker/configuration-envelope-aad/v2")
|| opaque(UTF8(envelope_version))
|| uint64_be(configuration_version)
|| opaque(UTF8(recipient_key_id))
|| uint64_be(recipient_generation)
|| opaque(UTF8(key_agreement))
|| opaque(UTF8(key_derivation))
|| opaque(UTF8(encryption))
|| opaque(UTF8(plaintext_media_type))
|| opaque(raw_sender_public_key)
|| opaque(raw_nonce)
```

Derive the 32-byte AEAD key from the X25519 shared secret as follows:

```text
envelope_PRK = HKDF-Extract-SHA256(
    salt = "databacker configuration encryption v2",
    IKM  = x25519_shared_secret
)

AEAD_key = HKDF-Expand-SHA256(
    envelope_PRK,
    "databacker/configuration-aead-key/chacha20-poly1305/v2\x00" || AAD,
    32
)
```

`recipientKeyId` is the deterministic `config:` fingerprint computed from the
recipient X25519 public key and generation above; it is not a database-assigned
identifier. An encryptor computes it from the registered recipient key, and an
engine recomputes it from its derived key before accepting the envelope.

The recipient key ID and generation bind the envelope to the intended engine
key. Cloud's database-assigned engine ID is deliberately absent: it is neither
locally derivable nor needed for decryption, and authentication-key rotation
must remain independent of configuration encryption.

Plaintext is limited to 4 MiB. Validate envelope algorithms, identifiers,
decoded lengths, version monotonicity, and recipient key identity before
decryption. Never log plaintext, the seed, derived private keys, the shared
secret, signatures, or nonces.
