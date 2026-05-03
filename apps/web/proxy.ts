import { NextResponse, type NextRequest } from "next/server"

const authCookies = ["clove_access_token", "clove_refresh_token"]

export function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl

  if (pathname === "/login") {
    return NextResponse.redirect(new URL(`/api/auth/login${search}`, request.url))
  }
  if (pathname === "/register" || pathname === "/signup") {
    return NextResponse.redirect(new URL(`/api/auth/register${search}`, request.url))
  }

  const hasSessionCookie = authCookies.some((name) =>
    Boolean(request.cookies.get(name)?.value)
  )

  if (hasSessionCookie) {
    return NextResponse.next()
  }

  const loginURL = new URL("/login", request.url)
  return NextResponse.redirect(loginURL)
}

export const config = {
  matcher: [
    "/",
    "/dashboard/:path*",
    "/new/:path*",
    "/login",
    "/register",
    "/signup",
    "/((?!api|_next|favicon.ico).*)",
  ],
}
