# One Audience per Client

Each Client is registered with exactly one Audience (JWT `aud` for the intended Resource Server / Protected Resource identifier). Every Access Token issued to that Client carries that value. Per-request resource indicators and multi-audience Clients are out of v1.
