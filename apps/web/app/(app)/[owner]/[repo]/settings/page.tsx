import Link from "next/link"
import { notFound, redirect } from "next/navigation"
import {
  ArrowLeft,
  Database,
  GitBranch,
  HardDrive,
  LockKeyhole,
  ShieldCheck,
} from "lucide-react"

import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"

import { APIRequestError, getRepository } from "@/apps/web/lib/api"

export default async function RepositorySettingsPage({
  params,
}: {
  params: Promise<{ owner: string; repo: string }>
}) {
  const { owner, repo } = await params
  const repository = await loadRepositoryData(owner, repo)

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div>
        <Button variant="ghost" size="sm" asChild>
          <Link href={`/${repository.owner}/${repository.name}`}>
            <ArrowLeft />
            {repository.owner}/{repository.name}
          </Link>
        </Button>
      </div>

      <header className="flex flex-col gap-2">
        <p className="text-sm font-medium text-muted-foreground">
          Repository settings
        </p>
        <h1 className="text-3xl font-semibold tracking-normal">
          {repository.owner}/{repository.name}
        </h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
          Metadata and storage details for this repository. Visibility is set at
          creation and stored with the repository record.
        </p>
      </header>

      <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <Card variant="outline" className="rounded-lg">
          <CardHeader className="border-b">
            <CardTitle>General</CardTitle>
          </CardHeader>
          <CardPanel className="grid gap-5">
            <ReadOnlyField label="Name" value={repository.name} />
            <ReadOnlyField label="Owner" value={repository.owner} />
            <ReadOnlyField
              label="Description"
              value={repository.description || "No description"}
            />
            <ReadOnlyField
              label="Default branch"
              value={repository.default_branch}
            />
            <div className="grid gap-2">
              <p className="text-sm font-medium">Visibility</p>
              <div className="flex flex-wrap gap-2">
                <Badge variant="outline">{repository.visibility}</Badge>
                {repository.visibility === "private" ? (
                  <Badge variant="outline">
                    <LockKeyhole className="size-3" />
                    Restricted
                  </Badge>
                ) : (
                  <Badge variant="success">
                    <ShieldCheck className="size-3" />
                    Discoverable
                  </Badge>
                )}
              </div>
            </div>
          </CardPanel>
        </Card>

        <div className="space-y-4">
          <Card variant="outline" className="rounded-lg">
            <CardPanel>
              <HardDrive className="mb-4 size-5 text-primary" />
              <h2 className="text-sm font-semibold">Bare Git repository</h2>
              <p className="mt-2 break-all font-mono text-xs leading-5 text-muted-foreground">
                {repository.git_path}
              </p>
            </CardPanel>
          </Card>

          <Card variant="outline" className="rounded-lg">
            <CardPanel className="grid gap-3 text-sm">
              <Detail icon={Database} label="Repository ID" value={repository.id} />
              <Detail
                icon={GitBranch}
                label="Created"
                value={formatDate(repository.created_at)}
              />
              <Detail
                icon={GitBranch}
                label="Updated"
                value={formatDate(repository.updated_at)}
              />
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
    if (error instanceof APIRequestError && error.status === 404) {
      notFound()
    }
    if (
      error instanceof APIRequestError &&
      error.status === 503 &&
      error.code === "service_unavailable"
    ) {
      redirect("/dashboard")
    }
    throw error
  }
}

function ReadOnlyField({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-2">
      <p className="text-sm font-medium">{label}</p>
      <p className="rounded-lg border bg-background px-3 py-2 text-sm text-muted-foreground">
        {value}
      </p>
    </div>
  )
}

function Detail({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string
}) {
  return (
    <div className="flex items-start gap-3">
      <Icon className="mt-0.5 size-4 text-muted-foreground" />
      <div className="min-w-0">
        <p className="text-muted-foreground">{label}</p>
        <p className="break-all font-medium">{value}</p>
      </div>
    </div>
  )
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
