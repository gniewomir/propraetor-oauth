# Client Authentication methods

Public Clients use Client Authentication method `none` and rely on PKCE at the Token Endpoint. Confidential Clients use `client_secret_basic` only. Other methods (including `client_secret_post`) are rejected.
