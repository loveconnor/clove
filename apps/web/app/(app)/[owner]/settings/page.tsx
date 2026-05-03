import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import {
  ArrowLeft,
  Building2,
  ShieldCheck,
  UserPlus,
  Users,
} from "lucide-react"

import { Avatar, AvatarFallback } from "@loveui/ui/ui/avatar"
import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@loveui/ui/ui/empty"

import { OrganizationSettingsForm } from "@/apps/web/components/organization-settings-form"
import { InviteMemberForm } from "@/apps/web/components/invite-member-form"
import {
  APIRequestError,
  getOrganization,
  getOrganizationMembers,
  getViewer,
  type OrganizationMember,
} from "@/apps/web/lib/api"

export default async function OrganizationSettingsPage({
  params,
}: {
  params: Promise<{ owner: string }>
}) {
  const { owner } = await params
  const { organization, members } = await loadSettingsData(owner)
  const role = organization.role || "member"
  const canManage = role === "owner" || role === "admin"
  if (!canManage) {
    redirect(`/${owner}`)
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" asChild>
          <Link href={`/${owner}`}>
            <ArrowLeft />
            {organization.display_name || organization.name}
          </Link>
        </Button>
      </div>

      <header className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">
          Organization settings
        </p>
        <h1 className="text-3xl font-semibold tracking-normal">
          {organization.display_name || organization.name}
        </h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
          Edit organization metadata, review members, and manage invitations.
          Inherited policy, SSO, and audit retention will land here as their
          backend surfaces are wired up.
        </p>
      </header>

      <section
        id="general"
        className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]"
      >
        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <CardTitle>General</CardTitle>
          </CardHeader>
          <CardPanel>
            <OrganizationSettingsForm organization={organization} />
          </CardPanel>
        </Card>

        <div className="space-y-4">
          <Card variant="outline" className="rounded-lg">
            <CardPanel>
              <Building2 className="mb-4 size-5 text-primary" />
              <h2 className="text-sm font-semibold">Identity</h2>
              <dl className="mt-3 grid gap-2 text-sm">
                <div className="flex justify-between gap-2">
                  <dt className="text-muted-foreground">Name</dt>
                  <dd className="font-mono">{organization.name}</dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-muted-foreground">ID</dt>
                  <dd className="truncate font-mono text-xs">
                    {organization.id}
                  </dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-muted-foreground">Created</dt>
                  <dd>{formatDate(organization.created_at)}</dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-muted-foreground">Updated</dt>
                  <dd>{formatDate(organization.updated_at)}</dd>
                </div>
              </dl>
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardPanel className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <span className="flex items-center gap-2 text-sm font-medium">
                  <ShieldCheck className="size-4 text-success" />
                  Policy
                </span>
                <Badge variant="success">Inherited</Badge>
              </div>
              <p className="text-xs leading-5 text-muted-foreground">
                Required reviews, token scopes, and audit retention come from
                organization-level policy. Editable surfaces will appear here
                when their endpoints land.
              </p>
            </CardPanel>
          </Card>
        </div>
      </section>

      <section id="members" className="space-y-4">
        <header className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="text-xl font-semibold tracking-normal">Members</h2>
            <p className="text-sm leading-6 text-muted-foreground">
              People with access to this organization and its repositories.
            </p>
          </div>
          <Badge variant="outline">{members.length} total</Badge>
        </header>

        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <CardTitle className="flex items-center gap-2">
              <UserPlus className="size-4" />
              Invite a member
            </CardTitle>
          </CardHeader>
          <CardPanel>
            <InviteMemberForm organizationName={organization.name} />
          </CardPanel>
        </Card>

        {members.length === 0 ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users />
              </EmptyMedia>
              <EmptyTitle>No members loaded</EmptyTitle>
              <EmptyDescription>
                Once invitations and SCIM are wired up, members will appear
                here with their roles, last-seen activity, and audit links.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <Card variant="outline" className="rounded-lg">
            <CardPanel className="grid gap-1 p-2">
              {members.map((member) => (
                <MemberRow key={member.user_id} member={member} />
              ))}
            </CardPanel>
          </Card>
        )}
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
    <div className="flex items-center gap-3 rounded-md px-3 py-2">
      <Avatar className="size-9">
        <AvatarFallback>{initials}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">
          {member.display_name || member.username}
        </p>
        <p className="truncate text-xs text-muted-foreground">
          @{member.username} · {member.email}
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Badge variant="outline">{member.role}</Badge>
        <span className="hidden text-xs text-muted-foreground sm:inline">
          Joined {formatDate(member.joined_at)}
        </span>
      </div>
    </div>
  )
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value))
}

async function loadSettingsData(owner: string) {
  try {
    const viewer = await getViewer()
    const organization = await getOrganization(owner)
    let members: OrganizationMember[] = []
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
    return { viewer, organization, members }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    if (error instanceof APIRequestError && error.status === 404) {
      notFound()
    }
    if (error instanceof APIRequestError && error.status === 403) {
      redirect(`/${owner}`)
    }
    throw error
  }
}
