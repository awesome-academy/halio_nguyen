import { BookingDetailClient } from "../_components/booking-detail-client";

export const metadata = { title: "Booking Detail" };

interface BookingDetailPageProps {
  params: {
    id: string;
  };
}

export default function BookingDetailPage({ params }: BookingDetailPageProps) {
  return <BookingDetailClient id={params.id} />;
}
