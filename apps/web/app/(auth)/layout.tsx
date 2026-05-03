import { BrandMark } from "@/apps/web/components/brand"

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="grid min-h-screen bg-background lg:grid-cols-[minmax(0,1fr)_28rem]">
      <section className="flex min-h-screen items-center justify-center px-4 py-10">
        <div className="w-full max-w-md">{children}</div>
      </section>
      <aside className="hidden border-l bg-card px-10 py-12 lg:flex lg:flex-col">
        <BrandMark />
        <div className="mt-auto">
          <p className="text-2xl font-semibold leading-9 tracking-normal">
            Review-first collaboration with explicit policy, export, and
            failure boundaries.
          </p>
          <dl className="mt-8 grid gap-4 text-sm">
            <div className="rounded-lg border bg-background p-4">
              <dt className="font-medium">Core forge isolation</dt>
              <dd className="mt-1 text-muted-foreground">
                Repository access stays legible when CI, packages, or search
                are degraded.
              </dd>
            </div>
            <div className="rounded-lg border bg-background p-4">
              <dt className="font-medium">Human authorship first</dt>
              <dd className="mt-1 text-muted-foreground">
                Review state separates people, automation, and optional AI
                surfaces.
              </dd>
            </div>
          </dl>
        </div>
      </aside>
    </main>
  )
}
