import { ReviewDetailClient } from "../_components/review-detail-client";

export const metadata = { title: "Review Detail" };

interface ReviewDetailPageProps {
  params: {
    id: string;
  };
}

export default function ReviewDetailPage({ params }: ReviewDetailPageProps) {
  return <ReviewDetailClient id={params.id} />;
}
