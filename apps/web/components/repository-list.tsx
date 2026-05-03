import Link from "next/link"
import { GitBranch, LockKeyhole, ShieldCheck } from "lucide-react"

import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Card, CardHeader, CardPanel, CardTitle } from "@loveui/ui/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@loveui/ui/ui/table"

import { EmptyRepositoryState } from "@/apps/web/components/state-views"
import type { Repository } from "@/apps/web/lib/api"

export function RepositoryList({
  title = "Repositories",
  items,
}: {
  title?: string
  items: Repository[]
}) {
  if (items.length === 0) {
    return <EmptyRepositoryState />
  }

  return (
    <Card variant="outline" className="rounded-lg">
      <CardHeader className="border-b">
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardPanel className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Repository</TableHead>
              <TableHead>Review state</TableHead>
              <TableHead>Checks</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead className="w-20 text-right">Open</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((repo) => (
              <TableRow key={`${repo.owner}/${repo.name}`}>
                <TableCell>
                  <div className="flex min-w-0 items-center gap-2">
                    {repo.visibility === "private" ? (
                      <LockKeyhole className="size-4 text-muted-foreground" />
                    ) : (
                      <ShieldCheck className="size-4 text-muted-foreground" />
                    )}
                    <div className="min-w-0">
                      <Link
                        href={`/${repo.owner}/${repo.name}`}
                        className="font-medium text-foreground hover:underline"
                      >
                        {repo.owner}/{repo.name}
                      </Link>
                      <p className="truncate text-xs text-muted-foreground">
                        {repo.description || repo.default_branch}
                      </p>
                    </div>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge
                    variant="outline"
                  >
                    {repo.default_branch}
                  </Badge>
                </TableCell>
                <TableCell>
                  <span className="text-sm text-muted-foreground">
                    {repo.visibility}
                  </span>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatDate(repo.updated_at)}
                </TableCell>
                <TableCell className="text-right">
                  <Button variant="ghost" size="sm" asChild>
                    <Link href={`/${repo.owner}/${repo.name}`}>
                      <GitBranch />
                      Open
                    </Link>
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardPanel>
    </Card>
  )
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value))
}
