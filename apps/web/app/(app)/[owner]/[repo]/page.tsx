import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import {
  ArrowLeft,
  Code2,
  GitBranch,
  GitPullRequestArrow,
  LockKeyhole,
  ShieldCheck,
} from "lucide-react"

import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { APIRequestError, getRepository } from "@/apps/web/lib/api"

export default async function RepositoryPage({
  params,
}: {
  params: Promise<{ owner: string; repo: string }>
}) {
  const { owner, repo } = await params
  const repository = await loadRepositoryData(owner, repo)

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
            <CardTitle>Repository metadata</CardTitle>
          </CardHeader>
          <CardPanel className="grid gap-4">
            <div className="grid gap-3 md:grid-cols-2">
              <ReviewBlock
                title="Repository ID"
                detail={repository.id}
                badge="DB"
              />
              <ReviewBlock
                title="Git path"
                detail={repository.git_path}
                badge="Git"
              />
              <ReviewBlock
                title="Created"
                detail={formatDate(repository.created_at)}
                badge="Audit"
              />
              <ReviewBlock
                title="Updated"
                detail={formatDate(repository.updated_at)}
                badge="Audit"
              />
            </div>
          </CardPanel>
        </Card>

        <div className="space-y-6">
          <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
              <CardTitle>Review surfaces</CardTitle>
            </CardHeader>
            <CardPanel>
              <p className="text-xs leading-5 text-muted-foreground">
                Pull requests, checks, branch protections, and review snapshots
                are intentionally not faked here. These panels will populate
                when their backend tables and endpoints exist.
              </p>
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardHeader className="border-b">
              <CardTitle>Clone</CardTitle>
            </CardHeader>
            <CardPanel>
              <code className="block overflow-x-auto rounded-md border bg-background px-3 py-2 font-mono text-xs">
                git@clove.dev:{repository.owner}/{repository.name}.git
              </code>
              <div className="mt-3 flex gap-2">
                <Button size="sm" variant="outline">
                  SSH
                </Button>
                <Button size="sm" variant="ghost">
                  HTTPS
                </Button>
              </div>
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
  badge,
}: {
  title: string
  detail: string
  badge: string
}) {
  return (
    <article className="rounded-lg border bg-background p-4">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold">{title}</h2>
        <Badge variant="outline">{badge}</Badge>
      </div>
      <p className="mt-2 text-sm leading-6 text-muted-foreground">{detail}</p>
    </article>
  )
}
