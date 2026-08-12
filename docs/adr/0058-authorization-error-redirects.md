# Authorization errors: redirect when redirect_uri is valid

When the Authorization Endpoint can validate a registered Redirect URI (exact match, ADR-0012), OAuth errors (`access_denied`, `invalid_scope`, etc.) are returned by redirecting to that URI with `error`, a non-sensitive `error_description`, and `state` when known. If `redirect_uri` is missing or does not exactly match a registered value, the AS serves an error page and does not redirect—preventing open redirects. Success redirects remain exact-match only.
