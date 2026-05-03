"use client"

import { useEffect } from "react"

import { PageErrorState } from "@/apps/web/components/state-views"

export default function Error({
  error,
  unstable_retry,
}: {
  error: Error & { digest?: string }
  unstable_retry: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <PageErrorState
      title="This workspace could not render"
      description="The shell stayed available, but this route hit an unexpected rendering error."
      onRetry={unstable_retry}
    />
  )
}
