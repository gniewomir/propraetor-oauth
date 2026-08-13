package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/gniewomir/propraetor-oauth/internal/domain"
)

func TestParseClientIDRejectsEmptyAndIllegalCharset(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "has space", `quote"x`, `slash\x`, "a\nb"} {
		_, err := domain.ParseClientID(in)
		if !errors.Is(err, domain.ErrInvalidClientID) {
			t.Fatalf("ParseClientID(%q) err = %v, want ErrInvalidClientID", in, err)
		}
	}
}

func TestParseClientIDAcceptsScopeTokenCharset(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseClientID("spa-web_1!~$&'()*+,;=:@/")
	if err != nil {
		t.Fatalf("ParseClientID: %v", err)
	}
	if got := id.String(); got != "spa-web_1!~$&'()*+,;=:@/" {
		t.Fatalf("String = %q", got)
	}
}

func TestEnsureClientIDAvailableRejectsASCIICaseOnlyDuplicate(t *testing.T) {
	t.Parallel()

	existingID, err := domain.ParseClientID("Acme")
	if err != nil {
		t.Fatalf("ParseClientID existing: %v", err)
	}
	candidate, err := domain.ParseClientID("acme")
	if err != nil {
		t.Fatalf("ParseClientID candidate: %v", err)
	}

	err = domain.EnsureClientIDAvailable(candidate, []domain.ClientID{existingID})
	if !errors.Is(err, domain.ErrClientIDConflict) {
		t.Fatalf("err = %v, want ErrClientIDConflict", err)
	}
}

func TestEnsureClientIDAvailableAllowsDistinctIDs(t *testing.T) {
	t.Parallel()

	existingID, err := domain.ParseClientID("acme")
	if err != nil {
		t.Fatalf("ParseClientID existing: %v", err)
	}
	candidate, err := domain.ParseClientID("acme-admin")
	if err != nil {
		t.Fatalf("ParseClientID candidate: %v", err)
	}

	if err := domain.EnsureClientIDAvailable(candidate, []domain.ClientID{existingID}); err != nil {
		t.Fatalf("EnsureClientIDAvailable: %v", err)
	}
}

func TestNewPublicClient(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseClientID("spa")
	if err != nil {
		t.Fatalf("ParseClientID: %v", err)
	}
	c, err := domain.NewClient(id, domain.AuthMethodNone, "", "https://rs.example")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.ID().String() != "spa" || c.AuthMethod() != domain.AuthMethodNone || c.SecretHash() != "" {
		t.Fatalf("unexpected client: id=%q method=%q secret=%q", c.ID(), c.AuthMethod(), c.SecretHash())
	}
	if c.Audience() != "https://rs.example" {
		t.Fatalf("audience = %q", c.Audience())
	}
	if c.DeactivatedAt() != nil || c.NotBefore() != nil {
		t.Fatalf("new client should have nil deactivated_at and not_before")
	}
}

func TestNewConfidentialClientRequiresSecretHash(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseClientID("svc")
	if err != nil {
		t.Fatalf("ParseClientID: %v", err)
	}
	_, err = domain.NewClient(id, domain.AuthMethodClientSecretBasic, "", "https://rs.example")
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("err = %v, want ErrInvalidClient", err)
	}
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("err = %v, want secret mention", err)
	}
}

func TestNewPublicClientRejectsSecretHash(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseClientID("spa")
	if err != nil {
		t.Fatalf("ParseClientID: %v", err)
	}
	_, err = domain.NewClient(id, domain.AuthMethodNone, "argon2id$...", "https://rs.example")
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("err = %v, want ErrInvalidClient", err)
	}
}

func TestNewClientRejectsEmptyAudienceAndUnknownAuthMethod(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseClientID("spa")
	if err != nil {
		t.Fatalf("ParseClientID: %v", err)
	}
	_, err = domain.NewClient(id, domain.AuthMethodNone, "", "")
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("empty audience err = %v, want ErrInvalidClient", err)
	}
	_, err = domain.NewClient(id, domain.AuthMethod("client_secret_post"), "", "https://rs.example")
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("unknown method err = %v, want ErrInvalidClient", err)
	}
}
