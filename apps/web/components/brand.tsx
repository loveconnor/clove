import Link from "next/link"

export function BrandMark({ compact = false }: { compact?: boolean }) {
  return (
    <Link href="/dashboard" className="flex min-w-0 items-center gap-2">
      <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-foreground text-sm font-semibold text-background">
        C
      </span>
      {!compact && (
        <span className="truncate text-sm font-semibold tracking-normal">
          CLOVE
        </span>
      )}
    </Link>
  )
}
