CREATE TABLE clients (
    client_id      text        PRIMARY KEY,
    auth_method    text        NOT NULL,
    secret_hash    text        NULL,
    audience       text        NOT NULL,
    deactivated_at timestamptz NULL,
    not_before     timestamptz NULL
);

CREATE UNIQUE INDEX clients_client_id_lower_uidx
    ON clients (lower(client_id));
