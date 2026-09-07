import Link from "next/link";
import { Compass, MapPin, Calendar, Users, Star, ArrowRight, ShieldCheck, HeartHandshake, Award } from "lucide-react";
import { formatCurrencyVND } from "@/lib/utils";

// Mock data representing tours for SSR landing page
const FEATURED_TOURS = [
  {
    id: "44444444-4444-4444-4444-444444444001",
    title: "Phu Quoc Tropical Island Discovery 3D2N",
    slug: "phu-quoc-tropical-island-discovery-3d2n",
    destination: "Phu Quoc, Kien Giang",
    duration: "3 Days 2 Nights",
    price: 4500000,
    discount_price: 3990000,
    rating: 4.9,
    total_ratings: 128,
    category: "Island & Coastal",
    thumbnail: "https://images.unsplash.com/photo-1540555700478-4be289fbecef?auto=format&fit=crop&w=800&q=80",
  },
  {
    id: "44444444-4444-4444-4444-444444444002",
    title: "Misty Sapa & Fansipan Peak Conquest 2D1N",
    slug: "misty-sapa-fansipan-conquest-2d1n",
    destination: "Sapa, Lao Cai",
    duration: "2 Days 1 Night",
    price: 2800000,
    discount_price: 2490000,
    rating: 4.8,
    total_ratings: 94,
    category: "Mountain & Trekking",
    thumbnail: "https://images.unsplash.com/photo-1528127269322-539801943592?auto=format&fit=crop&w=800&q=80",
  },
];

const CATEGORIES = [
  { name: "Island & Coastal", slug: "island-coastal", count: "12+ tours", icon: "🏝️" },
  { name: "Mountain & Trekking", slug: "mountain-trekking", count: "8+ tours", icon: "⛰️" },
  { name: "Culture & Heritage", slug: "culture-heritage", count: "15+ tours", icon: "🏛️" },
  { name: "Resort & Relaxation", slug: "resort-relaxation", count: "6+ tours", icon: "💆" },
];

