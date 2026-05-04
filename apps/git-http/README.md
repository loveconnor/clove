# Clove Git HTTP

Smart HTTP service for Git clone, fetch, and push.

## Routes

- `GET /:owner/:repo.git/info/refs?service=git-upload-pack`
- `GET /:owner/:repo.git/info/refs?service=git-receive-pack`
- `POST /:owner/:repo.git/git-upload-pack`
- `POST /:owner/:repo.git/git-receive-pack`

## Authentication

Public repositories can be cloned and fetched anonymously. Private repositories
and all pushes require HTTP Basic credentials:

- Username: the user's Clove username
- Password: a personal access token

Personal access tokens are managed through the API service:

- `GET /api/personal-access-tokens`
- `POST /api/personal-access-tokens`
- `DELETE /api/personal-access-tokens/{id}`

Push authorization is restricted to personal repository owners and organization
owners/admins.

## Run

```sh
cd apps/git-http
go run ./cmd/git-http
```

The service defaults to `GIT_HTTP_PORT=8081` and reads the shared
`DATABASE_URL` and `REPOSITORY_ROOT` environment variables.
