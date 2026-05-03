import Link from "next/link"
import { redirect } from "next/navigation"
import { ArrowLeft, Building2, ShieldCheck, Users } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { NewOrganizationForm } from "@/apps/web/components/new-organization-form"
import { APIRequestError, getViewer } from "@/apps/web/lib/api"

export default async function NewOrganizationPage() {
  await loadNewOrganizationData()

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/dashboard">
            <ArrowLeft />
            Dashboard
          </Link>
        </Button>
      </div>
      <div>
        <p className="text-sm font-medium text-muted-foreground">
          New organization
        </p>
        <h1 className="mt-1 text-3xl font-semibold tracking-normal">
          Create a workspace for your team
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
          Organizations group repositories, members, policy, and audit. You
          become the owner and can invite members and admins after creation.
        </p>
      </div>

      <Alert variant="info">
        <AlertTitle>Names are durable identifiers</AlertTitle>
        <AlertDescription>
          The organization name appears in URLs, repository paths, and webhook
          payloads. Pick something stable; renames are a heavier migration.
        </AlertDescription>
      </Alert>

      <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <CardTitle>Organization details</CardTitle>
          </CardHeader>
          <CardPanel>
            <NewOrganizationForm />
          </CardPanel>
        </Card>

        <div className="space-y-4">
          <Card variant="outline" className="rounded-lg">
            <CardPanel>
              <Building2 className="mb-4 size-5 text-primary" />
              <h2 className="text-sm font-semibold">What gets created</h2>
              <ul className="mt-3 grid gap-2 text-sm text-muted-foreground">
                <li>An organization record owned by your account.</li>
                <li>An owner membership row for you.</li>
                <li>
                  An empty repository namespace at <code>/{"<name>"}</code>.
                </li>
              </ul>
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardPanel className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <span className="flex items-center gap-2 text-sm font-medium">
                  <Users className="size-4" />
                  Members
                </span>
                <Badge variant="outline">After create</Badge>
              </div>
              <p className="text-xs leading-5 text-muted-foreground">
                Invite admins and members from organization settings. SSO,
                SCIM, and guest access plug in later without changing the URL.
              </p>
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardPanel className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <span className="flex items-center gap-2 text-sm font-medium">
                  <ShieldCheck className="size-4 text-success" />
                  Policy source
                </span>
                <Badge variant="success">Inherited</Badge>
              </div>
              <p className="text-xs leading-5 text-muted-foreground">
                Required reviews, token scopes, and audit retention will be
                inherited from organization-level policy when those surfaces
                land.
              </p>
            </CardPanel>
          </Card>
        </div>
      </section>
    </div>
  )
}

async function loadNewOrganizationData() {
  try {
    return await getViewer()
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    throw error
  }
}
