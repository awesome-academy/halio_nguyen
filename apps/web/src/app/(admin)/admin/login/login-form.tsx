"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { login } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/types";
import { loginSchema, type LoginFormValues } from "@/lib/validation/auth-schema";
import { adminSessionKey } from "@/hooks/use-admin-session";

export function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const [formError, setFormError] = useState<string | null>(null);

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  async function onSubmit(values: LoginFormValues) {
    setFormError(null);
    try {
      const from = searchParams.get("from") ?? undefined;
      const { user, redirect_to } = await login(values.email, values.password, from);
      queryClient.setQueryData(adminSessionKey, user);
      router.replace(redirect_to);
    } catch (err) {
      // One form-level message regardless of cause (never field-level —
      // that would leak which field failed, BR-001).
      const message = err instanceof ApiError ? err.message : "Something went wrong. Please try again.";
      setFormError(message);
      form.resetField("password");
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {formError && (
          <div role="alert" aria-live="assertive" className="text-sm text-destructive bg-destructive/10 rounded-md px-3 py-2">
            {formError}
          </div>
        )}

        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input type="email" autoFocus autoComplete="username" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input type="password" autoComplete="current-password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
          Sign In
        </Button>
      </form>
    </Form>
  );
}
