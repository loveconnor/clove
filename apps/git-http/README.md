# Clove Git HTTP

Smart HTTP service for Git clone, fetch, and push.

## Routes

- `GET /:owner/:repo.git/info/refs?service=git-upload-pack`
- `GET /:owner/:repo.git/info/refs?service=git-receive-pack`
- `POST /:owner/:repo.git/git-upload-pack`
- `POST /:owner/:repo.git/git-receive-pack`

## Authentication

Public repositories can be cloned and fetched anonymously. Private repositories
and all pushes require a valid WorkOS access token supplied as either:

- `Authorization: Bearer <access-token>`
- HTTP Basic auth, where the password is the access token
- the configured access-token cookie

Push authorization is restricted to personal repository owners and organization
owners/admins.

## Run

```sh
cd apps/git-http
go run ./cmd/git-http
```

The service defaults to `GIT_HTTP_PORT=8081` and reads the shared
`DATABASE_URL`, `REPOSITORY_ROOT`, `WORKOS_CLIENT_ID`, and `WORKOS_ISSUER`
environment variables.
