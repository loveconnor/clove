import Link from "next/link"
import { redirect } from "next/navigation"
import { ArrowLeft, GitFork } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { NewRepositoryForm } from "@/apps/web/components/new-repository-form"
import { APIRequestError, getOrganizations, getViewer } from "@/apps/web/lib/api"

export default async function NewRepositoryPage() {
  const { viewer, organizations } = await loadNewRepositoryData()

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
          New repository
        </p>
        <h1 className="mt-1 text-3xl font-semibold tracking-normal">
          Create a dependable workspace
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
          Start empty or import from GitHub while keeping migration, policy, and
          review-state boundaries visible.
        </p>
      </div>

      <Alert variant="info">
        <AlertTitle>Import and export stay first-class</AlertTitle>
        <AlertDescription>
          Repository metadata, issues, pull requests, labels, releases, and wiki
          data are modeled as portable surfaces.
        </AlertDescription>
      </Alert>

      <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <CardTitle>Repository details</CardTitle>
          </CardHeader>
          <CardPanel>
            <NewRepositoryForm viewer={viewer} organizations={organizations} />
          </CardPanel>
        </Card>

        <div className="space-y-4">
          <Card variant="outline" className="rounded-lg">
            <CardPanel>
              <GitFork className="mb-4 size-5 text-primary" />
              <h2 className="text-sm font-semibold">Creation plan</h2>
              <ul className="mt-3 grid gap-2 text-sm text-muted-foreground">
                <li>Create the repository record in Postgres.</li>
                <li>Attach it to your WorkOS-backed user or organization.</li>
                <li>Leave CI and package planes empty until their APIs exist.</li>
              </ul>
            </CardPanel>
          </Card>
          <Card variant="outline" className="rounded-lg">
            <CardPanel className="grid gap-2">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium">Policy source</span>
                <Badge variant="success">Inherited</Badge>
              </div>
              <p className="text-xs leading-5 text-muted-foreground">
                SSO, required reviews, token scopes, and audit retention come
                from the authenticated user and database membership.
              </p>
            </CardPanel>
          </Card>
        </div>
      </section>
    </div>
  )
}

async function loadNewRepositoryData() {
  try {
    const [viewer, organizations] = await Promise.all([
      getViewer(),
      getOrganizations(),
    ])
    return { viewer, organizations }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    throw error
  }
}
