# Authorization Session cookie: __Host- when https

Authorization Session cookies use a fixed name, `Path=/`, no `Domain`, plus ADR-0048 attributes (HttpOnly; Secure when Issuer URL is https; SameSite=Lax). When the Issuer URL scheme is https, the cookie name uses the `__Host-` prefix. When the Issuer URL is http (ADR-0018 cleartext public URLs), the same fixed name is used without `__Host-` (prefix requires Secure). Cookie name is not Operator-configurable.
