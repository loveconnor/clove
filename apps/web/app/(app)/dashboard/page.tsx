import Link from "next/link"
import { redirect } from "next/navigation"
import {
  ArrowRight,
  Database,
  GitBranch,
  GitPullRequestArrow,
  UserPlus,
} from "lucide-react"

import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { RepositoryList } from "@/apps/web/components/repository-list"
import {
  APIRequestError,
  getOrganizations,
  getRepositories,
  getViewer,
} from "@/apps/web/lib/api"

export default async function DashboardPage() {
  const { viewer, organizations, repositories } = await loadDashboardData()

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground">
            Dashboard
          </p>
          <h1 className="mt-1 text-3xl font-semibold tracking-normal">
            Review work that needs attention
          </h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
            Current repositories and organization membership loaded from the
            backend for your active WorkOS session.
          </p>
        </div>
        <Button asChild>
          <Link href="/new/repository">
            New repository
            <ArrowRight />
          </Link>
        </Button>
      </div>

      <section className="grid gap-3 md:grid-cols-3">
        <MetricCard
          icon={Database}
          label="Organizations"
          value={organizations.length}
        />
        <MetricCard
          icon={GitBranch}
          label="Repositories"
          value={repositories.length}
        />
        <MetricCard
          icon={GitPullRequestArrow}
          label="Signed in as"
          value={viewer.user.username}
        />
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="space-y-6">
          <RepositoryList items={repositories} />
        </div>

        <div className="space-y-6">
          <Card variant="outline" className="rounded-lg">
            <CardHeader className="border-b">
              <CardTitle>Account</CardTitle>
            </CardHeader>
            <CardPanel className="grid gap-3 text-sm">
              <div>
                <p className="text-muted-foreground">Email</p>
                <p className="font-medium">{viewer.user.email}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Session</p>
                <p className="truncate font-mono text-xs">{viewer.session_id}</p>
              </div>
              {viewer.permissions && viewer.permissions.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {viewer.permissions.map((permission) => (
                    <Badge key={permission} variant="outline">
                      {permission}
                    </Badge>
                  ))}
                </div>
              )}
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardHeader className="flex items-center justify-between gap-3 border-b">
              <CardTitle>Organizations</CardTitle>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/new/organization">
                  <UserPlus />
                  New
                </Link>
              </Button>
            </CardHeader>
            <CardPanel className="grid gap-3">
              {organizations.length === 0 ? (
                <p className="text-sm leading-6 text-muted-foreground">
                  No organizations are associated with this WorkOS user yet.
                  Create one to group repositories, members, and policy.
                </p>
              ) : (
                organizations.map((organization) => (
                  <Link
                    key={organization.id}
                    href={`/${organization.name}`}
                    className="flex items-center justify-between gap-3 rounded-lg border bg-background p-3 text-sm hover:bg-accent/50"
                  >
                    <span className="font-medium">
                      {organization.display_name || organization.name}
                    </span>
                    <Badge variant="outline">{organization.role || "member"}</Badge>
                  </Link>
                ))
              )}
            </CardPanel>
          </Card>
        </div>
      </section>
    </div>
  )
}

async function loadDashboardData() {
  const [viewerResult, organizationsResult, repositoriesResult] =
    await Promise.allSettled([
      getViewer(),
      getOrganizations(),
      getRepositories(),
    ])

  if (viewerResult.status === "rejected") {
    const error = viewerResult.reason
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    throw error
  }

  let organizations: Awaited<ReturnType<typeof getOrganizations>> = []
  if (organizationsResult.status === "fulfilled") {
    organizations = organizationsResult.value
  } else {
    const error = organizationsResult.reason
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    if (
      error instanceof APIRequestError &&
      error.status === 503 &&
      error.code === "service_unavailable"
    ) {
      organizations = []
    } else {
      throw error
    }
  }

  let repositories: Awaited<ReturnType<typeof getRepositories>> = []
  if (repositoriesResult.status === "fulfilled") {
    repositories = repositoriesResult.value
  } else {
    const error = repositoriesResult.reason
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    if (
      error instanceof APIRequestError &&
      error.status === 503 &&
      error.code === "service_unavailable"
    ) {
      repositories = []
    } else {
      throw error
    }
  }

  return { viewer: viewerResult.value, organizations, repositories }
}

function MetricCard({
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
        <div className="min-w-0">
          <p className="text-sm text-muted-foreground">{label}</p>
          <p className="mt-1 truncate text-xl font-semibold tracking-normal">
            {value}
          </p>
        </div>
      </CardPanel>
    </Card>
  )
}
