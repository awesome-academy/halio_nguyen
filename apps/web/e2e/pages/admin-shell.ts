import type { Locator, Page } from "@playwright/test";

/** Mirrors `admin-nav-items.ts` — labels and paths are read from source, not guessed. */
export const ADMIN_NAV = [
  { label: "Dashboard", path: "/admin/dashboard" },
  { label: "Tour Packages", path: "/admin/tours" },
  { label: "Tour Categories", path: "/admin/categories" },
  { label: "Booking Requests", path: "/admin/bookings" },
  { label: "Reviews & Comments", path: "/admin/reviews" },
  { label: "User Management", path: "/admin/users" },
  { label: "Revenue Analytics", path: "/admin/revenue" },
] as const;

/** The persistent sidebar/header chrome (`admin-shell.tsx`) around every authenticated screen. */
export class AdminShellPage {
  constructor(private readonly page: Page) {}

  navLink(label: string): Locator {
    return this.page.getByRole("link", { name: label, exact: true });
  }

  get brand(): Locator {
    return this.page.getByRole("link", { name: "SUN Admin" });
  }

  get customerWebsite(): Locator {
    return this.page.getByRole("link", { name: "Customer Website" });
  }

  get signOut(): Locator {
    return this.page.getByRole("button", { name: "Sign Out" });
  }

  /** The full_name / email block in the header — matched by substring since it's plain text, not a label. */
  identity(text: string): Locator {
    return this.page.getByText(text, { exact: false });
  }
}
