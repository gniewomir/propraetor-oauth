# JWT Access Tokens

Access Tokens are JWTs so external Resource Servers can validate them locally without calling this Authorization Server. Opaque Access Tokens were rejected because they would require an introspection API (out of v1 scope) or shared database access (breaks the AS-only boundary).
