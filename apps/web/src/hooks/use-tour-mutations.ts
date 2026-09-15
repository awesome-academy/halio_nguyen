"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createTour,
  createTourImage,
  createTourSchedule,
  deleteTour,
  deleteTourImage,
  deleteTourSchedule,
  updateTour,
  updateTourImage,
  updateTourSchedule,
  updateTourScheduleStatus,
  updateTourStatus,
} from "@/lib/api/tours";
import { tourKeys } from "@/lib/api/query-keys";
import type { TourCreatePayload, TourImagePayload, TourImageUpdatePayload, TourSchedulePayload, TourUpdatePayload } from "@/types/tour.types";

/**
 * One `useMutation` per A3-A13 action (step 3), each owning its own
 * invalidation so no component invents its own cache key. List mutations
 * invalidate `tourKeys.lists()`; every child (image/schedule) mutation
 * invalidates the parent tour's `tourKeys.detail(tourId)`.
 */

export function useCreateTour() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: TourCreatePayload) => createTour(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.lists() }),
  });
}

export function useUpdateTour(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: TourUpdatePayload) => updateTour(id, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tourKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: tourKeys.detail(id) });
    },
  });
}

/** Unbound (takes `id` per call) so both the list's row menu (any row) and
 * the form header (its own tour) can share one mutation definition. */
export function useUpdateTourStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => updateTourStatus(id, status),
    onSuccess: (_data, { id }) => {
      void queryClient.invalidateQueries({ queryKey: tourKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: tourKeys.detail(id) });
    },
  });
}

export function useDeleteTour() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteTour(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.lists() }),
  });
}

export function useCreateTourImage(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: TourImagePayload) => createTourImage(tourId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}

export function useUpdateTourImage(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ imageId, payload }: { imageId: string; payload: TourImageUpdatePayload }) => updateTourImage(tourId, imageId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}

export function useDeleteTourImage(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (imageId: string) => deleteTourImage(tourId, imageId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}

export function useCreateTourSchedule(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: TourSchedulePayload) => createTourSchedule(tourId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}

export function useUpdateTourSchedule(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ scheduleId, payload }: { scheduleId: string; payload: TourSchedulePayload }) => updateTourSchedule(tourId, scheduleId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}

export function useUpdateTourScheduleStatus(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ scheduleId, status }: { scheduleId: string; status: string }) => updateTourScheduleStatus(tourId, scheduleId, status),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}

export function useDeleteTourSchedule(tourId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (scheduleId: string) => deleteTourSchedule(tourId, scheduleId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tourKeys.detail(tourId) }),
  });
}
