import { TourFormPage } from "../_components/tour-form-page";

export const metadata = { title: "Edit Tour" };

interface EditTourPageProps {
  params: {
    id: string;
  };
}

export default function EditTourPage({ params }: EditTourPageProps) {
  return <TourFormPage mode="edit" id={params.id} />;
}
