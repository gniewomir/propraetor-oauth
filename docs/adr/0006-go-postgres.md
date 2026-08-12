# Go and Postgres for v1

The Authorization Server is implemented in Go with Postgres as the system of record. Database credentials come from the environment. Chosen for boring operational defaults and transactional semantics around Authorization Codes and Refresh Tokens.
