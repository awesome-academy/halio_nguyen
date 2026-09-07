import type { Metadata, ResolvingMetadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { MapPin, Calendar, Users, Star, CheckCircle, XCircle, ArrowRight, ShieldCheck } from "lucide-react";
import { formatCurrencyVND } from "@/lib/utils";

interface TourPageProps {
  params: {
    slug: string;
  };
}

// Sample tour lookup for SEO rendering
const SAMPLE_TOURS_DATA: Record<string, any> = {
  "phu-quoc-tropical-island-discovery-3d2n": {
    id: "44444444-4444-4444-4444-444444444001",
    title: "Phu Quoc Tropical Island Discovery 3D2N",
    slug: "phu-quoc-tropical-island-discovery-3d2n",
    destination: "Phu Quoc, Kien Giang, Vietnam",
    duration_days: 3,
    duration_nights: 2,
    price: 4500000,
    discount_price: 3990000,
    rating: 4.9,
    total_ratings: 128,
    category: "Island & Coastal Tours",
    thumbnail: "https://images.unsplash.com/photo-1540555700478-4be289fbecef?auto=format&fit=crop&w=1200&q=80",
    description: "Experience the ultimate island escape in Phu Quoc with SUN Booking Tours. Explore Grand World, snorkel vibrant coral reefs at May Rut Island, and savor fresh coastal seafood delicacies.",
    highlights: [
      "Full-day 4-island speedboat adventure with snorkeling",
      "World-record sea-crossing Hon Thom cable car ride",
      "Sunset dining and vibrant nightlife at Grand World",
      "Deluxe 4-star beachfront resort accommodation"
    ],
    inclusions: "All transfers, 4-star resort room, 5 meals, speedboat tour, snorkeling gear, tour guide, travel insurance.",
    exclusions: "Personal expenses, alcoholic beverages, round-trip flights to Phu Quoc.",
    itinerary: [
      { day: 1, title: "Airport Welcome & Grand World Sleepless City", detail: "Arrival at Phu Quoc airport, check in to resort, evening exploration of the Grand World canals and water show." },
      { day: 2, title: "4-Island Speedboat Odyssey & Coral Snorkeling", detail: "Cruise to Gam Ghi and May Rut islands. Snorkel amidst vibrant reefs and enjoy an island seafood BBQ lunch." },
      { day: 3, title: "Hon Thom Cable Car & Local Specialties", detail: "Ride the world-longest sea-crossing cable car, shop for local pearls and fish sauce, transfer to airport." }
    ],
    upcoming_schedules: [
      { id: "55555555-5555-5555-5555-555555555001", departure: "Next Friday", return: "Next Sunday", slots: 18, status: "open" },
      { id: "55555555-5555-5555-5555-555555555002", departure: "In 2 Weeks", return: "In 2 Weeks + 2 Days", slots: 25, status: "open" }
    ]
  },
  "misty-sapa-fansipan-conquest-2d1n": {
    id: "44444444-4444-4444-4444-444444444002",
    title: "Misty Sapa & Fansipan Peak Conquest 2D1N",
    slug: "misty-sapa-fansipan-conquest-2d1n",
    destination: "Sapa, Lao Cai, Vietnam",
    duration_days: 2,
    duration_nights: 1,
    price: 2800000,
    discount_price: 2490000,
    rating: 4.8,
    total_ratings: 94,
    category: "Mountain & Trekking",
    thumbnail: "https://images.unsplash.com/photo-1528127269322-539801943592?auto=format&fit=crop&w=1200&q=80",
    description: "Trek through picturesque valleys, visit authentic ethnic villages, and conquer the 3,143m rooftop of Indochina in Sapa with SUN Booking.",
    highlights: [
      "Reach the summit of Mount Fansipan (3,143m)",
      "Cultural trek through Cat Cat H'Mong village",
      "Sturgeon hotpot culinary experience",
      "Scenic mountain valley train journey"
    ],
    inclusions: "Hanoi-Sapa luxury limousine transfer, hotel, Fansipan cable car ticket, meals, English-speaking guide.",
    exclusions: "Tips for guide and driver, personal shopping.",
    itinerary: [
      { day: 1, title: "Hanoi to Sapa & Cat Cat Village", detail: "Scenic morning drive across the Northwest highway. Afternoon trek through rice terraces and waterfalls in Cat Cat village." },
      { day: 2, title: "Fansipan Peak Summit & Return to Hanoi", detail: "Ascend via Fansipan Cable car to 3,143m. Afternoon limousine ride back to Hanoi." }
    ],
    upcoming_schedules: [
      { id: "55555555-5555-5555-5555-555555555003", departure: "This Saturday", return: "This Sunday", slots: 12, status: "open" }
    ]
  }
};

// Next.js Dynamic SEO Metadata
export async function generateMetadata(
  { params }: TourPageProps,
  parent: ResolvingMetadata
): Promise<Metadata> {
  const tour = SAMPLE_TOURS_DATA[params.slug];
  if (!tour) {
    return { title: "Tour Not Found | SUN Booking Tours" };
  }

  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000";
  const canonicalUrl = `${siteUrl}/tours/${tour.slug}`;

  return {
    title: `${tour.title} - ${tour.destination}`,
    description: tour.description,
    keywords: [tour.destination, tour.category, "book tour", "vietnam vacation", "sun booking"],
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      title: tour.title,
      description: tour.description,
      url: canonicalUrl,
      siteName: "SUN Booking Tours",
      images: [
        {
          url: tour.thumbnail,
          width: 1200,
          height: 630,
          alt: tour.title,
        },
      ],
      type: "website",
    },
    twitter: {
      card: "summary_large_image",
      title: tour.title,
      description: tour.description,
      images: [tour.thumbnail],
    },
  };
}

