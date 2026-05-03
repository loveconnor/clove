import Link from "next/link"
import { AlertTriangle, Inbox, Loader2, RotateCw } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Button } from "@loveui/ui/ui/button"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@loveui/ui/ui/empty"
import { Skeleton } from "@loveui/ui/ui/skeleton"

export function InlineErrorState({
  title,
  description,
  action,
}: {
  title: string
  description: string
  action?: React.ReactNode
}) {
  return (
    <Alert variant="error">
      <AlertTriangle />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
      {action}
    </Alert>
  )
}

export function EmptyRepositoryState() {
  return (
    <Empty className="border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Inbox />
        </EmptyMedia>
        <EmptyTitle>No repositories yet</EmptyTitle>
        <EmptyDescription>
          Create a repository or import from GitHub to start review, policy, and
          migration workflows.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button asChild>
          <Link href="/new">New repository</Link>
        </Button>
      </EmptyContent>
    </Empty>
  )
}

export function PageErrorState({
  title,
  description,
  onRetry,
}: {
  title: string
  description: string
  onRetry?: () => void
}) {
  return (
    <main className="flex min-h-screen items-center justify-center p-6">
      <section className="w-full max-w-lg rounded-xl border bg-card p-8 shadow-sm">
        <div className="mb-6 flex size-10 items-center justify-center rounded-lg bg-destructive/8 text-destructive">
          <AlertTriangle className="size-5" />
        </div>
        <h1 className="text-2xl font-semibold tracking-normal">{title}</h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          {description}
        </p>
        <div className="mt-6 flex flex-wrap gap-2">
          {onRetry && (
            <Button onClick={onRetry}>
              <RotateCw />
              Try again
            </Button>
          )}
          <Button variant="outline" asChild>
            <Link href="/dashboard">Back to dashboard</Link>
          </Button>
        </div>
      </section>
    </main>
  )
}

export function ShellLoadingState() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div className="space-y-2">
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-72" />
        </div>
        <Skeleton className="h-9 w-28" />
      </div>
      <div className="grid gap-3 md:grid-cols-3">
        <Skeleton className="h-28" />
        <Skeleton className="h-28" />
        <Skeleton className="h-28" />
      </div>
      <Skeleton className="h-80" />
    </div>
  )
}

export function RouteLoadingState() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex items-center gap-2 rounded-lg border bg-card px-3 py-2 text-sm text-muted-foreground shadow-sm">
        <Loader2 className="size-4 animate-spin text-primary" />
        Loading workspace
      </div>
    </div>
  )
}
