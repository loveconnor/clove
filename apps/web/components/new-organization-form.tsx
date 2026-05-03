"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2 } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Button } from "@loveui/ui/ui/button"
import { Input } from "@loveui/ui/ui/input"
import { Label } from "@loveui/ui/ui/label"
import { Textarea } from "@loveui/ui/ui/textarea"

import type { Organization } from "@/apps/web/lib/api"

const namePattern = /^[a-z0-9](?:[a-z0-9-]{0,38}[a-z0-9])?$/

export function NewOrganizationForm() {
  const router = useRouter()
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [name, setName] = useState("")
  const [displayName, setDisplayName] = useState("")
  const [description, setDescription] = useState("")

  const nameInvalid = name.length > 0 && !namePattern.test(name)

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError("")

    const trimmedName = name.trim().toLowerCase()
    if (!trimmedName) {
      setError("Name is required")
      return
    }
    if (!namePattern.test(trimmedName)) {
      setError(
        "Name must be 1-40 characters using lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen."
      )
      return
    }

    setPending(true)
    try {
      const response = await fetch("/api/organizations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: trimmedName,
          display_name: displayName.trim(),
          description: description.trim(),
        }),
      })

      if (!response.ok) {
        const body = (await response.json().catch(() => null)) as
          | { error?: { message?: string } }
          | null
        throw new Error(
          body?.error?.message ?? "Could not create organization"
        )
      }

      const body = (await response.json()) as { organization: Organization }
      router.push(`/${body.organization.name}`)
      router.refresh()
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "Could not create organization"
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="space-y-4">
      {error && (
        <Alert variant="error">
          <AlertTitle>Organization was not created</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <form className="grid gap-5" onSubmit={submit}>
        <div className="grid gap-2">
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            name="name"
            value={name}
            onChange={(event) =>
              setName(event.target.value.toLowerCase().replace(/\s+/g, "-"))
            }
            placeholder="acme-co"
            autoComplete="off"
            spellCheck={false}
            aria-invalid={nameInvalid || undefined}
            aria-describedby="name-help"
            required
          />
          <p id="name-help" className="text-xs leading-5 text-muted-foreground">
            Lowercase letters, numbers, and hyphens. Used in URLs and
            repository paths.
          </p>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="display_name">Display name</Label>
          <Input
            id="display_name"
            name="display_name"
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            placeholder="Acme, Inc."
            autoComplete="off"
          />
          <p className="text-xs leading-5 text-muted-foreground">
            Optional. Shown in navigation and member lists. Defaults to the
            organization name.
          </p>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="description">Description</Label>
          <Textarea
            id="description"
            name="description"
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            placeholder="Short summary of what this organization is for"
            rows={3}
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Button type="submit" disabled={pending || nameInvalid}>
            {pending ? <Loader2 className="animate-spin" /> : null}
            Create organization
          </Button>
          <Button
            type="button"
            variant="ghost"
            disabled={pending}
            onClick={() => router.push("/dashboard")}
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  )
}
