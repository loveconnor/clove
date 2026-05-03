import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import {
  ArrowRight,
  Building2,
  Settings,
  ShieldCheck,
  UserPlus,
  Users,
} from "lucide-react"

import { Avatar, AvatarFallback } from "@loveui/ui/ui/avatar"
import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { RepositoryList } from "@/apps/web/components/repository-list"
import {
  APIRequestError,
  getOrganization,
  getOrganizationMembers,
  getRepositories,
  getViewer,
  type Organization,
  type OrganizationMember,
} from "@/apps/web/lib/api"

export default async function OwnerPage({
  params,
}: {
  params: Promise<{ owner: string }>
}) {
  const { owner } = await params
  const { organization, members, repositories } = await loadOwnerData(owner)
  const displayName = organization?.display_name || organization?.name || owner
  const role = organization?.role || "owner"
  const canManage = role === "owner" || role === "admin"
  const isOrganization = Boolean(organization)
  const description =
    organization?.description ||
    "Repositories, inherited policy, active review work, and migration boundaries for this namespace."

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
                <p className="text-sm text-muted-foreground">
                  {isOrganization ? "Organization" : "Personal namespace"}
                </p>
                <h1 className="truncate text-3xl font-semibold tracking-normal">
                  {displayName}
                </h1>
                <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
                  /{owner}
                </p>
              </div>
            </div>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-muted-foreground">
              {description}
            </p>
          </div>
          <div className="flex flex-wrap items-start gap-2">
            <Badge variant="outline">{role}</Badge>
            {isOrganization && canManage && (
              <Button variant="outline" size="sm" asChild>
                <Link href={`/${owner}/settings`}>
                  <Settings />
                  Settings
                </Link>
              </Button>
            )}
            <Button size="sm" asChild>
              <Link href="/new/repository">
                <ArrowRight />
                New repository
              </Link>
            </Button>
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
                {isOrganization ? members.length : 1}
              </p>
            </div>
          </CardPanel>
        </Card>
        <Card variant="outline" className="rounded-lg">
          <CardPanel className="flex items-center gap-3">
            <ShieldCheck className="size-5 text-success" />
            <div>
              <p className="text-sm text-muted-foreground">Policy</p>
              <p className="text-xl font-semibold tracking-normal">
                Inherited
              </p>
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
              <Link href="/new/repository">
                New
                <ArrowRight />
              </Link>
            </Button>
          </CardPanel>
        </Card>
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <RepositoryList
          title={`${displayName} repositories`}
          items={repositories}
        />
        <div className="space-y-6">
          {isOrganization && (
            <Card variant="outline" className="h-fit rounded-lg">
              <CardHeader className="flex items-center justify-between border-b">
                <CardTitle>Members</CardTitle>
                {canManage && (
                  <Button variant="ghost" size="sm" asChild>
                    <Link href={`/${owner}/settings#members`}>
                      <UserPlus />
                      Invite
                    </Link>
                  </Button>
                )}
              </CardHeader>
              <CardPanel className="grid gap-2 p-3">
                {members.length === 0 ? (
                  <p className="px-2 py-1 text-sm text-muted-foreground">
                    No members loaded.
                  </p>
                ) : (
                  members.slice(0, 5).map((member) => (
                    <MemberRow key={member.user_id} member={member} />
                  ))
                )}
                {members.length > 5 && (
                  <Link
                    href={`/${owner}/settings#members`}
                    className="px-2 pt-1 text-sm text-muted-foreground hover:text-foreground"
                  >
                    View all {members.length} members
                  </Link>
                )}
              </CardPanel>
            </Card>
          )}

          <Card variant="outline" className="h-fit rounded-lg">
            <CardHeader className="border-b">
              <CardTitle>Effective access</CardTitle>
            </CardHeader>
            <CardPanel className="grid gap-4 text-sm">
              <div>
                <p className="font-medium">Why you can create repositories</p>
                <p className="mt-1 leading-6 text-muted-foreground">
                  Your role is loaded from organization membership stored in
                  the database. Owners and admins can change settings.
                </p>
              </div>
              <div>
                <p className="font-medium">Guest access</p>
                <p className="mt-1 leading-6 text-muted-foreground">
                  External guest access will be driven by organization
                  membership and repository grants, not frontend fixtures.
                </p>
              </div>
            </CardPanel>
          </Card>
        </div>
      </section>
    </div>
  )
}

function MemberRow({ member }: { member: OrganizationMember }) {
  const initials =
    (member.display_name || member.username || member.email)
      .split(/\s+|@/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join("") || "??"

  return (
    <div className="flex items-center gap-3 rounded-md px-2 py-1.5">
      <Avatar className="size-8">
        <AvatarFallback>{initials}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">
          {member.display_name || member.username}
        </p>
        <p className="truncate text-xs text-muted-foreground">
          @{member.username}
        </p>
      </div>
      <Badge variant="outline">{member.role}</Badge>
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

    let members: OrganizationMember[] = []
    if (organization) {
      try {
        members = await getOrganizationMembers(owner)
      } catch (error) {
        if (
          error instanceof APIRequestError &&
          error.status === 503 &&
          error.code === "service_unavailable"
        ) {
          members = []
        } else if (
          !(error instanceof APIRequestError && error.status === 404)
        ) {
          throw error
        }
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
    return { viewer, organization, members, repositories }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    throw error
  }
}
