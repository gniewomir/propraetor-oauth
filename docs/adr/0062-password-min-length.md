# End-User password minimum length 12

Every Operator CLI path that sets an End-User password (create, set-password, and any future mutation) rejects passwords shorter than 12 characters. No complexity rules, breach checks, or rotation policy in v1. End-User self-service password change remains out of scope. First-party, Operator-provisioned Users: block trivial passwords without inventing a full policy engine.
