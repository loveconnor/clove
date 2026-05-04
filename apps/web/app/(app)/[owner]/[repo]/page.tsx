import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import {
  ArrowLeft,
  Code2,
  FileCode2,
  FileText,
  Folder,
  GitBranch,
  GitCommitHorizontal,
  GitPullRequestArrow,
  LockKeyhole,
  Settings,
  ShieldCheck,
} from "lucide-react"

import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@loveui/ui/ui/empty"

import {
  APIRequestError,
  getRepositoryBlob,
  getRepository,
  getRepositoryTree,
  type RepositoryTreeEntry,
} from "@/apps/web/lib/api"

export default async function RepositoryPage({
  params,
}: {
  params: Promise<{ owner: string; repo: string }>
}) {
  const { owner, repo } = await params
  const { repository, tree, readme } = await loadRepositoryData(owner, repo)
  const cloneURL = `https://clove.dev/${repository.owner}/${repository.name}.git`
  const hasFiles = Boolean(tree && tree.entries.length > 0)

  return (
    <div className="space-y-6">
      <div>
        <Button variant="ghost" size="sm" asChild>
          <Link href={`/${owner}`}>
            <ArrowLeft />
            {owner}
          </Link>
        </Button>
      </div>

      <section className="rounded-lg border bg-card p-6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-3">
              <div className="flex size-10 items-center justify-center rounded-lg border bg-background">
                {repository.visibility === "private" ? (
                  <LockKeyhole className="size-5" />
                ) : (
                  <Code2 className="size-5" />
                )}
              </div>
              <div className="min-w-0">
                <p className="text-sm text-muted-foreground">
                  {repository.visibility} repository
                </p>
                <h1 className="truncate text-3xl font-semibold tracking-normal">
                  {repository.owner}/{repository.name}
                </h1>
              </div>
            </div>
            <p className="mt-4 max-w-3xl text-sm leading-6 text-muted-foreground">
              {repository.description}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant="outline">{repository.default_branch}</Badge>
            <Badge variant={hasFiles ? "success" : "outline"}>
              {hasFiles ? "Pushed" : "Empty"}
            </Badge>
            <Button variant="outline" size="sm" asChild>
              <Link href={`/${repository.owner}/${repository.name}/settings`}>
                <Settings />
                Settings
              </Link>
            </Button>
          </div>
        </div>
      </section>

      <section className="grid gap-3 md:grid-cols-4">
        <Metric
          icon={GitPullRequestArrow}
          label="Visibility"
          value={repository.visibility}
        />
        <Metric
          icon={GitBranch}
          label="Default branch"
          value={repository.default_branch}
        />
        <Metric
          icon={ShieldCheck}
          label="Owner type"
          value={repository.owner_type}
        />
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <div className="flex min-w-0 items-center justify-between gap-3">
              <CardTitle>Files</CardTitle>
              {tree ? (
                <div className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                  <GitCommitHorizontal className="size-4 shrink-0" />
                  <span className="truncate font-mono">
                    {tree.commit_sha.slice(0, 12)}
                  </span>
                </div>
              ) : null}
            </div>
          </CardHeader>
          <CardPanel>
            {tree && tree.entries.length > 0 ? (
              <FileList entries={tree.entries} />
            ) : (
              <EmptyRepository cloneURL={cloneURL} branch={repository.default_branch} />
            )}
          </CardPanel>
        </Card>

        <div className="space-y-6">
          <Card variant="outline" className="rounded-lg">
            <CardHeader className="border-b">
              <CardTitle>Repository metadata</CardTitle>
            </CardHeader>
            <CardPanel className="grid gap-3">
              <ReviewBlock title="Clone URL" detail={cloneURL} />
              <ReviewBlock title="Repository ID" detail={repository.id} />
              <ReviewBlock title="Git path" detail={repository.git_path} />
              <ReviewBlock
                title="Created"
                detail={formatDate(repository.created_at)}
              />
              <ReviewBlock
                title="Updated"
                detail={formatDate(repository.updated_at)}
              />
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardHeader className="border-b">
              <CardTitle>Clone</CardTitle>
            </CardHeader>
            <CardPanel>
              <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 font-mono text-xs">
                {cloneURL}
              </code>
            </CardPanel>
          </Card>
        </div>
      </section>

      {readme ? (
        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <div className="flex items-center gap-2">
              <FileText className="size-4 text-muted-foreground" />
              <CardTitle>{readme.path}</CardTitle>
            </div>
          </CardHeader>
          <CardPanel>
            <pre className="max-h-[32rem] overflow-auto rounded-md border bg-background p-4 font-mono text-sm leading-6 whitespace-pre-wrap">
              {readme.content}
            </pre>
          </CardPanel>
        </Card>
      ) : null}
    </div>
  )
}

