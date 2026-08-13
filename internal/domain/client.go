package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Client Identifier and Client catalog rules (ADR-0013, 0016, 0075, 0069, 0077).

var (
	ErrInvalidClientID  = errors.New("invalid client id")
	ErrClientIDConflict = errors.New("client id conflict")
	ErrInvalidClient    = errors.New("invalid client")
)

// AuthMethod is how the Client authenticates at the Token Endpoint (ADR-0013).
type AuthMethod string

const (
	AuthMethodNone              AuthMethod = "none"
	AuthMethodClientSecretBasic AuthMethod = "client_secret_basic"
)

// ClientID is an Operator-supplied Client Identifier (natural PK, ADR-0077).
type ClientID struct {
	value string
}

// ParseClientID validates RFC 6749 §3.3 scope-token charset (ADR-0077).
func ParseClientID(raw string) (ClientID, error) {
	if raw == "" || !isScopeToken(raw) {
		return ClientID{}, fmt.Errorf("%w: %q", ErrInvalidClientID, raw)
	}
	return ClientID{value: raw}, nil
}

func (id ClientID) String() string { return id.value }

// EnsureClientIDAvailable rejects a candidate that matches an existing id
// exactly or differs only by ASCII case (ADR-0077).
func EnsureClientIDAvailable(candidate ClientID, existing []ClientID) error {
	for _, id := range existing {
		if strings.EqualFold(candidate.value, id.value) {
			return fmt.Errorf("%w: %q", ErrClientIDConflict, candidate.value)
		}
	}
	return nil
}

// Client is a provisioned OAuth Client.
type Client struct {
	id            ClientID
	authMethod    AuthMethod
	secretHash    string
	audience      string
	deactivatedAt *time.Time
	notBefore     *time.Time
}

// NewClient constructs an active Client with no Not-Before watermark.
func NewClient(id ClientID, method AuthMethod, secretHash, audience string) (Client, error) {
	if strings.TrimSpace(audience) == "" {
		return Client{}, fmt.Errorf("%w: audience is required", ErrInvalidClient)
	}
	switch method {
	case AuthMethodNone:
		if secretHash != "" {
			return Client{}, fmt.Errorf("%w: public client must not have secret hash", ErrInvalidClient)
		}
	case AuthMethodClientSecretBasic:
		if secretHash == "" {
			return Client{}, fmt.Errorf("%w: confidential client requires secret hash", ErrInvalidClient)
		}
	default:
		return Client{}, fmt.Errorf("%w: unsupported auth method %q", ErrInvalidClient, method)
	}
	return Client{
		id:         id,
		authMethod: method,
		secretHash: secretHash,
		audience:   audience,
	}, nil
}

func (c Client) ID() ClientID              { return c.id }
func (c Client) AuthMethod() AuthMethod    { return c.authMethod }
func (c Client) SecretHash() string        { return c.secretHash }
func (c Client) Audience() string          { return c.audience }
func (c Client) DeactivatedAt() *time.Time { return c.deactivatedAt }
func (c Client) NotBefore() *time.Time     { return c.notBefore }

// isScopeToken reports whether s matches RFC 6749 §3.3 scope-token:
// 1*( %x21 / %x23-5B / %x5D-7E ).
func isScopeToken(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
		switch {
		case r == 0x21:
		case r >= 0x23 && r <= 0x5B:
		case r >= 0x5D && r <= 0x7E:
		default:
			return false
		}
	}
	return true
}
