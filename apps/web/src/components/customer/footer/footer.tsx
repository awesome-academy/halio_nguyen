import Link from "next/link";
import { Compass, Mail, Phone, MapPin, ShieldCheck, CreditCard } from "lucide-react";

export function CustomerFooter() {
  return (
    <footer className="w-full border-t bg-muted/40 mt-auto">
      <div className="container mx-auto px-4 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Company Info */}
          <div className="space-y-3">
            <Link href="/" className="flex items-center gap-2 font-bold text-xl text-primary">
              <Compass className="h-6 w-6" />
              <span>SUN Booking</span>
            </Link>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Your trusted partner for extraordinary travel and tour experiences. Discover breathtaking natural wonders, rich culture, and authentic local cuisines.
            </p>
            <div className="flex items-center gap-2 text-xs text-muted-foreground pt-2">
              <ShieldCheck className="h-4 w-4 text-emerald-600" />
              <span>Licensed International Tour Operator</span>
            </div>
          </div>

          {/* Quick Links */}
          <div className="space-y-3">
            <h4 className="font-semibold text-sm">Top Tour Categories</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <Link href="/tours?category=island-coastal" className="hover:text-primary transition-colors">
                  Island & Beach Vacations
                </Link>
              </li>
              <li>
                <Link href="/tours?category=mountain-trekking" className="hover:text-primary transition-colors">
                  Mountain Trekking & Highlands
                </Link>
              </li>
              <li>
                <Link href="/tours?category=culture-heritage" className="hover:text-primary transition-colors">
                  Culture & Ancient Di San
                </Link>
              </li>
              <li>
                <Link href="/tours?category=resort-relaxation" className="hover:text-primary transition-colors">
                  Luxury Wellness & Resorts
                </Link>
              </li>
            </ul>
          </div>

          {/* Community & Reviews */}
          <div className="space-y-3">
            <h4 className="font-semibold text-sm">Travel Guides & Reviews</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <Link href="/reviews?category=place" className="hover:text-primary transition-colors">
                  Scenic Places & Viewpoints
                </Link>
              </li>
              <li>
                <Link href="/reviews?category=food" className="hover:text-primary transition-colors">
                  Vietnamese & Local Cuisine
                </Link>
              </li>
              <li>
                <Link href="/reviews?category=news" className="hover:text-primary transition-colors">
                  Latest Travel Trends & Tips
                </Link>
              </li>
              <li>
                <Link href="/admin" className="hover:text-primary transition-colors text-xs text-muted-foreground/60">
                  Staff Portal
                </Link>
              </li>
            </ul>
          </div>

          {/* Contact & Payment */}
          <div className="space-y-3">
            <h4 className="font-semibold text-sm">Support & Payment</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-center gap-2">
                <Phone className="h-4 w-4 text-primary" />
                <span>+84 (0) 901 234 567 (24/7)</span>
              </li>
              <li className="flex items-center gap-2">
                <Mail className="h-4 w-4 text-primary" />
                <span>support@sunbooking.com</span>
              </li>
              <li className="flex items-center gap-2">
                <MapPin className="h-4 w-4 text-primary" />
                <span>Hanoi & Ho Chi Minh City, Vietnam</span>
              </li>
            </ul>
            <div className="pt-2">
              <div className="flex items-center gap-2 text-xs text-muted-foreground mb-1">
                <CreditCard className="h-3.5 w-3.5" />
                <span>Accepted Payment Methods</span>
              </div>
              <p className="text-xs text-muted-foreground/80">
                Internet Banking (Vietcombank, Techcombank, MB Bank), VNPay QR, Bank Transfer
              </p>
            </div>
          </div>
        </div>

        <div className="border-t mt-8 pt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-muted-foreground">
          <p>© {new Date().getFullYear()} SUN Booking Tours. All rights reserved.</p>
          <div className="flex gap-4">
            <Link href="/terms" className="hover:underline">Terms of Service</Link>
            <Link href="/privacy" className="hover:underline">Privacy Policy</Link>
            <Link href="/refund" className="hover:underline">Cancellation & Refund</Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
