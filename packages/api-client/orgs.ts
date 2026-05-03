import { type ApiClient, request } from "./core"
import type { User } from "./users"

export type OrganizationVisibility = "public" | "private"

export type OrganizationRole = "owner" | "admin" | "member" | "guest"

export type Organization = {
  id: string
  name: string
  display_name?: string
  description?: string
  owner_id: string
  visibility: OrganizationVisibility
  avatar_url?: string
  role?: OrganizationRole
  created_at: string
  updated_at: string
}

export type OrganizationMember = {
  user: User
  role: OrganizationRole
  joined_at: string
}

export type ListOrganizationsParams = {
  query?: string
  limit?: number
  cursor?: string
}

export type ListOrganizationsResponse = {
  organizations: Organization[]
  next_cursor?: string
}

export type CreateOrganizationRequest = {
  name: string
  display_name?: string
  description?: string
  visibility?: OrganizationVisibility
}

export type UpdateOrganizationRequest = {
  display_name?: string | null
  description?: string | null
  visibility?: OrganizationVisibility
  avatar_url?: string | null
}

export type ListMembersParams = {
  role?: OrganizationRole
  query?: string
  limit?: number
  cursor?: string
}

export type ListMembersResponse = {
  members: OrganizationMember[]
  next_cursor?: string
}

export type AddMemberRequest = {
  username: string
  role: OrganizationRole
}

export type UpdateMemberRequest = {
  role: OrganizationRole
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

export function listOrganizations(
  client: ApiClient,
  params: ListOrganizationsParams = {},
  signal?: AbortSignal
) {
  return request<ListOrganizationsResponse>(client, {
    path: `/api/organizations${buildQuery(params)}`,
    signal,
  })
}

export function getOrganization(
  client: ApiClient,
  name: string,
  signal?: AbortSignal
) {
  return request<{ organization: Organization }>(client, {
    path: `/api/organizations/${encodeURIComponent(name)}`,
    signal,
  }).then((r) => r.organization)
}

export function createOrganization(
  client: ApiClient,
  body: CreateOrganizationRequest,
  signal?: AbortSignal
) {
  return request<{ organization: Organization }>(client, {
    method: "POST",
    path: "/api/organizations",
    body,
    signal,
  }).then((r) => r.organization)
}

export function updateOrganization(
  client: ApiClient,
  name: string,
  body: UpdateOrganizationRequest,
  signal?: AbortSignal
) {
  return request<{ organization: Organization }>(client, {
    method: "PATCH",
    path: `/api/organizations/${encodeURIComponent(name)}`,
    body,
    signal,
  }).then((r) => r.organization)
}

export function deleteOrganization(
  client: ApiClient,
  name: string,
  signal?: AbortSignal
) {
  return request<void>(client, {
    method: "DELETE",
    path: `/api/organizations/${encodeURIComponent(name)}`,
    signal,
  })
}

export function listMembers(
  client: ApiClient,
  org: string,
  params: ListMembersParams = {},
  signal?: AbortSignal
) {
  return request<ListMembersResponse>(client, {
    path: `/api/organizations/${encodeURIComponent(org)}/members${buildQuery(params)}`,
    signal,
  })
}

export function addMember(
  client: ApiClient,
  org: string,
  body: AddMemberRequest,
  signal?: AbortSignal
) {
  return request<{ member: OrganizationMember }>(client, {
    method: "POST",
    path: `/api/organizations/${encodeURIComponent(org)}/members`,
    body,
    signal,
  }).then((r) => r.member)
}

export function updateMember(
  client: ApiClient,
  org: string,
  username: string,
  body: UpdateMemberRequest,
  signal?: AbortSignal
) {
  return request<{ member: OrganizationMember }>(client, {
    method: "PATCH",
    path: `/api/organizations/${encodeURIComponent(org)}/members/${encodeURIComponent(username)}`,
    body,
    signal,
  }).then((r) => r.member)
}

export function removeMember(
  client: ApiClient,
  org: string,
  username: string,
  signal?: AbortSignal
) {
  return request<void>(client, {
    method: "DELETE",
    path: `/api/organizations/${encodeURIComponent(org)}/members/${encodeURIComponent(username)}`,
    signal,
  })
}
