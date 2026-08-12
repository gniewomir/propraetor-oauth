# v2 browser token crossroads: A, B, D only

v1 uses RFC 6749 JSON refresh for all Public Clients (ADR-0050). For stronger browser posture, v2 exploration is limited to crossroads **A** (HttpOnly refresh cookie on AS), **B** (AS token-mediating backend), and **D** (BFF or AS API proxy). Crossroads C and E are set aside. Detail in docs/crossroads/v2-public-client-token-handling.md.