async function loadRepositoryData(owner: string, repo: string) {
  try {
    const repository = await getRepository(owner, repo)
    const tree = await loadRepositoryTree(owner, repo, repository.default_branch)
    const readmeEntry = tree?.entries.find((entry) =>
      /^readme(?:\.[^.]+)?$/i.test(entry.name)
    )
    const readme =
      readmeEntry?.type === "blob"
        ? await loadRepositoryBlob(owner, repo, readmeEntry.path, repository.default_branch)
        : null
    return { repository, tree, readme }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    if (
      error instanceof APIRequestError &&
      error.status === 503 &&
      error.code === "service_unavailable"
    ) {
      redirect("/dashboard")
    }
    if (error instanceof APIRequestError && error.status === 404) {
      notFound()
    }
    throw error
  }
}

async function loadRepositoryTree(owner: string, repo: string, ref: string) {
  try {
    return await getRepositoryTree(owner, repo, ref)
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 404) {
      return null
    }
    throw error
  }
}

async function loadRepositoryBlob(
  owner: string,
  repo: string,
  path: string,
  ref: string
) {
  try {
    const data = await getRepositoryBlob(owner, repo, path, ref)
    return data.blob
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 404) {
      return null
    }
    throw error
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value))
}

function Metric({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string | number
}) {
  return (
    <Card variant="outline" className="rounded-lg">
      <CardPanel className="flex items-center gap-3">
        <Icon className="size-5 text-primary" />
        <div>
          <p className="text-sm text-muted-foreground">{label}</p>
          <p className="text-xl font-semibold tracking-normal">{value}</p>
        </div>
      </CardPanel>
    </Card>
  )
}

function FileList({ entries }: { entries: RepositoryTreeEntry[] }) {
  return (
    <div className="overflow-hidden rounded-md border">
      {entries.map((entry) => (
        <div
          key={`${entry.type}:${entry.path}`}
          className="grid grid-cols-[minmax(0,1fr)_6.5rem] items-center gap-3 border-b px-3 py-2.5 last:border-b-0"
        >
          <div className="flex min-w-0 items-center gap-2">
            {entry.type === "tree" ? (
              <Folder className="size-4 shrink-0 text-primary" />
            ) : (
              <FileCode2 className="size-4 shrink-0 text-muted-foreground" />
            )}
            <span className="truncate text-sm font-medium">{entry.name}</span>
          </div>
          <span className="truncate text-right font-mono text-xs text-muted-foreground">
            {entry.sha.slice(0, 7)}
          </span>
        </div>
      ))}
    </div>
  )
}

function EmptyRepository({
  cloneURL,
  branch,
}: {
  cloneURL: string
  branch: string
}) {
  return (
    <Empty className="border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <GitBranch />
        </EmptyMedia>
        <EmptyTitle>This repository is empty</EmptyTitle>
        <EmptyDescription>
          Push an existing project or add a first commit to populate files.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent className="max-w-full items-stretch">
        <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 text-left font-mono text-xs">
          git remote add origin {cloneURL}
        </code>
        <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 text-left font-mono text-xs">
          git branch -M {branch}
        </code>
        <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 text-left font-mono text-xs">
          git push -u origin {branch}
        </code>
      </EmptyContent>
    </Empty>
  )
}

function ReviewBlock({
  title,
  detail,
}: {
  title: string
  detail: string
}) {
  return (
    <article>
      <h2 className="text-sm font-medium">{title}</h2>
      <p className="mt-1 break-all text-sm leading-6 text-muted-foreground">
        {detail}
      </p>
    </article>
  )
}
