import type { Metadata } from "next";
import Link from "next/link";
import { ThumbsUp, MessageSquare, Calendar, User, ArrowRight } from "lucide-react";
import { formatDate } from "@/lib/utils";

export const metadata: Metadata = {
  title: "Travel Guides, Food & News | SUN Booking Community",
  description: "Read genuine traveler reviews and local food guides for destinations across Vietnam and Asia.",
};

const SAMPLE_REVIEWS = [
  {
    id: "r1",
    title: "Top 10 Hidden Gem Seafood Eateries in Ham Ninh, Phu Quoc",
    slug: "top-10-hidden-gem-seafood-eateries-phu-quoc",
    category: "Food & Dining",
    category_slug: "food",
    author: "Nguyen Van Du Khach",
    created_at: "2026-09-01T10:00:00Z",
    thumbnail: "https://images.unsplash.com/photo-1544025162-d76694265947?auto=format&fit=crop&w=800&q=80",
    excerpt: "Discover the most authentic steamed sea crab, grilled sea urchins with scallion oil, and herring salad tucked away in the fishing village.",
    likes: 42,
    comments: 9,
  },
  {
    id: "r2",
    title: "Best Scenic Viewpoints Along Fansipan Mountain Trekking Trail",
    slug: "best-scenic-viewpoints-fansipan-trekking",
    category: "Places & Attractions",
    category_slug: "place",
    author: "Hoang Minh Anh",
    created_at: "2026-08-28T14:30:00Z",
    thumbnail: "https://images.unsplash.com/photo-1528127269322-539801943592?auto=format&fit=crop&w=800&q=80",
    excerpt: "An insider guide to capturing the sea of clouds at 2,800m station and safety advice for first-time high-altitude trekkers.",
    likes: 76,
    comments: 15,
  },
  {
    id: "r3",
    title: "Vietnam E-Visa Updates & Essential Autumn 2026 Travel Notice",
    slug: "vietnam-evisa-updates-autumn-2026",
    category: "Travel News",
    category_slug: "news",
    author: "Editorial Team",
    created_at: "2026-09-03T09:00:00Z",
    thumbnail: "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=800&q=80",
    excerpt: "Crucial immigration, visa extension policy updates, and domestic flight regulations for foreign visitors entering Vietnam.",
    likes: 31,
    comments: 4,
  },
];

export default function ReviewsPage() {
  return (
    <div className="container mx-auto px-4 py-12 max-w-6xl space-y-10">
      <div className="text-center max-w-2xl mx-auto space-y-3">
        <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight">
          Traveler Stories & Guides
        </h1>
        <p className="text-sm text-muted-foreground">
          Explore firsthand reviews on scenic places, culinary discoveries, and essential travel news.
        </p>

        {/* Category Pills */}
        <div className="flex flex-wrap justify-center gap-2 pt-4">
          <button className="px-4 py-1.5 rounded-full text-xs font-semibold bg-primary text-primary-foreground">
            All Topics
          </button>
          <button className="px-4 py-1.5 rounded-full text-xs font-semibold bg-muted hover:bg-muted/80 text-foreground transition-colors">
            Places & Attractions
          </button>
          <button className="px-4 py-1.5 rounded-full text-xs font-semibold bg-muted hover:bg-muted/80 text-foreground transition-colors">
            Food & Dining
          </button>
          <button className="px-4 py-1.5 rounded-full text-xs font-semibold bg-muted hover:bg-muted/80 text-foreground transition-colors">
            Travel News
          </button>
        </div>
      </div>

      {/* Reviews Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {SAMPLE_REVIEWS.map((review) => (
          <article key={review.id} className="border rounded-2xl overflow-hidden bg-card flex flex-col hover:shadow-lg transition-shadow">
            <div className="relative h-48 w-full bg-muted">
              <img
                src={review.thumbnail}
                alt={review.title}
                className="w-full h-full object-cover"
              />
              <span className="absolute top-3 left-3 bg-background/90 backdrop-blur-sm text-[11px] font-semibold px-2.5 py-0.5 rounded-full">
                {review.category}
              </span>
            </div>

            <div className="p-5 flex-1 flex flex-col justify-between space-y-4">
              <div className="space-y-2">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <User className="h-3.5 w-3.5" />
                  <span>{review.author}</span>
                  <span>•</span>
                  <Calendar className="h-3.5 w-3.5" />
                  <span>{formatDate(review.created_at)}</span>
                </div>

                <Link href={`/reviews/${review.slug}`}>
                  <h2 className="font-bold text-base hover:text-primary transition-colors line-clamp-2">
                    {review.title}
                  </h2>
                </Link>

                <p className="text-xs text-muted-foreground line-clamp-3 leading-relaxed">
                  {review.excerpt}
                </p>
              </div>

              <div className="border-t pt-3 flex items-center justify-between text-xs text-muted-foreground">
                <div className="flex items-center gap-3">
                  <span className="flex items-center gap-1 hover:text-primary cursor-pointer transition-colors">
                    <ThumbsUp className="h-3.5 w-3.5" />
                    <span>{review.likes}</span>
                  </span>
                  <span className="flex items-center gap-1">
                    <MessageSquare className="h-3.5 w-3.5" />
                    <span>{review.comments}</span>
                  </span>
                </div>

                <Link
                  href={`/reviews/${review.slug}`}
                  className="font-semibold text-primary flex items-center gap-1 hover:underline"
                >
                  <span>Read Article</span>
                  <ArrowRight className="h-3 w-3" />
                </Link>
              </div>
            </div>
          </article>
        ))}
      </div>
    </div>
  );
}
