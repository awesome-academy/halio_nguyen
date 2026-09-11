import { Suspense } from "react";
import { Compass } from "lucide-react";
import { LoginForm } from "./login-form";

export const metadata = { title: "Admin Login" };

/**
 * Standalone screen (SCR-AdminLogin) — a single centered card with no
 * sidebar/header chrome. AdminShell renders this route's children bare, so
 * this page owns its own minimal layout.
 */
export default function AdminLoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/20 px-4">
      <div className="w-full max-w-sm rounded-xl border bg-card p-8 shadow-sm">
        <div className="flex flex-col items-center gap-2 mb-6">
          <Compass className="h-8 w-8 text-primary" />
          <h1 className="text-lg font-bold">SUN Admin</h1>
          <p className="text-sm text-muted-foreground">Sign in to the management console</p>
        </div>

        <Suspense>
          <LoginForm />
        </Suspense>
      </div>
    </div>
  );
}
