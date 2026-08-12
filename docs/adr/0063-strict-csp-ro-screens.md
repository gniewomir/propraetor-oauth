# Strict CSP on Resource Owner screens

Login, Consent, and their static assets send a strict Content-Security-Policy: `default-src 'self'; script-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'self'`. No inline script or style (hash/`'self'` only); no CDNs (ADR-0035). XSS on the AS origin can hijack Authorization Sessions and, with JSON Refresh Tokens (ADR-0050), enable token exfiltration—CSP is the highest-leverage v1 browser hardening short of v2 crossroads A/B/D.
