# ES256 keys loaded from PEM paths; generate via CLI

Server mode loads ES256 signing material from PEM path(s) given on the CLI and fails closed if missing or invalid. It does not auto-generate keys on boot. A separate CLI command exists to generate key material for Operator use.
