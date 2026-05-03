import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import { ArrowRight, Building2, ShieldCheck, Users } from "lucide-react"

import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { RepositoryList } from "@/apps/web/components/repository-list"
import {
  APIRequestError,
  getOrganization,
  getRepositories,
  getViewer,
  type Organization,
} from "@/apps/web/lib/api"

export default async function OwnerPage({
  params,
}: {
  params: Promise<{ owner: string }>
}) {
  const { owner } = await params
  const { organization, repositories } = await loadOwnerData(owner)
  const displayName = organization?.display_name || organization?.name || owner
  const role = organization?.role || "owner"

  return (
    <div className="space-y-6">
      <section className="rounded-lg border bg-card p-6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-3">
              <div className="flex size-10 items-center justify-center rounded-lg border bg-background">
                <Building2 className="size-5" />
              </div>
              <div className="min-w-0">
                <p className="text-sm text-muted-foreground">Organization</p>
                <h1 className="truncate text-3xl font-semibold tracking-normal">
                  {displayName}
                </h1>
              </div>
            </div>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-muted-foreground">
              Repositories, inherited policy, active review work, and migration
              boundaries for this organization.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant="outline">{role}</Badge>
            <Badge variant="success">WorkOS session</Badge>
          </div>
        </div>
      </section>

      <section className="grid gap-3 md:grid-cols-3">
        <Card variant="outline" className="rounded-lg">
          <CardPanel className="flex items-center gap-3">
            <Users className="size-5 text-primary" />
            <div>
              <p className="text-sm text-muted-foreground">Members</p>
              <p className="text-xl font-semibold tracking-normal">
                {role}
              </p>
            </div>
          </CardPanel>
        </Card>
        <Card variant="outline" className="rounded-lg">
          <CardPanel className="flex items-center gap-3">
            <ShieldCheck className="size-5 text-success" />
            <div>
              <p className="text-sm text-muted-foreground">Policy</p>
              <p className="text-xl font-semibold tracking-normal">Inherited</p>
            </div>
          </CardPanel>
        </Card>
        <Card variant="outline" className="rounded-lg">
          <CardPanel className="flex items-center justify-between gap-3">
            <div>
              <p className="text-sm text-muted-foreground">Repositories</p>
              <p className="text-xl font-semibold tracking-normal">
                {repositories.length}
              </p>
            </div>
            <Button variant="outline" size="sm" asChild>
              <Link href="/new">
                New
                <ArrowRight />
              </Link>
            </Button>
          </CardPanel>
        </Card>
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_20rem]">
        <RepositoryList
          title={`${displayName} repositories`}
          items={repositories}
        />
        <Card variant="outline" className="h-fit rounded-lg">
          <CardHeader className="border-b">
            <CardTitle>Effective access</CardTitle>
          </CardHeader>
          <CardPanel className="grid gap-4 text-sm">
            <div>
              <p className="font-medium">Why you can create repositories</p>
              <p className="mt-1 leading-6 text-muted-foreground">
                Your current role is loaded from the organization membership
                stored in the backend database.
              </p>
            </div>
            <div>
              <p className="font-medium">Guest access</p>
              <p className="mt-1 leading-6 text-muted-foreground">
                External guest access will be driven by organization membership
                and repository grants, not frontend fixtures.
              </p>
            </div>
          </CardPanel>
        </Card>
      </section>
    </div>
  )
}

async function loadOwnerData(owner: string) {
  try {
    const viewer = await getViewer()
    let organization: Organization | null = null
    try {
      organization = await getOrganization(owner)
    } catch (error) {
      if (
        error instanceof APIRequestError &&
        error.status === 404 &&
        owner !== viewer.user.username
      ) {
        notFound()
      }
      if (
        error instanceof APIRequestError &&
        error.status === 503 &&
        error.code === "service_unavailable"
      ) {
        organization = null
      } else if (!(error instanceof APIRequestError && error.status === 404)) {
        throw error
      }
    }
    let repositories: Awaited<ReturnType<typeof getRepositories>> = []
    try {
      repositories = await getRepositories(owner)
    } catch (error) {
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
    return { viewer, organization, repositories }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    throw error
  }
}
