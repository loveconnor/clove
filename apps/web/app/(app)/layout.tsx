import { redirect } from "next/navigation"

import { AppShell } from "@/apps/web/components/app-shell"
import { APIRequestError, getOrganizations, getViewer } from "@/apps/web/lib/api"

export default async function ProductLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { viewer, organizations } = await loadShellData()

  return (
    <AppShell viewer={viewer} organizations={organizations}>
      {children}
    </AppShell>
  )
}

async function loadShellData() {
  const [viewerResult, organizationsResult] = await Promise.allSettled([
    getViewer(),
    getOrganizations(),
  ])

  if (viewerResult.status === "rejected") {
    const error = viewerResult.reason
    if (error instanceof APIRequestError && error.status === 401) {
      redirect("/login")
    }
    throw error
  }

  let organizations = [] as Awaited<ReturnType<typeof getOrganizations>>
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

  return { viewer: viewerResult.value, organizations }
}
