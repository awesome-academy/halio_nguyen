"use client";

import { formatCurrencyVND, formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils";
import type { Payment, PaymentStatus } from "@/types/booking.types";

const PAYMENT_STATUS_STYLES: Record<PaymentStatus, { label: string; className: string }> = {
  pending: { label: "Pending", className: "text-amber-700 bg-amber-50 ring-amber-600/20" },
  processing: { label: "Processing", className: "text-sky-700 bg-sky-50 ring-sky-600/20" },
  completed: { label: "Paid", className: "text-emerald-700 bg-emerald-50 ring-emerald-600/20" },
  failed: { label: "Failed", className: "text-rose-700 bg-rose-50 ring-rose-600/20" },
  refunded: { label: "Refunded", className: "text-slate-700 bg-slate-100 ring-slate-600/20" },
};

const PAYMENT_METHOD_LABELS: Record<Payment["payment_method"], string> = {
  internet_banking: "Internet banking",
  bank_transfer: "Bank transfer",
  vnpay: "VNPay",
  momo: "MoMo",
};

interface BookingPaymentCardProps {
  /** Undefined whenever the booking has no payments row — a legitimate
   * state, not an error (FR-212 / R5). */
  payment: Payment | undefined;
}

/**
 * FR-212: the linked payment, strictly read-only. Nothing in F004 writes the
 * payments table — refunds are out of scope — so this card has no actions at
 * all, by design.
 */
export function BookingPaymentCard({ payment }: BookingPaymentCardProps) {
  if (!payment) {
    return (
      <section className="rounded-lg border p-5">
        <h2 className="text-sm font-semibold mb-1">Payment</h2>
        <p className="text-sm text-muted-foreground">No payment recorded for this booking.</p>
      </section>
    );
  }

  const status = PAYMENT_STATUS_STYLES[payment.status];

  return (
    <section className="rounded-lg border p-5">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold">Payment</h2>
        <span className={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-semibold ring-1 ring-inset", status.className)}>
          {status.label}
        </span>
      </div>
      <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
        <Field label="Amount" value={formatCurrencyVND(payment.amount)} />
        <Field label="Method" value={PAYMENT_METHOD_LABELS[payment.payment_method] ?? payment.payment_method} />
        <Field label="Bank" value={payment.bank_name ?? "—"} />
        <Field label="Reference" value={payment.transaction_ref ?? "—"} />
        <Field label="Paid at" value={payment.paid_at ? formatDate(payment.paid_at) : "—"} />
      </dl>
    </section>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 break-words">{value}</dd>
    </div>
  );
}
