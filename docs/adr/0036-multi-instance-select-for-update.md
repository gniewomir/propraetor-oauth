# Multi-instance safety via SELECT FOR UPDATE

Multiple `cli server` processes may run against one Postgres. One-time Authorization Code redemption at the Token Endpoint and Refresh Token rotation/reuse detection run inside database transactions that lock the relevant row with `SELECT … FOR UPDATE` before validating and consuming it. In-process-only mutual exclusion is not relied on for these semantics.
