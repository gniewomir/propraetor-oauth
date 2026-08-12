# Authorization Session hardening

Authorization Sessions use cookie `HttpOnly`; `Secure` when the Issuer URL is https; `SameSite=Lax`. Cookie name, Path, Domain, and `__Host-` prefix rules are ADR-0064. On successful login the session id is rotated (anti-fixation). Login and Consent POSTs require a CSRF secret bound to the session. The session row is bound to the Authorization Request fields (`client_id`, `redirect_uri`, PKCE challenge, `state`, requested scopes) so those cannot be swapped mid-flow.
