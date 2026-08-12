# Access Token JWT claims

Access Tokens include at least: iss, sub, aud, exp, iat, jti, client_id, and scope. sub is the Resource Owner (End-User) id for the Authorization Code grant and the Client Identifier for Client Credentials. aud is the Client’s single registered Audience (ADR-0016). Issuer URL populates iss. `jti` is for uniqueness and audit correlation only—not an Access Token revocation or denylist mechanism in v1.
