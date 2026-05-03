import { redirect } from "next/navigation"

export default function SignupPage() {
  redirect("/api/auth/register")
}
