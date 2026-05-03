# Clove API

Phase 1 backend foundation for the Clove application.

## Run locally

```bash
go run ./cmd/api
```

The server listens on `0.0.0.0:8080` by default.

## Configuration

| Environment variable | Default | Description |
| --- | --- | --- |
| `APP_NAME` | `clove-api` | Service name in logs and health responses. |
| `APP_ENV` | `development` | Runtime environment. Development uses text logs; other envs use JSON logs. |
| `HOST` | `0.0.0.0` | HTTP bind host. |
| `PORT` | `8080` | HTTP bind port. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error`. |
| `DATABASE_URL` | empty | Postgres connection string. When empty, DB and migrations are skipped. |
| `DATABASE_DRIVER` | `pgx` | SQL driver name. |
| `WORKOS_API_KEY` | empty | WorkOS API key. Required for auth endpoints. |
| `WORKOS_CLIENT_ID` | empty | WorkOS AuthKit client ID. Required for auth endpoints and JWT validation. |
| `WORKOS_REDIRECT_URI` | `http://localhost:8080/api/auth/callback` | WorkOS AuthKit callback URL. |
| `WORKOS_ISSUER` | `https://api.workos.com/` | Expected access-token issuer. Use your custom auth domain if configured. |
| `AUTH_SUCCESS_REDIRECT_URL` | `/` | Browser redirect after hosted AuthKit callback succeeds. |
| `AUTH_LOGOUT_REDIRECT_URL` | `/` | Browser redirect after logout. |
| `AUTH_COOKIE_SECURE` | `false` in development | Whether auth cookies require HTTPS. |

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Service and database health. |
| `GET` | `/api/auth/register` | Redirect to hosted WorkOS AuthKit sign-up. |
| `POST` | `/api/auth/register` | Create a WorkOS email/password user and set session cookies. |
| `GET` | `/api/auth/login` | Redirect to hosted WorkOS AuthKit sign-in. |
| `POST` | `/api/auth/login` | Authenticate with WorkOS email/password and set session cookies. |
| `GET` | `/api/auth/callback` | Complete hosted AuthKit code exchange and redirect. |
| `POST` | `/api/auth/callback` | Complete code exchange for API clients. |
| `GET` | `/api/auth/logout` | Clear local cookies and redirect. |
| `POST` | `/api/auth/logout` | Clear local cookies and revoke the WorkOS session when possible. |
| `GET` | `/api/me` | Protected current-user endpoint backed by WorkOS session middleware. |
| `GET` | `/api/organizations` | Protected list of organizations for the current user. |
| `GET` | `/api/organizations/{owner}` | Protected organization lookup by slug/name. |
| `GET` | `/api/repositories` | Protected repository list visible to the current user. |
| `POST` | `/api/repositories` | Protected repository creation for the current user or an accessible organization. |
| `GET` | `/api/repositories/{owner}/{repo}` | Protected repository lookup. |

Auth uses HTTP-only cookies for the WorkOS access and refresh tokens. Protected
routes validate access tokens against the WorkOS JWKS and refresh sessions when
an access token has expired but a refresh token is still present.

For local web development, set:

```bash
WORKOS_REDIRECT_URI=http://localhost:8080/api/auth/callback
AUTH_SUCCESS_REDIRECT_URL=http://localhost:3001/dashboard
AUTH_LOGOUT_REDIRECT_URL=http://localhost:3001/login
```

## Authorization

Repository authorization lives in `internal/permissions` and exposes shared
permission checks for repository-dependent workflows:

- `CanViewRepo`
- `CanPushRepo`
- `CanCreatePullRequest`
- `CanReviewPullRequest`
- `CanMergePullRequest`
- `CanAdminRepo`

Roles are ordered as `owner`, `admin`, `maintainer`, `write`, `triage`, `read`.
Public repositories are viewable by anyone; internal repositories require an
authenticated user; private repositories require ownership, organization, or
repository-level access.

## Migrations

Migrations are embedded from `internal/db/migrations` and run at API startup when
`DATABASE_URL` is configured. The initial schema creates:

- `users`
- `sessions`
- `organizations`
- `organization_members`
- `repositories`
- `ssh_keys`
- `audit_logs`
