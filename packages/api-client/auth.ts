import { type ApiClient, request } from "./core"
import type { User } from "./users"

export type Session = {
  id: string
  user_id: string
  created_at: string
  expires_at: string
  ip_address?: string
  user_agent?: string
  current?: boolean
}

export type LoginRequest = {
  identifier: string
  password: string
  remember_me?: boolean
}

export type LoginResponse = {
  user: User
  session: Session
}

export type RegisterRequest = {
  username: string
  email: string
  password: string
  display_name?: string
}

export type RegisterResponse = {
  user: User
  session: Session
}

export type RefreshResponse = {
  session: Session
}

export type ListSessionsResponse = {
  sessions: Session[]
}

export function login(client: ApiClient, body: LoginRequest, signal?: AbortSignal) {
  return request<LoginResponse>(client, {
    method: "POST",
    path: "/api/auth/login",
    body,
    signal,
  })
}

export function register(
  client: ApiClient,
  body: RegisterRequest,
  signal?: AbortSignal
) {
  return request<RegisterResponse>(client, {
    method: "POST",
    path: "/api/auth/register",
    body,
    signal,
  })
}

export function logout(client: ApiClient, signal?: AbortSignal) {
  return request<void>(client, {
    method: "POST",
    path: "/api/auth/logout",
    signal,
  })
}

export function refresh(client: ApiClient, signal?: AbortSignal) {
  return request<RefreshResponse>(client, {
    method: "POST",
    path: "/api/auth/refresh",
    signal,
  })
}

export function listSessions(client: ApiClient, signal?: AbortSignal) {
  return request<ListSessionsResponse>(client, {
    path: "/api/auth/sessions",
    signal,
  }).then((r) => r.sessions)
}

export function revokeSession(
  client: ApiClient,
  sessionId: string,
  signal?: AbortSignal
) {
  return request<void>(client, {
    method: "DELETE",
    path: `/api/auth/sessions/${encodeURIComponent(sessionId)}`,
    signal,
  })
}
