# Authorization Request state entropy bar

`state` is required on the Authorization Code Authorization Grant (ADR-0004). The AS rejects `state` outside 22–512 characters or outside the RFC 7636 unreserved charset (`A-Za-z0-9-._~`). Valid `state` is bound into the Authorization Session (ADR-0048) and returned unchanged on success and error redirects to the Client. No global `state` replay cache in v1—session binding covers mid-flow swap; Clients remain responsible for CSRF against their own redirect.
