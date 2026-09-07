import Link from "next/link";
import { Compass, Search, User as UserIcon, Calendar, Heart } from "lucide-react";

export function CustomerNavbar() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        {/* Brand Logo */}
        <Link href="/" className="flex items-center gap-2 font-bold text-xl text-primary">
          <Compass className="h-6 w-6 text-primary" />
          <span>SUN Booking</span>
        </Link>

        {/* Navigation Links */}
        <nav className="hidden md:flex items-center gap-6 text-sm font-medium">
          <Link href="/tours" className="transition-colors hover:text-primary">
            Explore Tours
          </Link>
          <Link href="/tours?category=island-coastal" className="transition-colors hover:text-primary">
            Island & Beach
          </Link>
          <Link href="/tours?category=mountain-trekking" className="transition-colors hover:text-primary">
            Mountain & Trekking
          </Link>
          <Link href="/reviews" className="transition-colors hover:text-primary">
            Travel Guide & Reviews
          </Link>
        </nav>

        {/* Action Buttons */}
        <div className="flex items-center gap-3">
          <Link
            href="/tours"
            className="p-2 text-muted-foreground hover:text-foreground transition-colors"
            title="Search Tours"
          >
            <Search className="h-5 w-5" />
          </Link>

          <Link
            href="/user/bookings"
            className="hidden sm:flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-full border hover:bg-muted transition-colors"
          >
            <Calendar className="h-3.5 w-3.5" />
            <span>My Bookings</span>
          </Link>

          <Link
            href="/login"
            className="flex items-center gap-1.5 bg-primary text-primary-foreground text-sm font-medium px-4 py-2 rounded-lg hover:opacity-90 transition-opacity"
          >
            <UserIcon className="h-4 w-4" />
            <span>Sign In</span>
          </Link>
        </div>
      </div>
    </header>
  );
}
