import { UserDetailClient } from "../_components/user-detail-client";

export const metadata = { title: "User Detail" };

interface UserDetailPageProps {
  params: {
    id: string;
  };
}

export default function UserDetailPage({ params }: UserDetailPageProps) {
  return <UserDetailClient id={params.id} />;
}
