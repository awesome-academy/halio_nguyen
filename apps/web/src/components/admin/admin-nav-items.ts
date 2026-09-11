import { LayoutDashboard, Map, Tag, CalendarCheck, MessageSquare, Users, BarChart3, type LucideIcon } from "lucide-react";

export interface AdminNavItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

export const ADMIN_NAV_ITEMS: AdminNavItem[] = [
  { label: "Dashboard", href: "/admin/dashboard", icon: LayoutDashboard },
  { label: "Tour Packages", href: "/admin/tours", icon: Map },
  { label: "Tour Categories", href: "/admin/categories", icon: Tag },
  { label: "Booking Requests", href: "/admin/bookings", icon: CalendarCheck },
  { label: "Reviews & Comments", href: "/admin/reviews", icon: MessageSquare },
  { label: "User Management", href: "/admin/users", icon: Users },
  { label: "Revenue Analytics", href: "/admin/revenue", icon: BarChart3 },
];
