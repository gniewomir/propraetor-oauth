# Grant matrix by client type

Public Clients use the Authorization Code Authorization Grant with required PKCE and state, and receive Refresh Tokens (all Public Clients — ADR-0050). Confidential Clients use the Client Credentials Authorization Grant only — no Authorization Code and no Refresh Tokens in v1.
