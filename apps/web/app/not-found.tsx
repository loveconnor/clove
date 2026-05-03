import { PageErrorState } from "@/apps/web/components/state-views"

export default function NotFound() {
  return (
    <PageErrorState
      title="Page not found"
      description="The route may have moved, or the repository may not be visible to this account."
    />
  )
}
