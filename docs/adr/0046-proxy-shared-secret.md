# Proxy authenticity via shared secret only

When the AS honors proxy forwarding headers (e.g. X-Forwarded-Proto), authenticity is established solely by a shared proxy secret (e.g. a required header value known only to the proxy and the AS). That secret is mandatory for treating forwarded proto as trusted. IP/CIDR/hostname allowlists and “network placement alone” are not used as authenticity mechanisms. Spoofing risk if the secret leaks or any peer that can reach the AS also knows the secret remains an Operator control problem for secret distribution and listener exposure.
