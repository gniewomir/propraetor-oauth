# Authorization Server only — no in-process Resource Server

v1 is an Authorization Server that issues Access Tokens (and Refresh Tokens where applicable) for use at external Resource Servers. Enforcing access to Protected Resources is out of process and out of repo. There is no RFC 7009 revocation endpoint in v1 (ADR-0003).
