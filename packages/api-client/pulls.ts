import { type ApiClient, request } from "./core"
import type { User } from "./users"

export type PullRequestState = "open" | "closed" | "merged" | "draft"

export type MergeMethod = "merge" | "squash" | "rebase"

export type ReviewState =
  | "pending"
  | "approved"
  | "changes_requested"
  | "commented"
  | "dismissed"

export type GitRef = {
  label: string
  ref: string
  sha: string
  repository_id: string
}

export type PullRequest = {
  id: string
  repository_id: string
  number: number
  title: string
  body?: string
  state: PullRequestState
  draft: boolean
  author: User
  base: GitRef
  head: GitRef
  merge_commit_sha?: string
  merged_by?: User
  merged_at?: string
  closed_at?: string
  requested_reviewers?: User[]
  requested_teams?: string[]
  labels?: string[]
  comments_count: number
  review_comments_count: number
  commits_count: number
  additions: number
  deletions: number
  changed_files: number
  created_at: string
  updated_at: string
}

export type Review = {
  id: string
  pull_request_id: string
  author: User
  state: ReviewState
  body?: string
  commit_sha?: string
  submitted_at?: string
  created_at: string
  updated_at: string
}

export type ReviewComment = {
  id: string
  pull_request_id: string
  review_id?: string
  author: User
  body: string
  path: string
  line: number
  side: "left" | "right"
  commit_sha: string
  in_reply_to_id?: string
  resolved: boolean
  created_at: string
  updated_at: string
}

export type PullRequestComment = {
  id: string
  pull_request_id: string
  author: User
  body: string
  created_at: string
  updated_at: string
}

export type ListPullsParams = {
  state?: PullRequestState | "all"
  author?: string
  reviewer?: string
  base?: string
  head?: string
  label?: string
  query?: string
  sort?: "created" | "updated" | "popularity"
  direction?: "asc" | "desc"
  limit?: number
  cursor?: string
}

export type ListPullsResponse = {
  pull_requests: PullRequest[]
  next_cursor?: string
}

export type CreatePullRequest = {
  title: string
  body?: string
  base: string
  head: string
  draft?: boolean
  reviewers?: string[]
  labels?: string[]
}

export type UpdatePullRequest = {
  title?: string
  body?: string | null
  state?: Exclude<PullRequestState, "merged">
  base?: string
  draft?: boolean
}

export type MergePullRequest = {
  method?: MergeMethod
  commit_title?: string
  commit_message?: string
  expected_head_sha?: string
}

export type MergePullResponse = {
  merged: boolean
  merge_commit_sha?: string
  message?: string
}

export type ListReviewsResponse = {
  reviews: Review[]
}

export type CreateReviewRequest = {
  commit_sha?: string
  body?: string
  event: "approve" | "request_changes" | "comment" | "pending"
  comments?: Array<{
    path: string
    line: number
    side?: "left" | "right"
    body: string
    in_reply_to_id?: string
  }>
}

export type ListReviewCommentsResponse = {
  comments: ReviewComment[]
}

export type CreateReviewCommentRequest = {
  body: string
  path: string
  line: number
  side?: "left" | "right"
  commit_sha: string
  in_reply_to_id?: string
}

export type ListPullCommentsResponse = {
  comments: PullRequestComment[]
}

export type CreatePullCommentRequest = {
  body: string
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

function pullsPath(owner: string, repo: string, suffix = "") {
  return `/api/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/pulls${suffix}`
}

export function listPullRequests(
  client: ApiClient,
  owner: string,
  repo: string,
  params: ListPullsParams = {},
  signal?: AbortSignal
) {
  return request<ListPullsResponse>(client, {
    path: pullsPath(owner, repo, buildQuery(params)),
    signal,
  })
}

export function getPullRequest(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  signal?: AbortSignal
) {
  return request<{ pull_request: PullRequest }>(client, {
    path: pullsPath(owner, repo, `/${number}`),
    signal,
  }).then((r) => r.pull_request)
}

export function createPullRequest(
  client: ApiClient,
  owner: string,
  repo: string,
  body: CreatePullRequest,
  signal?: AbortSignal
) {
  return request<{ pull_request: PullRequest }>(client, {
    method: "POST",
    path: pullsPath(owner, repo),
    body,
    signal,
  }).then((r) => r.pull_request)
}

export function updatePullRequest(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: UpdatePullRequest,
  signal?: AbortSignal
) {
  return request<{ pull_request: PullRequest }>(client, {
    method: "PATCH",
    path: pullsPath(owner, repo, `/${number}`),
    body,
    signal,
  }).then((r) => r.pull_request)
}

export function mergePullRequest(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: MergePullRequest = {},
  signal?: AbortSignal
) {
  return request<MergePullResponse>(client, {
    method: "POST",
    path: pullsPath(owner, repo, `/${number}/merge`),
    body,
    signal,
  })
}

export function listReviews(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  signal?: AbortSignal
) {
  return request<ListReviewsResponse>(client, {
    path: pullsPath(owner, repo, `/${number}/reviews`),
    signal,
  }).then((r) => r.reviews)
}

export function createReview(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: CreateReviewRequest,
  signal?: AbortSignal
) {
  return request<{ review: Review }>(client, {
    method: "POST",
    path: pullsPath(owner, repo, `/${number}/reviews`),
    body,
    signal,
  }).then((r) => r.review)
}

export function listReviewComments(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  signal?: AbortSignal
) {
  return request<ListReviewCommentsResponse>(client, {
    path: pullsPath(owner, repo, `/${number}/review-comments`),
    signal,
  }).then((r) => r.comments)
}

export function createReviewComment(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: CreateReviewCommentRequest,
  signal?: AbortSignal
) {
  return request<{ comment: ReviewComment }>(client, {
    method: "POST",
    path: pullsPath(owner, repo, `/${number}/review-comments`),
    body,
    signal,
  }).then((r) => r.comment)
}

export function resolveReviewComment(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  commentId: string,
  resolved: boolean,
  signal?: AbortSignal
) {
  return request<{ comment: ReviewComment }>(client, {
    method: "PATCH",
    path: pullsPath(
      owner,
      repo,
      `/${number}/review-comments/${encodeURIComponent(commentId)}`
    ),
    body: { resolved },
    signal,
  }).then((r) => r.comment)
}

export function listPullComments(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  signal?: AbortSignal
) {
  return request<ListPullCommentsResponse>(client, {
    path: pullsPath(owner, repo, `/${number}/comments`),
    signal,
  }).then((r) => r.comments)
}

export function createPullComment(
  client: ApiClient,
  owner: string,
  repo: string,
  number: number,
  body: CreatePullCommentRequest,
  signal?: AbortSignal
) {
  return request<{ comment: PullRequestComment }>(client, {
    method: "POST",
    path: pullsPath(owner, repo, `/${number}/comments`),
    body,
    signal,
  }).then((r) => r.comment)
}
