# Fixed Argon2id parameters (OWASP-aligned)

End-User passwords and Confidential Client secrets use fixed Argon2id parameters in code (not Operator-tunable): memory 64 MiB, time (iterations) 3, parallelism 1, salt length 16 bytes, hash length 32 bytes. Chosen to align with the OWASP Password Storage Cheat Sheet Argon2id guidance for interactive verification. Same parameters for both secret types so one verifier path; weaker Operator-tuned configs are rejected as a v1 footgun.
