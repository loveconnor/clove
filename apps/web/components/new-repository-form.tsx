"use client"

import { useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2, UploadCloud } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Button } from "@loveui/ui/ui/button"
import { Input } from "@loveui/ui/ui/input"
import { Label } from "@loveui/ui/ui/label"
import { Separator } from "@loveui/ui/ui/separator"

import type { Organization, Repository, Viewer } from "@/apps/web/lib/api"

export function NewRepositoryForm({
  viewer,
  organizations,
}: {
  viewer: Viewer
  organizations: Organization[]
}) {
  const router = useRouter()
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const ownerOptions = useMemo(
    () => [
      viewer.user.username,
      ...organizations.map((organization) => organization.name),
    ],
    [organizations, viewer.user.username]
  )

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError("")
    setPending(true)

    const formData = new FormData(event.currentTarget)
    const payload = {
      owner: stringValue(formData.get("owner")),
      name: stringValue(formData.get("name")),
      description: stringValue(formData.get("description")),
      visibility: stringValue(formData.get("visibility")),
    }

    try {
      const response = await fetch("/api/repositories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        const body = (await response.json().catch(() => null)) as
          | { error?: { message?: string } }
          | null
        throw new Error(body?.error?.message ?? "Could not create repository")
      }

      const body = (await response.json()) as { repository: Repository }
      router.push(`/${body.repository.owner}/${body.repository.name}`)
      router.refresh()
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : "Could not create repository"
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="space-y-4">
      {error && (
        <Alert variant="error">
          <AlertTitle>Repository was not created</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <form className="grid gap-5" onSubmit={submit}>
        <div className="grid gap-2">
          <Label htmlFor="owner">Owner</Label>
          <Input
            id="owner"
            name="owner"
            list="repository-owners"
            defaultValue={ownerOptions[0]}
            required
          />
          <datalist id="repository-owners">
            {ownerOptions.map((owner) => (
              <option key={owner} value={owner} />
            ))}
          </datalist>
        </div>
        <div className="grid gap-2">
          <Label htmlFor="name">Repository name</Label>
          <Input id="name" name="name" placeholder="web" required />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="description">Description</Label>
          <Input
            id="description"
            name="description"
            placeholder="Short, portable project summary"
          />
        </div>
        <Separator />
        <fieldset className="grid gap-3">
          <legend className="text-sm font-medium">Visibility</legend>
          <label className="flex cursor-pointer items-start gap-3 rounded-lg border bg-background p-3">
            <input
              type="radio"
              name="visibility"
              value="private"
              defaultChecked
              className="mt-1"
            />
            <span>
              <span className="block text-sm font-medium">Private</span>
              <span className="block text-xs leading-5 text-muted-foreground">
                Visible to members with explicit repository access.
              </span>
            </span>
          </label>
          <label className="flex cursor-pointer items-start gap-3 rounded-lg border bg-background p-3">
            <input
              type="radio"
              name="visibility"
              value="internal"
              className="mt-1"
            />
            <span>
              <span className="block text-sm font-medium">Internal</span>
              <span className="block text-xs leading-5 text-muted-foreground">
                Visible across authenticated CLOVE users for now.
              </span>
            </span>
          </label>
        </fieldset>
        <div className="flex flex-wrap gap-2">
          <Button type="submit" disabled={pending}>
            {pending ? <Loader2 className="animate-spin" /> : null}
            Create repository
          </Button>
          <Button type="button" variant="outline" disabled>
            <UploadCloud />
            Import from GitHub
          </Button>
        </div>
      </form>
    </div>
  )
}

function stringValue(value: FormDataEntryValue | null) {
  return typeof value === "string" ? value.trim() : ""
}
