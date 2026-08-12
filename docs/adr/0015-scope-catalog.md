# Operator-managed Scope catalog with per-Client allowlists

Scopes are an Operator-managed catalog stored in the database (not a compile-time fixed set). The Operator CLI defines/deletes Scope names and adds/removes Scopes on a Client. Access Tokens carry the granted Scopes. Free-form scope strings and scopeless Access Tokens are out of v1.
