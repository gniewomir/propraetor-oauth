# Exact Redirect URI matching only

After Resource Owner authorization, redirects carrying an Authorization Code must match a pre-registered Redirect URI by exact string equality. No wildcards and no localhost port wildcards in v1. Custom URI schemes and explicit loopback URIs may be registered (ADR-0066). Native/dev Clients register each exact Redirect URI.