export default function LandingPage() {
  return (
    <div className="flex flex-col gap-16 pb-16">
      {/* 1. Hero Section */}
      <section className="relative min-h-[560px] flex items-center justify-center bg-gradient-to-r from-slate-900/80 to-slate-900/40 text-white">
        <div 
          className="absolute inset-0 -z-10 bg-cover bg-center" 
          style={{ backgroundImage: `url('https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=2000&q=80')` }}
        />
        <div className="absolute inset-0 -z-10 bg-black/40 backdrop-blur-[1px]" />

        <div className="container mx-auto px-4 py-16 text-center max-w-4xl space-y-6">
          <div className="inline-flex items-center gap-2 rounded-full bg-primary/20 backdrop-blur-md px-3 py-1 text-xs font-semibold text-primary-foreground border border-primary/40">
            <Compass className="h-3.5 w-3.5" />
            <span>Discover Vietnam & Southeast Asia with SUN Booking</span>
          </div>

          <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight leading-tight">
            Unforgettable Journeys, <br />
            <span className="text-primary">Instant Booking Guaranteed</span>
          </h1>

          <p className="text-lg text-slate-200 max-w-2xl mx-auto">
            Book top-rated guided tours, vacation packages, and authentic cultural activities with secure Internet Banking payment and 24/7 travel assistance.
          </p>

          {/* Quick Search Widget */}
          <div className="bg-background text-foreground p-3 rounded-2xl shadow-2xl border max-w-3xl mx-auto mt-8 grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="flex items-center gap-3 px-3 py-2 border rounded-xl hover:border-primary transition-colors">
              <MapPin className="h-5 w-5 text-primary shrink-0" />
              <div className="text-left">
                <span className="block text-xs font-semibold text-muted-foreground uppercase">Destination</span>
                <input 
                  type="text" 
                  placeholder="Where to? (e.g. Phu Quoc)" 
                  className="w-full text-sm font-medium bg-transparent focus:outline-none"
                />
              </div>
            </div>

            <div className="flex items-center gap-3 px-3 py-2 border rounded-xl hover:border-primary transition-colors">
              <Calendar className="h-5 w-5 text-primary shrink-0" />
              <div className="text-left">
                <span className="block text-xs font-semibold text-muted-foreground uppercase">Departure</span>
                <input 
                  type="date" 
                  className="w-full text-sm font-medium bg-transparent focus:outline-none"
                />
              </div>
            </div>

            <Link
              href="/tours"
              className="flex items-center justify-center gap-2 bg-primary text-primary-foreground font-semibold rounded-xl py-3 hover:opacity-90 transition-opacity"
            >
              <span>Search Tours</span>
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* 2. Tour Categories */}
      <section className="container mx-auto px-4">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Popular Travel Categories</h2>
            <p className="text-sm text-muted-foreground">Find curated adventures matched to your travel style</p>
          </div>
          <Link href="/tours" className="text-sm font-semibold text-primary hover:underline flex items-center gap-1">
            <span>View All</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {CATEGORIES.map((category) => (
            <Link
              key={category.slug}
              href={`/tours?category=${category.slug}`}
              className="group p-5 rounded-xl border bg-card hover:border-primary hover:shadow-md transition-all flex items-center gap-4"
            >
              <div className="text-3xl p-3 bg-muted rounded-xl group-hover:scale-110 transition-transform">
                {category.icon}
              </div>
              <div>
                <h3 className="font-semibold text-sm group-hover:text-primary transition-colors">{category.name}</h3>
                <span className="text-xs text-muted-foreground">{category.count}</span>
              </div>
            </Link>
          ))}
        </div>
      </section>

      {/* 3. Featured Tours */}
      <section className="container mx-auto px-4">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Handpicked Featured Tours</h2>
            <p className="text-sm text-muted-foreground">Top-rated vacation packages loved by thousands of travelers</p>
          </div>
          <Link href="/tours" className="text-sm font-semibold text-primary hover:underline flex items-center gap-1">
            <span>Explore All Tours</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {FEATURED_TOURS.map((tour) => (
            <div key={tour.id} className="group rounded-2xl border bg-card overflow-hidden hover:shadow-xl transition-all flex flex-col">
              <div className="relative h-56 w-full overflow-hidden bg-muted">
                <img
                  src={tour.thumbnail}
                  alt={tour.title}
                  className="h-full w-full object-cover group-hover:scale-105 transition-transform duration-300"
                />
                <span className="absolute top-3 left-3 bg-background/90 backdrop-blur-sm text-xs font-semibold px-2.5 py-1 rounded-full shadow-sm">
                  {tour.category}
                </span>
              </div>

              <div className="p-5 flex-1 flex flex-col justify-between space-y-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span className="flex items-center gap-1 font-medium text-foreground">
                      <MapPin className="h-3.5 w-3.5 text-primary" />
                      {tour.destination}
                    </span>
                    <span>{tour.duration}</span>
                  </div>

                  <Link href={`/tours/${tour.slug}`}>
                    <h3 className="font-bold text-base group-hover:text-primary transition-colors line-clamp-2">
                      {tour.title}
                    </h3>
                  </Link>

                  <div className="flex items-center gap-1 text-xs">
                    <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400" />
                    <span className="font-semibold">{tour.rating}</span>
                    <span className="text-muted-foreground">({tour.total_ratings} verified reviews)</span>
                  </div>
                </div>

                <div className="border-t pt-4 flex items-center justify-between">
                  <div>
                    <span className="text-xs text-muted-foreground block">Starting from</span>
                    <div className="flex items-baseline gap-2">
                      <span className="text-lg font-extrabold text-primary">
                        {formatCurrencyVND(tour.discount_price || tour.price)}
                      </span>
                      {tour.discount_price && (
                        <span className="text-xs text-muted-foreground line-through">
                          {formatCurrencyVND(tour.price)}
                        </span>
                      )}
                    </div>
                  </div>

                  <Link
                    href={`/tours/${tour.slug}`}
                    className="text-xs font-semibold px-4 py-2 bg-primary/10 text-primary hover:bg-primary hover:text-primary-foreground rounded-lg transition-colors"
                  >
                    View Details
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* 4. Trust & Value Proposition */}
      <section className="bg-muted/40 border-y py-16">
        <div className="container mx-auto px-4">
          <div className="text-center max-w-2xl mx-auto mb-12">
            <h2 className="text-2xl font-bold tracking-tight">Why Book With SUN Booking Tours?</h2>
            <p className="text-sm text-muted-foreground mt-2">
              We ensure seamless booking, genuine traveler reviews, and transparent pricing.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
            <div className="p-6 bg-card rounded-2xl border shadow-sm space-y-3">
              <div className="h-12 w-12 rounded-xl bg-primary/10 text-primary flex items-center justify-center mx-auto">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <h3 className="font-bold text-base">Verified Travelers Only</h3>
              <p className="text-sm text-muted-foreground leading-relaxed">
                Tour ratings and feedback are strictly verified through completed booking records, ensuring 100% genuine insights.
              </p>
            </div>

            <div className="p-6 bg-card rounded-2xl border shadow-sm space-y-3">
              <div className="h-12 w-12 rounded-xl bg-primary/10 text-primary flex items-center justify-center mx-auto">
                <HeartHandshake className="h-6 w-6" />
              </div>
              <h3 className="font-bold text-base">Internet Banking & QR Pay</h3>
              <p className="text-sm text-muted-foreground leading-relaxed">
                Pay instantly via your favorite banking apps (Vietcombank, Techcombank, MB Bank, VNPay) with 256-bit bank data encryption.
              </p>
            </div>

            <div className="p-6 bg-card rounded-2xl border shadow-sm space-y-3">
              <div className="h-12 w-12 rounded-xl bg-primary/10 text-primary flex items-center justify-center mx-auto">
                <Award className="h-6 w-6" />
              </div>
              <h3 className="font-bold text-base">Best Price Guarantee</h3>
              <p className="text-sm text-muted-foreground leading-relaxed">
                Direct partnerships with licensed tour operators provide exclusive packages, flexible cancellation, and 24/7 hotline support.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
