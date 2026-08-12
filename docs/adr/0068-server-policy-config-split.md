# Server Policy file; CLI topology; env secrets; hardcoded caps

Operator configuration for `cli server` is split by concern. **Server Policy** is a required TOML file (`--policy`) and is the sole source of token/session lifetimes and rate-limit thresholds and windows—no CLI overrides for those values, no silent defaults. The schema is closed: every required key must be present and unknown keys are rejected; failure reports the first missing or superfluous key. Hardcoded warn/refuse caps and `code_ttl < access_ttl < refresh_ttl` stay in code (ADR-0017, 0047, 0052, 0040, 0059). CLI supplies process topology and mode only (categories; flag spellings are not locked here): listen address, Issuer URL, signing PEM paths, whether to trust forwarded headers, rate-limit store (`postgres` or `memory`, required, no default), and production-risk allowances (ADR-0018, 0067). Secrets stay in the environment (Postgres credentials; proxy shared secret when trusting forwarded headers). Enabling trust-forwarded without that secret refuses start. Missing required environment values are reported by the driven adapter that needs them (Postgres today; same rule for future adapters). Admin subcommands do not take Server Policy. Chosen over all-argv or env-for-all so bulk policy stays reviewable without putting secrets on the command line or weakening fail-closed intent.

Example shape (normative keys; values illustrative):

```toml
[ttl]
authorization_code    = "1m"
access_token          = "5m"
refresh_token         = "336h"
authorization_session = "15m"

[rate_limit.global]
per_ip_limit  = 120
per_ip_window = "1m"

[rate_limit.login]
per_ip_limit        = 10
per_ip_window       = "1m"
per_username_limit  = 5
per_username_window = "1m"

[rate_limit.authorization]
per_ip_limit         = 30
per_ip_window        = "1m"
per_client_id_limit  = 15
per_client_id_window = "1m"

[rate_limit.token]
per_ip_limit         = 60
per_ip_window        = "1m"
per_client_id_limit  = 30
per_client_id_window = "1m"
```
