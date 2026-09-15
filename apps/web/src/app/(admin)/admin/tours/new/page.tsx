import { TourFormPage } from "../_components/tour-form-page";

export const metadata = { title: "New Tour" };

export default function NewTourPage() {
  return <TourFormPage mode="create" />;
}
