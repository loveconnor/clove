"use client"

import { useEffect } from "react"

export default function GlobalError({
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
    <html lang="en">
      <body>
        <main style={{ padding: 24, fontFamily: "system-ui, sans-serif" }}>
          <h1>Application error</h1>
          <p>CLOVE could not load the application shell.</p>
          <button type="button" onClick={() => unstable_retry()}>
            Try again
          </button>
        </main>
      </body>
    </html>
  )
}
