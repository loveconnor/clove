import { cookies } from "next/headers"

export type User = {
  id: string
  username: string
  email: string
  display_name?: string
  avatar_url?: string
}

export type Viewer = {
  user: User
  session_id: string
  organization_id?: string
  role?: string
  permissions?: string[]
}

export type Organization = {
  id: string
  name: string
  display_name?: string
  description?: string
  owner_id: string
  role?: "owner" | "admin" | "member" | string
  created_at: string
  updated_at: string
}

export type OrganizationMember = {
  user_id: string
  username: string
  email: string
  display_name?: string
  avatar_url?: string
  role: "owner" | "admin" | "member"
  joined_at: string
}

export type Repository = {
  id: string
  owner_type: string
  owner_id: string
  owner: string
  name: string
  description?: string
  visibility: "public" | "private" | "internal"
  default_branch: string
  git_path: string
  created_at: string
  updated_at: string
}

export type RepositoryTreeEntry = {
  name: string
  path: string
  type: "blob" | "tree" | string
  mode: string
  sha: string
}

export type RepositoryTree = {
  repository: Repository
  ref: string
  commit_sha: string
  path: string
  entries: RepositoryTreeEntry[]
}

export type RepositoryBlob = {
  path: string
  sha: string
  size: number
  content: string
  encoding: string
}

export type RepositoryBlobResponse = {
  repository: Repository
  ref: string
  commit_sha: string
  blob: RepositoryBlob
}

type APIError = {
  error?: {
    code?: string
    message?: string
  }
}

export class APIRequestError extends Error {
  status: number
  code?: string

  constructor(status: number, message: string, code?: string) {
    super(message)
    this.name = "APIRequestError"
    this.status = status
    this.code = code
  }
}

const apiBaseURL = process.env.API_INTERNAL_URL ?? "http://localhost:8080"

export async function getViewer() {
  return apiFetch<Viewer>("/api/me")
}

export async function getOrganizations() {
  const data = await apiFetch<{ organizations: Organization[] }>(
    "/api/organizations"
  )
  return data.organizations
}

export async function getOrganization(owner: string) {
  const data = await apiFetch<{ organization: Organization }>(
    `/api/organizations/${encodeURIComponent(owner)}`
  )
  return data.organization
}

export async function getOrganizationMembers(owner: string) {
  const data = await apiFetch<{ members: OrganizationMember[] }>(
    `/api/organizations/${encodeURIComponent(owner)}/members`
  )
  return data.members
}

export async function getRepositories(owner?: string) {
  const query = owner ? `?owner=${encodeURIComponent(owner)}` : ""
  const data = await apiFetch<{ repositories: Repository[] }>(
    `/api/repositories${query}`
  )
  return data.repositories
}

export async function getRepository(owner: string, repo: string) {
  const data = await apiFetch<{ repository: Repository }>(
    `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`
  )
  return data.repository
}

export async function getRepositoryTree(
  owner: string,
  repo: string,
  ref?: string,
  path?: string
) {
  const query = new URLSearchParams()
  if (ref) query.set("ref", ref)
  if (path) query.set("path", path)
  const suffix = query.toString() ? `?${query.toString()}` : ""
  return apiFetch<RepositoryTree>(
    `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/tree${suffix}`
  )
}

export async function getRepositoryBlob(
  owner: string,
  repo: string,
  path: string,
  ref?: string
) {
  const query = new URLSearchParams({ path })
  if (ref) query.set("ref", ref)
  return apiFetch<RepositoryBlobResponse>(
    `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/blob?${query.toString()}`
  )
}

async function apiFetch<T>(path: string): Promise<T> {
  const cookieStore = await cookies()
  const response = await fetch(`${apiBaseURL}${path}`, {
    headers: {
      Cookie: cookieStore.toString(),
    },
    cache: "no-store",
  })

  if (!response.ok) {
    let payload: APIError = {}
    try {
      payload = (await response.json()) as APIError
    } catch {
      payload = {}
    }
    throw new APIRequestError(
      response.status,
      payload.error?.message ?? "API request failed",
      payload.error?.code
    )
  }

  return (await response.json()) as T
}
