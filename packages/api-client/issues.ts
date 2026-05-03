import { type ApiClient, request } from "./core"
import type { User } from "./users"

export type IssueState = "open" | "closed"

export type IssueStateReason = "completed" | "not_planned" | "reopened"

export type Label = {
  id: string
  name: string
  color: string
  description?: string
  created_at: string
  updated_at: string
}

export type Milestone = {
  id: string
  number: number
  title: string
  description?: string
  state: "open" | "closed"
  due_on?: string
  created_at: string
  updated_at: string
  closed_at?: string
}

export type Issue = {
  id: string
  repository_id: string
  number: number
  title: string
  body?: string
  state: IssueState
  state_reason?: IssueStateReason
  author: User
  assignees: User[]
  labels: Label[]
  milestone?: Milestone
  comments_count: number
  closed_at?: string
  closed_by?: User
  created_at: string
  updated_at: string
}

export type IssueComment = {
  id: string
  issue_id: string
  author: User
  body: string
  created_at: string
  updated_at: string
}

export type ListIssuesParams = {
  state?: IssueState | "all"
  author?: string
  assignee?: string
  label?: string
  milestone?: string
  query?: string
  sort?: "created" | "updated" | "comments"
  direction?: "asc" | "desc"
  limit?: number
  cursor?: string
}

export type ListIssuesResponse = {
  issues: Issue[]
  next_cursor?: string
}

export type CreateIssueRequest = {
  title: string
  body?: string
  assignees?: string[]
  labels?: string[]
  milestone?: string
}

export type UpdateIssueRequest = {
  title?: string
  body?: string | null
  state?: IssueState
  state_reason?: IssueStateReason
  assignees?: string[]
  labels?: string[]
  milestone?: string | null
}

export type ListIssueCommentsResponse = {
  comments: IssueComment[]
  next_cursor?: string
}

export type CreateIssueCommentRequest = {
  body: string
}

export type UpdateIssueCommentRequest = {
  body: string
}

export type ListLabelsResponse = {
  labels: Label[]
}

export type CreateLabelRequest = {
  name: string
  color: string
  description?: string
}

export type UpdateLabelRequest = {
  name?: string
  color?: string
  description?: string | null
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

function issuesPath(owner: string, repo: string, suffix = "") {
  return `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues${suffix}`
}

function labelsPath(owner: string, repo: string, suffix = "") {
  return `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/labels${suffix}`
}

export function listIssues(
  client: ApiClient,
  owner: string,
  repo: string,
  params: ListIssuesParams = {},
  signal?: AbortSignal
) {
  return request<ListIssuesResponse>(client, {
    path: issuesPath(owner, repo, buildQuery(params)),
    signal,
  })
}

export function getIssue(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  signal?: AbortSignal
) {
  return request<{ issue: Issue }>(client, {
    path: issuesPath(owner, repo, `/${number}`),
    signal,
  }).then((r) => r.issue)
}

export function createIssue(
  client: ApiClient,
  owner: string,
  repo: string,
  body: CreateIssueRequest,
  signal?: AbortSignal
) {
  return request<{ issue: Issue }>(client, {
    method: "POST",
    path: issuesPath(owner, repo),
    body,
    signal,
  }).then((r) => r.issue)
}

export function updateIssue(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: UpdateIssueRequest,
  signal?: AbortSignal
) {
  return request<{ issue: Issue }>(client, {
    method: "PATCH",
    path: issuesPath(owner, repo, `/${number}`),
    body,
    signal,
  }).then((r) => r.issue)
}

export function listIssueComments(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  signal?: AbortSignal
) {
  return request<ListIssueCommentsResponse>(client, {
    path: issuesPath(owner, repo, `/${number}/comments`),
    signal,
  })
}

export function createIssueComment(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: CreateIssueCommentRequest,
  signal?: AbortSignal
) {
  return request<{ comment: IssueComment }>(client, {
    method: "POST",
    path: issuesPath(owner, repo, `/${number}/comments`),
    body,
    signal,
  }).then((r) => r.comment)
}

export function updateIssueComment(
  client: ApiClient,
  owner: string,
  repo: string,
  commentId: string,
  body: UpdateIssueCommentRequest,
  signal?: AbortSignal
) {
  return request<{ comment: IssueComment }>(client, {
    method: "PATCH",
    path: issuesPath(owner, repo, `/comments/${encodeURIComponent(commentId)}`),
    body,
    signal,
  }).then((r) => r.comment)
}

export function deleteIssueComment(
  client: ApiClient,
  owner: string,
  repo: string,
  commentId: string,
  signal?: AbortSignal
) {
  return request<void>(client, {
    method: "DELETE",
    path: issuesPath(owner, repo, `/comments/${encodeURIComponent(commentId)}`),
    signal,
  })
}

export function listLabels(
  client: ApiClient,
  owner: string,
  repo: string,
  signal?: AbortSignal
) {
  return request<ListLabelsResponse>(client, {
    path: labelsPath(owner, repo),
    signal,
  }).then((r) => r.labels)
}

export function createLabel(
  client: ApiClient,
  owner: string,
  repo: string,
  body: CreateLabelRequest,
  signal?: AbortSignal
) {
  return request<{ label: Label }>(client, {
    method: "POST",
    path: labelsPath(owner, repo),
    body,
    signal,
  }).then((r) => r.label)
}

export function updateLabel(
  client: ApiClient,
  owner: string,
  repo: string,
  name: string,
  body: UpdateLabelRequest,
  signal?: AbortSignal
) {
  return request<{ label: Label }>(client, {
    method: "PATCH",
    path: labelsPath(owner, repo, `/${encodeURIComponent(name)}`),
    body,
    signal,
  }).then((r) => r.label)
}

export function deleteLabel(
  client: ApiClient,
  owner: string,
  repo: string,
  name: string,
  signal?: AbortSignal
) {
  return request<void>(client, {
    method: "DELETE",
    path: labelsPath(owner, repo, `/${encodeURIComponent(name)}`),
    signal,
  })
}
