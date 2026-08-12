# Rate limit dimensions, fixed window, 429, explicit CLI limits

Rate limits use fixed windows in v1. Dimensions: global per source IP; login per IP and per username; Authorization Endpoint and Token Endpoint per IP and per Client Identifier when known. Login is the strictest tier (ADR-0038). Exceeded limits return HTTP 429 with Retry-After. All limit thresholds and windows are required CLI configuration at server start—no silent defaults. Permissive ceilings refuse overly loose configs at start (ADR-0059). Account lockout is out of v1 (ADR-0061).
