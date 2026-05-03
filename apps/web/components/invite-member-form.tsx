"use client"

import { useState } from "react"
import { CheckCircle2, Info, Loader2 } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@loveui/ui/ui/alert"
import { Button } from "@loveui/ui/ui/button"
import { Input } from "@loveui/ui/ui/input"
import { Label } from "@loveui/ui/ui/label"
import {
  Select,
  SelectItem,
  SelectPopup,
  SelectTrigger,
  SelectValue,
} from "@loveui/ui/ui/select"

const roleOptions: Array<{ value: "owner" | "admin" | "member"; label: string; description: string }> = [
  {
    value: "member",
    label: "Member",
    description: "Read repositories they're added to. Default for most invitees.",
  },
  {
    value: "admin",
    label: "Admin",
    description: "Manage repositories, members, and settings.",
  },
  {
    value: "owner",
    label: "Owner",
    description: "Full control, including deleting the organization.",
  },
]

export function InviteMemberForm({
  organizationName,
}: {
  organizationName: string
}) {
  const [email, setEmail] = useState("")
  const [role, setRole] = useState<"owner" | "admin" | "member">("member")
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [info, setInfo] = useState("")

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError("")
    setInfo("")

    const trimmed = email.trim()
    if (!trimmed) {
      setError("Email is required")
      return
    }

    setPending(true)
    try {
      const response = await fetch(
        `/api/organizations/${encodeURIComponent(organizationName)}/invitations`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: trimmed, role }),
        }
      )
      const body = (await response.json().catch(() => null)) as
        | {
            error?: { message?: string }
            note?: string
            invitation?: { email?: string }
          }
        | null
      if (!response.ok) {
        throw new Error(body?.error?.message ?? "Could not create invitation")
      }
      setInfo(
        body?.note ??
          `Invitation queued for ${body?.invitation?.email ?? trimmed}.`
      )
      setEmail("")
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "Could not create invitation"
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="space-y-4">
      <Alert variant="info">
        <Info />
        <AlertTitle>Invitation delivery is not yet wired up</AlertTitle>
        <AlertDescription>
          The form posts to the API and is accepted, but no email is sent and
          no membership is created until the invitation pipeline lands.
        </AlertDescription>
      </Alert>

      {error && (
        <Alert variant="error">
          <AlertTitle>Invitation was not sent</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {info && !error && (
        <Alert variant="success">
          <CheckCircle2 />
          <AlertTitle>Invitation accepted</AlertTitle>
          <AlertDescription>{info}</AlertDescription>
        </Alert>
      )}

      <form className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_10rem_auto]" onSubmit={submit}>
        <div className="grid gap-2">
          <Label htmlFor="invite-email">Email</Label>
          <Input
            id="invite-email"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="teammate@example.com"
            required
          />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="invite-role">Role</Label>
          <Select
            value={role}
            onValueChange={(value) =>
              setRole(value as "owner" | "admin" | "member")
            }
          >
            <SelectTrigger id="invite-role">
              <SelectValue>
                {(value) =>
                  roleOptions.find((option) => option.value === value)?.label ??
                  value
                }
              </SelectValue>
            </SelectTrigger>
            <SelectPopup>
              {roleOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  <span className="flex flex-col">
                    <span>{option.label}</span>
                    <span className="text-xs text-muted-foreground">
                      {option.description}
                    </span>
                  </span>
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className="flex items-end">
          <Button type="submit" disabled={pending} className="w-full sm:w-auto">
            {pending ? <Loader2 className="animate-spin" /> : null}
            Send invite
          </Button>
        </div>
      </form>
    </div>
  )
}
