interface TourFormReadonlyStatsProps {
  avgRating: number;
  totalRatings: number;
}

/** BR-005: text only — no <input>/<textarea> round-trips avg_rating or
 * total_ratings, so there is nothing here for a 422 to ever reject. */
export function TourFormReadonlyStats({ avgRating, totalRatings }: TourFormReadonlyStatsProps) {
  return (
    <div className="rounded-lg border bg-card p-4 flex items-center gap-6 text-sm">
      <div>
        <span className="text-muted-foreground">Average rating: </span>
        <span className="font-medium">{avgRating.toFixed(1)} / 5</span>
      </div>
      <div>
        <span className="text-muted-foreground">Total ratings: </span>
        <span className="font-medium">{totalRatings}</span>
      </div>
    </div>
  );
}
