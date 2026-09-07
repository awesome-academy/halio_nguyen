import Link from "next/link";
import { 
  LayoutDashboard, 
  Map, 
  Tag, 
  CalendarCheck, 
  MessageSquare, 
  Users, 
  BarChart3, 
  LogOut, 
  Compass, 
  ExternalLink,
  Bell
} from "lucide-react";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const NAV_ITEMS = [
    { label: "Dashboard", href: "/admin/dashboard", icon: LayoutDashboard },
    { label: "Tour Packages", href: "/admin/tours", icon: Map },
    { label: "Tour Categories", href: "/admin/categories", icon: Tag },
    { label: "Booking Requests", href: "/admin/bookings", icon: CalendarCheck },
    { label: "Reviews & Comments", href: "/admin/reviews", icon: MessageSquare },
    { label: "User Management", href: "/admin/users", icon: Users },
    { label: "Revenue Analytics", href: "/admin/revenue", icon: BarChart3 },
  ];

  return (
    <div className="min-h-screen flex bg-muted/20">
      {/* Admin Sidebar */}
      <aside className="w-64 border-r bg-card flex flex-col shrink-0">
        {/* Admin Brand */}
        <div className="h-16 border-b flex items-center justify-between px-6">
          <Link href="/admin/dashboard" className="flex items-center gap-2 font-bold text-lg text-primary">
            <Compass className="h-5 w-5" />
            <span>SUN Admin</span>
          </Link>
          <span className="text-[10px] font-bold bg-primary/10 text-primary px-2 py-0.5 rounded uppercase">
            v1.0
          </span>
        </div>

        {/* Navigation Items */}
        <nav className="p-4 space-y-1.5 flex-1">
          {NAV_ITEMS.map((item) => {
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

        {/* Bottom Sidebar Links */}
        <div className="p-4 border-t space-y-2">
          <Link
            href="/"
            target="_blank"
            className="flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground px-3 py-2 rounded-lg hover:bg-muted transition-colors"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            <span>Customer Website</span>
          </Link>

          <button className="w-full flex items-center gap-2 text-xs text-rose-600 hover:bg-rose-50 px-3 py-2 rounded-lg transition-colors">
            <LogOut className="h-3.5 w-3.5" />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* Main Admin Area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top Header */}
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
              <div className="h-8 w-8 rounded-full bg-primary/20 text-primary font-bold flex items-center justify-center text-xs">
                AD
              </div>
              <div className="text-xs text-left">
                <span className="font-semibold block">Administrator</span>
                <span className="text-muted-foreground text-[11px]">admin@sunbooking.com</span>
              </div>
            </div>
          </div>
        </header>

        {/* Content Body */}
        <main className="p-8 flex-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  );
}
