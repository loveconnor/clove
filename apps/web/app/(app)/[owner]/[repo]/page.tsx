import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import {
  ArrowLeft,
  Code2,
  GitBranch,
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

import { APIRequestError, getRepository } from "@/apps/web/lib/api"

export default async function RepositoryPage({
  params,
}: {
  params: Promise<{ owner: string; repo: string }>
}) {
  const { owner, repo } = await params
  const repository = await loadRepositoryData(owner, repo)
  const cloneURL = `https://clove.dev/${repository.owner}/${repository.name}.git`

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
            <Badge variant="success">DB backed</Badge>
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
            <CardTitle>Files</CardTitle>
          </CardHeader>
          <CardPanel>
            <Empty className="border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <GitBranch />
                </EmptyMedia>
                <EmptyTitle>This repository is empty</EmptyTitle>
                <EmptyDescription>
                  Push an existing project or add a first commit to populate
                  files, branches, pull requests, and checks.
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent className="max-w-full items-stretch">
                <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 text-left font-mono text-xs">
                  git remote add origin {cloneURL}
                </code>
                <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 text-left font-mono text-xs">
                  git branch -M {repository.default_branch}
                </code>
                <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 text-left font-mono text-xs">
                  git push -u origin {repository.default_branch}
                </code>
              </EmptyContent>
            </Empty>
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
    </div>
  )
}

async function loadRepositoryData(owner: string, repo: string) {
  try {
    return await getRepository(owner, repo)
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
