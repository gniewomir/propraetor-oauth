# Admin CLI uses domain; no running server required

Admin subcommands do not require `cli server` to be running. They use the same Postgres credentials from the environment and go **CLI adapter → domain → persistence**, enforcing the same invariants as the HTTP path. “Direct Postgres” means the admin process opens the database itself—not that it bypasses the domain with ad-hoc SQL.
