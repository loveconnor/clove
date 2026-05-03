import { type ApiClient, request } from "./core"

export type RepositoryVisibility = "public" | "private" | "internal"

export type RepositoryOwnerType = "user" | "organization"

export type Repository = {
  id: string
  owner_type: RepositoryOwnerType
  owner_id: string
  owner: string
  name: string
  description?: string
  visibility: RepositoryVisibility
  default_branch: string
  git_path: string
  is_archived?: boolean
  is_fork?: boolean
  parent_id?: string
  size_bytes?: number
  pushed_at?: string
  created_at: string
  updated_at: string
}

export type Branch = {
  name: string
  commit_sha: string
  protected: boolean
}

export type Tag = {
  name: string
  commit_sha: string
}

export type Commit = {
  sha: string
  message: string
  author_name: string
  author_email: string
  author_date: string
  committer_name: string
  committer_email: string
  committer_date: string
  parents: string[]
}

export type ListRepositoriesParams = {
  owner?: string
  visibility?: RepositoryVisibility
  query?: string
  limit?: number
  cursor?: string
}

export type ListRepositoriesResponse = {
  repositories: Repository[]
  next_cursor?: string
}

export type CreateRepositoryRequest = {
  owner: string
  name: string
  description?: string
  visibility?: RepositoryVisibility
  default_branch?: string
  initialize?: boolean
}

export type UpdateRepositoryRequest = {
  description?: string | null
  visibility?: RepositoryVisibility
  default_branch?: string
  is_archived?: boolean
}

export type ListBranchesResponse = {
  branches: Branch[]
}

export type ListTagsResponse = {
  tags: Tag[]
}

export type ListCommitsParams = {
  ref?: string
  path?: string
  limit?: number
  cursor?: string
}

export type ListCommitsResponse = {
  commits: Commit[]
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

function repoPath(owner: string, repo: string, suffix = "") {
  return `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}${suffix}`
}

export function listRepositories(
  client: ApiClient,
  params: ListRepositoriesParams = {},
  signal?: AbortSignal
) {
  return request<ListRepositoriesResponse>(client, {
    path: `/api/repositories${buildQuery(params)}`,
    signal,
  })
}

export function getRepository(
  client: ApiClient,
  owner: string,
  repo: string,
  signal?: AbortSignal
) {
  return request<{ repository: Repository }>(client, {
    path: repoPath(owner, repo),
    signal,
  }).then((r) => r.repository)
}

export function createRepository(
  client: ApiClient,
  body: CreateRepositoryRequest,
  signal?: AbortSignal
) {
  return request<{ repository: Repository }>(client, {
    method: "POST",
    path: "/api/repositories",
    body,
    signal,
  }).then((r) => r.repository)
}

export function updateRepository(
  client: ApiClient,
  owner: string,
  repo: string,
  body: UpdateRepositoryRequest,
  signal?: AbortSignal
) {
  return request<{ repository: Repository }>(client, {
    method: "PATCH",
    path: repoPath(owner, repo),
    body,
    signal,
  }).then((r) => r.repository)
}

export function deleteRepository(
  client: ApiClient,
  owner: string,
  repo: string,
  signal?: AbortSignal
) {
  return request<void>(client, {
    method: "DELETE",
    path: repoPath(owner, repo),
    signal,
  })
}

export function listBranches(
  client: ApiClient,
  owner: string,
  repo: string,
  signal?: AbortSignal
) {
  return request<ListBranchesResponse>(client, {
    path: repoPath(owner, repo, "/branches"),
    signal,
  }).then((r) => r.branches)
}

export function listTags(
  client: ApiClient,
  owner: string,
  repo: string,
  signal?: AbortSignal
) {
  return request<ListTagsResponse>(client, {
    path: repoPath(owner, repo, "/tags"),
    signal,
  }).then((r) => r.tags)
}

export function listCommits(
  client: ApiClient,
  owner: string,
  repo: string,
  params: ListCommitsParams = {},
  signal?: AbortSignal
) {
  return request<ListCommitsResponse>(client, {
    path: repoPath(owner, repo, `/commits${buildQuery(params)}`),
    signal,
  })
}

export function getCommit(
  client: ApiClient,
  owner: string,
  repo: string,
  sha: string,
  signal?: AbortSignal
) {
  return request<{ commit: Commit }>(client, {
    path: repoPath(owner, repo, `/commits/${encodeURIComponent(sha)}`),
    signal,
  }).then((r) => r.commit)
}