export default function TourDetailPage({ params }: TourPageProps) {
  const tour = SAMPLE_TOURS_DATA[params.slug];
  if (!tour) {
    notFound();
  }

  // Schema.org JSON-LD Structured Data for Google Rich Snippets
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Product",
    "name": tour.title,
    "image": [tour.thumbnail],
    "description": tour.description,
    "brand": {
      "@type": "Brand",
      "name": "SUN Booking Tours"
    },
    "aggregateRating": {
      "@type": "AggregateRating",
      "ratingValue": tour.rating,
      "reviewCount": tour.total_ratings,
      "bestRating": 5,
      "worstRating": 1
    },
    "offers": {
      "@type": "Offer",
      "url": `${process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000"}/tours/${tour.slug}`,
      "priceCurrency": "VND",
      "price": tour.discount_price || tour.price,
      "availability": "https://schema.org/InStock",
      "validFrom": new Date().toISOString().split("T")[0]
    }
  };

  return (
    <article className="container mx-auto px-4 py-8 space-y-10 max-w-6xl">
      {/* Inject Structured Data (JSON-LD) */}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />

      {/* Breadcrumb Navigation */}
      <nav aria-label="Breadcrumb" className="text-xs text-muted-foreground flex items-center gap-2">
        <Link href="/" className="hover:underline">Home</Link>
        <span>/</span>
        <Link href="/tours" className="hover:underline">Tours</Link>
        <span>/</span>
        <span className="text-foreground font-medium truncate">{tour.title}</span>
      </nav>

      {/* Header Info */}
      <div className="space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <span className="bg-primary/10 text-primary text-xs font-semibold px-3 py-1 rounded-full">
            {tour.category}
          </span>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <MapPin className="h-3.5 w-3.5 text-primary" />
            <span>{tour.destination}</span>
          </div>
          <div className="flex items-center gap-1 text-xs text-amber-500 font-semibold">
            <Star className="h-3.5 w-3.5 fill-amber-400" />
            <span>{tour.rating}</span>
            <span className="text-muted-foreground">({tour.total_ratings} verified traveler reviews)</span>
          </div>
        </div>

        <h1 className="text-2xl sm:text-4xl font-extrabold tracking-tight leading-snug">
          {tour.title}
        </h1>
      </div>

      {/* Tour Gallery & Hero Image */}
      <div className="rounded-2xl overflow-hidden aspect-[16/9] max-h-[460px] w-full bg-muted">
        <img
          src={tour.thumbnail}
          alt={tour.title}
          className="w-full h-full object-cover"
        />
      </div>

      {/* Content Columns: Details + Sticky Booking Card */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-10">
        {/* Left: Description, Highlights, Itinerary, Inclusions */}
        <div className="lg:col-span-2 space-y-10">
          <section className="space-y-4">
            <h2 className="text-xl font-bold">Overview</h2>
            <p className="text-muted-foreground leading-relaxed">
              {tour.description}
            </p>
          </section>

          {/* Highlights */}
          <section className="space-y-4 bg-muted/30 p-6 rounded-2xl border">
            <h2 className="text-xl font-bold">Tour Highlights</h2>
            <ul className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {tour.highlights.map((item: string, idx: number) => (
                <li key={idx} className="flex items-start gap-2 text-sm">
                  <CheckCircle className="h-4 w-4 text-emerald-600 shrink-0 mt-0.5" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </section>

          {/* Detailed Itinerary */}
          <section className="space-y-6">
            <h2 className="text-xl font-bold">Day-by-Day Itinerary</h2>
            <div className="space-y-4">
              {tour.itinerary.map((dayPlan: any) => (
                <div key={dayPlan.day} className="border p-5 rounded-xl space-y-2 bg-card">
                  <div className="flex items-center gap-2">
                    <span className="bg-primary text-primary-foreground text-xs font-bold px-2 py-0.5 rounded">
                      Day {dayPlan.day}
                    </span>
                    <h3 className="font-bold text-base">{dayPlan.title}</h3>
                  </div>
                  <p className="text-sm text-muted-foreground leading-relaxed pl-12">
                    {dayPlan.detail}
                  </p>
                </div>
              ))}
            </div>
          </section>

          {/* Inclusions & Exclusions */}
          <section className="grid grid-cols-1 sm:grid-cols-2 gap-6 pt-4 border-t">
            <div className="space-y-2">
              <h3 className="font-bold text-sm text-emerald-700 flex items-center gap-1.5">
                <CheckCircle className="h-4 w-4" />
                <span>What&apos;s Included</span>
              </h3>
              <p className="text-xs text-muted-foreground leading-relaxed">{tour.inclusions}</p>
            </div>

            <div className="space-y-2">
              <h3 className="font-bold text-sm text-rose-700 flex items-center gap-1.5">
                <XCircle className="h-4 w-4" />
                <span>What&apos;s Excluded</span>
              </h3>
              <p className="text-xs text-muted-foreground leading-relaxed">{tour.exclusions}</p>
            </div>
          </section>
        </div>

        {/* Right: Booking Summary Card */}
        <aside className="lg:col-span-1">
          <div className="sticky top-24 border rounded-2xl p-6 bg-card shadow-lg space-y-6">
            <div>
              <span className="text-xs text-muted-foreground block">Price per adult</span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-2xl font-black text-primary">
                  {formatCurrencyVND(tour.discount_price || tour.price)}
                </span>
                {tour.discount_price && (
                  <span className="text-sm text-muted-foreground line-through">
                    {formatCurrencyVND(tour.price)}
                  </span>
                )}
              </div>
            </div>

            {/* Departure Schedules */}
            <div className="space-y-3 border-t pt-4">
              <label className="text-xs font-semibold uppercase text-muted-foreground block">
                Select Departure Date
              </label>
              <div className="space-y-2">
                {tour.upcoming_schedules.map((schedule: any) => (
                  <div 
                    key={schedule.id}
                    className="flex items-center justify-between p-3 border rounded-xl hover:border-primary cursor-pointer transition-colors"
                  >
                    <div>
                      <div className="text-xs font-semibold">{schedule.departure}</div>
                      <div className="text-[11px] text-muted-foreground">Return: {schedule.return}</div>
                    </div>
                    <span className="text-xs font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded">
                      {schedule.slots} slots left
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {/* Booking Action Button */}
            <Link
              href={`/booking/${tour.slug}`}
              className="w-full flex items-center justify-center gap-2 bg-primary text-primary-foreground font-bold py-3.5 px-4 rounded-xl hover:opacity-90 shadow-md transition-all"
            >
              <span>Book Tour Now</span>
              <ArrowRight className="h-4 w-4" />
            </Link>

            <div className="space-y-2 border-t pt-4 text-xs text-muted-foreground">
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 text-emerald-600 shrink-0" />
                <span>Free cancellation up to 7 days prior to departure</span>
              </div>
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 text-emerald-600 shrink-0" />
                <span>Instant booking confirmation via email</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </article>
  );
}
