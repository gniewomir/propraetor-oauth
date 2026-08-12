# Forwarding headers require shared proxy secret

Proxy forwarding headers are untrusted by default. They are honored only when accompanied by the configured shared proxy secret (ADR-0046). There is no IP/CIDR/hostname allowlist. Operator network isolation remains advisable but is not treated as proof of authenticity.
