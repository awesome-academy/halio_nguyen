"use client";

import { useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { Form } from "@/components/ui/form";
import { useForm, type Resolver } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/admin/page-header";
import { useCreateTour, useUpdateTour } from "@/hooks/use-tour-mutations";
import { listCategories, type CategoryListQuery } from "@/lib/api/categories";
import { getTour } from "@/lib/api/tours";
import { categoryKeys, tourKeys } from "@/lib/api/query-keys";
import { tourBaseSchema, tourCreateSchema, type TourCreateFormValues } from "@/lib/validation/tour-schema";
import type { TourCreatePayload, TourUpdatePayload } from "@/types/tour.types";
import { applyServerFieldErrors } from "./server-error-mapper";
import { TourFormBasicSection } from "./tour-form-basic-section";
import { TourFormContentSection } from "./tour-form-content-section";
import { TourFormPricingSection } from "./tour-form-pricing-section";
import { TourFormReadonlyStats } from "./tour-form-readonly-stats";
import { TourImagesEditor } from "./tour-images-editor";
import { TourSchedulesEditor } from "./tour-schedules-editor";
import { TourStatusActions } from "./tour-status-actions";

const EMPTY_VALUES: TourCreateFormValues = {
  category_id: "",
  title: "",
  destination: "",
  description: "",
  itinerary: "",
  duration_days: 1,
  duration_nights: 0,
  price: 0,
  max_participants: 1,
  thumbnail_url: "",
  highlights: [],
  inclusions: "",
  exclusions: "",
  images: [],
  schedules: [],
};

interface TourFormPageProps {
  mode: "create" | "edit";
  id?: string;
}

/** Host: owns the useForm instance and passes it down via FormProvider (step
 * 6). The mode drives the submit path (A3 vs A4) and whether the array
 * editors are RHF-backed or API-backed — kept as two named branches (R2). */
export function TourFormPage({ mode, id }: TourFormPageProps) {
  const router = useRouter();

  const tourQuery = useQuery({ queryKey: tourKeys.detail(id ?? ""), queryFn: () => getTour(id as string), enabled: mode === "edit" && Boolean(id) });
  const categoryListQuery: CategoryListQuery = { page: 1, page_size: 100, is_active: mode === "create" ? true : undefined };
  const categoryQuery = useQuery({
    queryKey: categoryKeys.list(categoryListQuery),
    queryFn: () => listCategories(categoryListQuery),
  });
  // R4: edit mode also offers the tour's current category even if it has since been deactivated.
  const categories = useMemo(() => {
    const items = categoryQuery.data?.items ?? [];
    if (mode === "create") return items;
    const currentId = tourQuery.data?.category_id;
    return items.filter((c) => c.is_active || c.id === currentId);
  }, [categoryQuery.data, tourQuery.data, mode]);

  const form = useForm<TourCreateFormValues>({
    // Create validates the whole payload (base + arrays, R2); edit validates
    // only the base fields — array rows are validated per-row on their own
    // save. The two schemas' inferred types only differ by images/schedules
    // (absent on the edit resolver, which never touches those form keys), so
    // this cast through `unknown` is safe for what the form actually submits.
    resolver: zodResolver(mode === "create" ? tourCreateSchema : tourBaseSchema) as unknown as Resolver<TourCreateFormValues>,
    defaultValues: EMPTY_VALUES,
  });

  useEffect(() => {
    const tour = tourQuery.data;
    if (!tour) return;
    form.reset({
      category_id: tour.category_id,
      title: tour.title,
      destination: tour.destination,
      description: tour.description,
      itinerary: tour.itinerary ?? "",
      duration_days: tour.duration_days,
      duration_nights: tour.duration_nights,
      price: tour.price,
      discount_price: tour.discount_price,
      max_participants: tour.max_participants,
      thumbnail_url: tour.thumbnail_url ?? "",
      highlights: (tour.highlights ?? []).map((value) => ({ value })),
      inclusions: tour.inclusions ?? "",
      exclusions: tour.exclusions ?? "",
      images: tour.images.map((img) => ({ id: img.id, image_url: img.image_url, caption: img.caption ?? "", sort_order: img.sort_order })),
      // Schedule dates arrive as full RFC3339 timestamps (Go's default
      // time.Time JSON encoding) but <input type="date"> needs YYYY-MM-DD —
      // slicing the first 10 chars is exact since these are DATE columns
      // scanned into midnight-UTC time.Time, no timezone math needed.
      schedules: tour.schedules.map((s) => ({ id: s.id, departure_date: s.departure_date.slice(0, 10), return_date: s.return_date.slice(0, 10), available_slots: s.available_slots, price_override: s.price_override, status: s.status })),
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps -- form.reset identity is stable; only re-run when the fetched tour changes
  }, [tourQuery.data]);

  const createTour = useCreateTour();
  const updateTour = useUpdateTour(id ?? "");
  const busy = createTour.isPending || updateTour.isPending;

  async function onSubmit(values: TourCreateFormValues) {
    const base: TourUpdatePayload = {
      category_id: values.category_id,
      title: values.title,
      description: values.description,
      itinerary: values.itinerary || undefined,
      destination: values.destination,
      duration_days: values.duration_days,
      duration_nights: values.duration_nights,
      price: values.price,
      discount_price: values.discount_price,
      max_participants: values.max_participants,
      thumbnail_url: values.thumbnail_url || undefined,
      highlights: values.highlights.map((h) => h.value.trim()).filter(Boolean),
      inclusions: values.inclusions || undefined,
      exclusions: values.exclusions || undefined,
    };

    if (mode === "create") {
      const payload: TourCreatePayload = {
        ...base,
        images: values.images.map((img) => ({ image_url: img.image_url, caption: img.caption || undefined })),
        schedules: values.schedules.map((s) => ({ departure_date: s.departure_date, return_date: s.return_date, available_slots: s.available_slots, price_override: s.price_override })),
      };
      try {
        const created = await createTour.mutateAsync(payload);
        toast.success("Tour created");
        router.replace(`/admin/tours/${created.id}`);
      } catch (err) {
        applyServerFieldErrors(err, form.setError);
      }
      return;
    }

    try {
      await updateTour.mutateAsync(base);
      toast.success("Tour updated");
    } catch (err) {
      applyServerFieldErrors(err, form.setError);
    }
  }

  if (mode === "edit" && tourQuery.isLoading) {
    return <div className="flex items-center justify-center py-24 text-muted-foreground"><Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading tour…</div>;
  }
  if (mode === "edit" && tourQuery.isError) {
    return <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm">Could not load this tour.</div>;
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <PageHeader title={mode === "create" ? "New tour" : tourQuery.data?.title ?? "Edit tour"} description={mode === "create" ? "Create a tour with its gallery and schedules in one step." : "Base fields save separately from the gallery and schedules."}>
          {mode === "edit" && id && tourQuery.data && <TourStatusActions tourId={id} status={tourQuery.data.status} />}
          <Button type="submit" disabled={busy}>
            {busy && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            {mode === "create" ? "Create tour" : "Save changes"}
          </Button>
        </PageHeader>

        {mode === "edit" && tourQuery.data && <TourFormReadonlyStats avgRating={tourQuery.data.avg_rating} totalRatings={tourQuery.data.total_ratings} />}

        <TourFormBasicSection categories={categories} categoriesLoading={categoryQuery.isLoading} />
        <TourFormPricingSection />
        <TourFormContentSection />
        <TourImagesEditor mode={mode} tourId={id} />
        <TourSchedulesEditor mode={mode} tourId={id} />
      </form>
    </Form>
  );
}
