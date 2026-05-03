"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Bell,
  Building2,
  ChevronDown,
  GitPullRequestArrow,
  Home,
  LogOut,
  Plus,
  Search,
  Settings,
  ShieldCheck,
  User,
  UserPlus,
} from "lucide-react"

import { Avatar, AvatarFallback } from "@loveui/ui/ui/avatar"
import { Badge } from "@loveui/ui/ui/badge"
import { Button } from "@loveui/ui/ui/button"
import { Input } from "@loveui/ui/ui/input"
import {
  Menu,
  MenuGroup,
  MenuGroupLabel,
  MenuItem,
  MenuPopup,
  MenuSeparator,
  MenuTrigger,
} from "@loveui/ui/ui/menu"

import { BrandMark } from "@/apps/web/components/brand"
import type { Organization, Viewer } from "@/apps/web/lib/api"

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: Home },
  { href: "/clove", label: "Repositories", icon: GitPullRequestArrow },
  { href: "/new/repository", label: "New", icon: Plus },
]

export function AppShell({
  children,
  organizations,
  viewer,
}: {
  children: React.ReactNode
  organizations: Organization[]
  viewer: Viewer
}) {
  const pathname = usePathname()

  return (
    <div className="min-h-screen bg-background text-foreground">
      <aside className="fixed inset-y-0 left-0 hidden w-64 border-r bg-card/80 px-3 py-4 backdrop-blur md:flex md:flex-col">
        <div className="px-2">
          <BrandMark />
        </div>
        <div className="mt-5">
          <OrganizationSwitcher organizations={organizations} viewer={viewer} />
        </div>
        <nav aria-label="Main navigation" className="mt-5 grid gap-1">
          {navItems.map((item) => {
            const isActive =
              item.href === "/dashboard"
                ? pathname === item.href
                : pathname.startsWith(item.href)
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={isActive ? "page" : undefined}
                className={`flex h-9 items-center gap-2 rounded-md px-2 text-sm transition-colors ${
                  isActive
                    ? "bg-accent text-accent-foreground"
                    : "text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                }`}
              >
                <item.icon className="size-4" />
                <span className="truncate">{item.label}</span>
              </Link>
            )
          })}
        </nav>
        <div className="mt-auto rounded-lg border bg-background p-3">
          <div className="flex items-center gap-2 text-sm font-medium">
            <ShieldCheck className="size-4 text-success" />
            Core forge
          </div>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            Repository access and review actions are operational.
          </p>
        </div>
      </aside>

      <header className="sticky top-0 z-40 border-b bg-card/90 backdrop-blur md:pl-64">
        <div className="flex h-14 items-center gap-3 px-4">
          <div className="md:hidden">
            <BrandMark compact />
          </div>
          <div className="hidden min-w-0 flex-1 md:block">
            <label className="relative block max-w-xl">
              <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="search"
                aria-label="Search repositories, pull requests, issues, and docs"
                placeholder="Search code, reviews, issues, packages..."
                className="pl-8"
              />
            </label>
          </div>
          <nav
            aria-label="Mobile navigation"
            className="ml-auto flex items-center gap-1 md:hidden"
          >
            {navItems.map((item) => {
              const isActive =
                item.href === "/dashboard"
                  ? pathname === item.href
                  : pathname.startsWith(item.href)
              return (
                <Button
                  key={item.href}
                  variant={isActive ? "secondary" : "ghost"}
                  size="icon-sm"
                  asChild
                >
                  <Link href={item.href} aria-label={item.label}>
                    <item.icon />
                  </Link>
                </Button>
              )
            })}
          </nav>
          <Button variant="ghost" size="icon" aria-label="Notifications">
            <Bell />
          </Button>
          <UserMenu viewer={viewer} />
        </div>
      </header>

      <main className="md:pl-64">
        <div className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
          {children}
        </div>
      </main>
    </div>
  )
}

function OrganizationSwitcher({
  organizations,
  viewer,
}: {
  organizations: Organization[]
  viewer: Viewer
}) {
  const activeOrganization = organizations[0]
  const label =
    activeOrganization?.display_name ||
    activeOrganization?.name ||
    viewer.user.username

  return (
    <Menu>
      <MenuTrigger
        render={
          <Button
            variant="outline"
            className="w-full justify-between overflow-hidden"
          >
            <span className="flex min-w-0 items-center gap-2">
              <Building2 className="size-4" />
              <span className="truncate">{label}</span>
            </span>
            <ChevronDown className="size-4" />
          </Button>
        }
      />
      <MenuPopup align="start" className="w-64">
        <MenuGroup>
          <MenuGroupLabel>Personal</MenuGroupLabel>
          <MenuItem render={<Link href={`/${viewer.user.username}`} />}>
            <span className="flex min-w-0 flex-1 flex-col">
              <span className="truncate">{viewer.user.username}</span>
              <span className="truncate text-xs text-muted-foreground">
                Personal namespace
              </span>
            </span>
          </MenuItem>
        </MenuGroup>
        {organizations.length > 0 && (
          <MenuGroup>
            <MenuGroupLabel>Organizations</MenuGroupLabel>
            {organizations.map((organization) => (
              <MenuItem
                key={organization.id}
                render={<Link href={`/${organization.name}`} />}
              >
                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate">
                    {organization.display_name || organization.name}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {organization.role || "member"}
                  </span>
                </span>
                <Badge variant="outline">Org</Badge>
              </MenuItem>
            ))}
          </MenuGroup>
        )}
        <MenuSeparator />
        <MenuItem render={<Link href="/new/organization" />}>
          <UserPlus />
          New organization
        </MenuItem>
      </MenuPopup>
    </Menu>
  )
}

function UserMenu({ viewer }: { viewer: Viewer }) {
  const initials =
    (viewer.user.display_name || viewer.user.username || viewer.user.email)
      .split(/\s+|@/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join("") || "CL"

  return (
    <Menu>
      <MenuTrigger
        render={
          <Button
            variant="ghost"
            size="icon"
            aria-label="Open user menu"
            className="rounded-full"
          >
            <Avatar>
              <AvatarFallback>{initials}</AvatarFallback>
            </Avatar>
          </Button>
        }
      />
      <MenuPopup align="end" className="w-56">
        <MenuGroup>
          <MenuGroupLabel>
            {viewer.user.display_name || viewer.user.email}
          </MenuGroupLabel>
          <MenuItem>
            <User />
            Profile
          </MenuItem>
          <MenuItem>
            <Settings />
            Settings
          </MenuItem>
        </MenuGroup>
        <MenuSeparator />
        <MenuItem render={<Link href="/api/auth/logout" />}>
          <LogOut />
          Sign out
        </MenuItem>
      </MenuPopup>
    </Menu>
  )
}
