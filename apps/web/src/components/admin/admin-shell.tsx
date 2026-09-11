"use client";

import { useEffect } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { LogOut, Compass, ExternalLink, Bell } from "lucide-react";
import { ADMIN_NAV_ITEMS } from "./admin-nav-items";
import { useAdminSession, useLogout } from "@/hooks/use-admin-session";
import { setOnUnauthorized } from "@/lib/api/client";
import { Skeleton } from "@/components/ui/skeleton";

const LOGIN_PATH = "/admin/login";

/**
 * Owns the interactive sidebar shell — real identity, Sign Out, and the
 * global 401 -> /admin/login redirect (FR-401/402/403). Renders children
 * bare on the login route itself, which lives inside this same route group
 * but must not be wrapped in the authenticated sidebar (R7).
 */
export function AdminShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const queryClient = useQueryClient();
  const isLoginRoute = pathname === LOGIN_PATH;

  useEffect(() => {
    setOnUnauthorized(() => {
      queryClient.clear();
      // apiFetch fires this for EVERY 401 — including a wrong-password
      // response on the login form itself. The pathname guard is what keeps
      // a failed login from redirecting (and looping) back to this page; do
      // not remove it without giving the login call its own 401 path.
      if (window.location.pathname === LOGIN_PATH) return;
      router.replace(`${LOGIN_PATH}?from=${encodeURIComponent(window.location.pathname)}`);
    });
    return () => setOnUnauthorized(null);
  }, [queryClient, router]);

  const { data: session, isLoading } = useAdminSession({ enabled: !isLoginRoute });
  const logout = useLogout();

  if (isLoginRoute) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen flex bg-muted/20">
      <aside className="w-64 border-r bg-card flex flex-col shrink-0">
        <div className="h-16 border-b flex items-center justify-between px-6">
          <Link href="/admin/dashboard" className="flex items-center gap-2 font-bold text-lg text-primary">
            <Compass className="h-5 w-5" />
            <span>SUN Admin</span>
          </Link>
          <span className="text-[10px] font-bold bg-primary/10 text-primary px-2 py-0.5 rounded uppercase">
            v1.0
          </span>
        </div>

        <nav className="p-4 space-y-1.5 flex-1">
          {ADMIN_NAV_ITEMS.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                href={item.href}
                className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
              >
                <Icon className="h-4 w-4" />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t space-y-2">
          <Link
            href="/"
            target="_blank"
            className="flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground px-3 py-2 rounded-lg hover:bg-muted transition-colors"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            <span>Customer Website</span>
          </Link>

          <button
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
            className="w-full flex items-center gap-2 text-xs text-rose-600 hover:bg-rose-50 px-3 py-2 rounded-lg transition-colors disabled:opacity-50"
          >
            <LogOut className="h-3.5 w-3.5" />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-w-0">
        <header className="h-16 border-b bg-card flex items-center justify-between px-8 sticky top-0 z-10">
          <div className="text-sm font-medium text-muted-foreground">
            SUN Booking Tours Management Console
          </div>

          <div className="flex items-center gap-4">
            <button className="p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted relative">
              <Bell className="h-4 w-4" />
              <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-primary" />
            </button>

            <div className="flex items-center gap-2 pl-4 border-l">
              {isLoading || !session ? (
                <Skeleton className="h-8 w-32" />
              ) : (
                <>
                  <div className="h-8 w-8 rounded-full bg-primary/20 text-primary font-bold flex items-center justify-center text-xs">
                    {session.full_name.slice(0, 2).toUpperCase()}
                  </div>
                  <div className="text-xs text-left">
                    <span className="font-semibold block">{session.full_name}</span>
                    <span className="text-muted-foreground text-[11px]">{session.email}</span>
                  </div>
                </>
              )}
            </div>
          </div>
        </header>

        <main className="p-8 flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  );
}
