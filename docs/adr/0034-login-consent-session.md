# Server-side Authorization Session across login and Consent

The Authorization Code login and Consent steps are tied by an Authorization Session (opaque session id in a cookie; session data server-side). Authorization Request continuity is not done solely via hidden form round-trips. Cookie attributes and hardening are ADR-0048 and ADR-0064 (HttpOnly; Secure / `__Host-` when Issuer URL is https; SameSite=Lax; CSRF; fixation rotation; bound request fields).
