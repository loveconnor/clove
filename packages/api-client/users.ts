import { type ApiClient, request } from "./core"

export type User = {
  id: string
  username: string
  email: string
  display_name?: string
  avatar_url?: string
  bio?: string
  location?: string
  website?: string
  created_at: string
  updated_at: string
}

export type Viewer = {
  user: User
  session_id: string
  organization_id?: string
  role?: string
  permissions?: string[]
}

export type UpdateUserRequest = {
  display_name?: string | null
  bio?: string | null
  location?: string | null
  website?: string | null
  avatar_url?: string | null
}

export type ListUsersParams = {
  query?: string
  limit?: number
  cursor?: string
}

export type ListUsersResponse = {
  users: User[]
  next_cursor?: string
}

function buildQuery(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null) continue
    search.set(key, String(value))
  }
  const qs = search.toString()
  return qs ? `?${qs}` : ""
}

export function getViewer(client: ApiClient, signal?: AbortSignal) {
  return request<Viewer>(client, { path: "/api/me", signal })
}

export function getUser(client: ApiClient, username: string, signal?: AbortSignal) {
  return request<{ user: User }>(client, {
    path: `/api/users/${encodeURIComponent(username)}`,
    signal,
  }).then((r) => r.user)
}

export function listUsers(
  client: ApiClient,
  params: ListUsersParams = {},
  signal?: AbortSignal
) {
  return request<ListUsersResponse>(client, {
    path: `/api/users${buildQuery(params)}`,
    signal,
  })
}

export function updateViewer(
  client: ApiClient,
  body: UpdateUserRequest,
  signal?: AbortSignal
) {
  return request<{ user: User }>(client, {
    method: "PATCH",
    path: "/api/me",
    body,
    signal,
  }).then((r) => r.user)
}
