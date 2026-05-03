"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { CheckCircle2, Loader2 } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Button } from "@loveui/ui/ui/button"
import { Input } from "@loveui/ui/ui/input"
import { Label } from "@loveui/ui/ui/label"
import { Textarea } from "@loveui/ui/ui/textarea"

import type { Organization } from "@/apps/web/lib/api"

export function OrganizationSettingsForm({
  organization,
}: {
  organization: Organization
}) {
  const router = useRouter()
  const [displayName, setDisplayName] = useState(organization.display_name ?? "")
  const [description, setDescription] = useState(organization.description ?? "")
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState(false)

  const dirty =
    displayName !== (organization.display_name ?? "") ||
    description !== (organization.description ?? "")

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError("")
    setSaved(false)
    setPending(true)
    try {
      const response = await fetch(
        `/api/organizations/${encodeURIComponent(organization.name)}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            display_name: displayName.trim(),
            description: description.trim(),
          }),
        }
      )
      if (!response.ok) {
        const body = (await response.json().catch(() => null)) as
          | { error?: { message?: string } }
          | null
        throw new Error(body?.error?.message ?? "Could not save settings")
      }
      setSaved(true)
      router.refresh()
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : "Could not save settings"
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="space-y-4">
      {error && (
        <Alert variant="error">
          <AlertTitle>Settings were not saved</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {saved && !error && (
        <Alert variant="success">
          <CheckCircle2 />
          <AlertTitle>Settings saved</AlertTitle>
          <AlertDescription>
            The organization metadata has been updated.
          </AlertDescription>
        </Alert>
      )}

      <form className="grid gap-5" onSubmit={submit}>
        <div className="grid gap-2">
          <Label htmlFor="org-name">Name</Label>
          <Input
            id="org-name"
            value={organization.name}
            disabled
            aria-describedby="org-name-help"
          />
          <p id="org-name-help" className="text-xs leading-5 text-muted-foreground">
            Renaming an organization is a separate, heavier migration and is
            not yet supported here.
          </p>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="display_name">Display name</Label>
          <Input
            id="display_name"
            name="display_name"
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            placeholder={organization.name}
          />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="description">Description</Label>
          <Textarea
            id="description"
            name="description"
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            placeholder="What is this organization for?"
            rows={3}
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Button type="submit" disabled={pending || !dirty}>
            {pending ? <Loader2 className="animate-spin" /> : null}
            Save changes
          </Button>
          <Button
            type="button"
            variant="ghost"
            disabled={pending || !dirty}
            onClick={() => {
              setDisplayName(organization.display_name ?? "")
              setDescription(organization.description ?? "")
              setSaved(false)
              setError("")
            }}
          >
            Reset
          </Button>
        </div>
      </form>
    </div>
  )
}
