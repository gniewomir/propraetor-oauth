package postgres

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// BootstrapInput is the Operator-chosen instance discriminator and connection
// target used to emit superuser SQL and OAUTH_STORAGE_URL (ADR-0072).
type BootstrapInput struct {
	Prefix   string
	Password string
	Host     string
	Port     int
	SSLMode  string // libpq sslmode; empty means "disable"
}

// BootstrapResult holds secret-bearing SQL and the app-role storage URL.
type BootstrapResult struct {
	SQL        string
	StorageURL string
}

var prefixPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// GenerateBootstrap builds CREATE ROLE / CREATE DATABASE SQL for a superuser
// and the matching OAUTH_STORAGE_URL for the app role ({prefix}_oauth).
func GenerateBootstrap(in BootstrapInput) (BootstrapResult, error) {
	if !prefixPattern.MatchString(in.Prefix) {
		return BootstrapResult{}, fmt.Errorf("prefix %q: must match %s", in.Prefix, prefixPattern.String())
	}
	if in.Password == "" {
		return BootstrapResult{}, fmt.Errorf("password is required")
	}
	if in.Host == "" {
		return BootstrapResult{}, fmt.Errorf("host is required")
	}
	if in.Port <= 0 || in.Port > 65535 {
		return BootstrapResult{}, fmt.Errorf("port %d: out of range", in.Port)
	}
	sslmode := in.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	name := in.Prefix + "_oauth"
	ident := quoteIdent(name)
	pass := quoteLiteral(in.Password)

	sql := strings.Join([]string{
		fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD %s;", ident, pass),
		fmt.Sprintf("CREATE DATABASE %s OWNER %s;", ident, ident),
	}, "\n") + "\n"

	q := url.Values{}
	q.Set("sslmode", sslmode)
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(name, in.Password),
		Host:     fmt.Sprintf("%s:%d", in.Host, in.Port),
		Path:     "/" + name,
		RawQuery: q.Encode(),
	}

	return BootstrapResult{SQL: sql, StorageURL: u.String()}, nil
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}
