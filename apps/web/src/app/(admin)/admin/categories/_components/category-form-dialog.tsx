"use client";

import { useEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api/types";
import { categorySchema, slugify, type CategoryFormValues } from "@/lib/validation/category-schema";
import type { CategoryListItem } from "@/types/category.types";

interface CategoryFormDialogProps {
  /** "create" for a fresh form, a category to edit, or null when closed. */
  target: "create" | CategoryListItem | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (values: CategoryFormValues) => Promise<void>;
}

const EMPTY: CategoryFormValues = { name: "", slug: "", description: "", image_url: "", sort_order: undefined, is_active: true };

/** One dialog for create and edit (FR-203/FR-204 share the shape). */
export function CategoryFormDialog({ target, onOpenChange, onSubmit }: CategoryFormDialogProps) {
  const open = target !== null;
  const editing = target !== null && target !== "create" ? target : null;
  // FR-401: slug follows name until the admin edits the slug field; an existing
  // category counts as already-touched so edits never overwrite its slug.
  const slugTouched = useRef(false);

  const form = useForm<CategoryFormValues>({ resolver: zodResolver(categorySchema), defaultValues: EMPTY });

  useEffect(() => {
    if (!open) return;
    slugTouched.current = editing !== null;
    form.reset(
      editing
        ? {
            name: editing.name,
            slug: editing.slug,
            description: editing.description ?? "",
            image_url: editing.image_url ?? "",
            sort_order: editing.sort_order,
            is_active: editing.is_active,
          }
        : EMPTY,
    );
  }, [open, editing, form]);

  const name = form.watch("name");
  useEffect(() => {
    if (open && !slugTouched.current) form.setValue("slug", slugify(name), { shouldValidate: form.formState.isSubmitted });
  }, [name, open, form]);

  async function handleSubmit(values: CategoryFormValues) {
    try {
      await onSubmit(values);
      toast.success(editing ? "Category updated" : "Category created");
      onOpenChange(false);
    } catch (err) {
      if (err instanceof ApiError && err.fields) {
        for (const [field, message] of Object.entries(err.fields)) {
          form.setError(field as keyof CategoryFormValues, { type: "server", message });
        }
        return;
      }
      toast.error(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
    }
  }

  const busy = form.formState.isSubmitting;

  return (
    <Dialog open={open} onOpenChange={(o) => !busy && onOpenChange(o)}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit category" : "New category"}</DialogTitle>
          <DialogDescription>Tour categories group packages on the customer site.</DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4" noValidate>
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem><FormLabel>Name</FormLabel><FormControl><Input autoFocus {...field} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name="slug" render={({ field }) => (
              <FormItem><FormLabel>Slug</FormLabel>
                <FormControl><Input {...field} onChange={(e) => { slugTouched.current = true; field.onChange(e); }} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="description" render={({ field }) => (
              <FormItem><FormLabel>Description</FormLabel><FormControl><Textarea rows={3} {...field} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name="image_url" render={({ field }) => (
              <FormItem><FormLabel>Image URL</FormLabel><FormControl><Input type="url" placeholder="https://…" {...field} /></FormControl><FormMessage /></FormItem>
            )} />
            <div className="grid grid-cols-2 gap-4">
              <FormField control={form.control} name="sort_order" render={({ field }) => (
                <FormItem><FormLabel>Position (0 = first)</FormLabel>
                  <FormControl><Input type="number" min={0} placeholder="end of list" {...field} value={field.value ?? ""} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="is_active" render={({ field }) => (
                <FormItem className="flex flex-col"><FormLabel>Active</FormLabel>
                  <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} aria-label="Active" /></FormControl>
                </FormItem>
              )} />
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={busy}>Cancel</Button>
              <Button type="submit" disabled={busy}>
                {busy && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                Save
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
