# Exact Redirect URIs: custom schemes and loopback allowed

Redirect URI matching remains exact string equality with no wildcards and no localhost port wildcards (ADR-0012). Operators may register custom URI schemes (e.g. `com.example.app:/callback`) and explicit loopback URIs (e.g. `http://127.0.0.1:8080/callback`) for native/dev Clients. Each distinct port or path is a separate registration. Registration friction is preferred over wildcard Redirect URI risk; RFC 8252 is in the baseline (ADR-0004) without relaxing exact match.
