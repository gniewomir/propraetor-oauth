# CLI writes a starter Server Policy file

A dedicated Operator CLI command writes a complete Server Policy TOML document to a path the Operator chooses (starter / scaffold values for every Policy schema key). This is an explicit file-generation step for Operator review and edit — not silent runtime defaults and not CLI overrides of TTL or rate-limit values at `server` start (ADR-0068). `oauth server` still requires `--policy` and fail-closed validation of a full Policy schema. Exact command spelling is `oauth policy init` (output path flag TBD at implementation).
