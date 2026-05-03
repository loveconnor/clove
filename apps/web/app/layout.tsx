import type { Metadata } from "next"
import "./globals.css"

export const metadata: Metadata = {
  title: "CLOVE",
  description:
    "A review-first, Git-compatible forge for dependable code collaboration.",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" className="h-full antialiased">
      <body className="min-h-full">{children}</body>
    </html>
  )
}
