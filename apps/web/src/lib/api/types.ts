// Shared shapes for every admin API call. These mirror the Go API's
// pagination envelope and error contract (apps/api/internal/apperror,
// apps/api/internal/repository/params.go) — keep them in sync with those.

/** Envelope every list endpoint returns. */
export interface Paginated<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

/** Error codes the API ever emits (apps/api/internal/apperror/apperror.go). */
export type ApiErrorCode =
  | "bad_request"
  | "unauthorized"
  | "forbidden"
  | "not_found"
  | "conflict"
  | "unprocessable"
  | "rate_limited"
  | "internal";

/**
 * Thrown by apiFetch for any non-2xx response. fields is only populated for
 * 409/422 responses (field-scoped validation/conflict messages).
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: ApiErrorCode;
  readonly fields?: Record<string, string>;

  constructor(status: number, code: ApiErrorCode, message: string, fields?: Record<string, string>) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

/** Query params every admin list endpoint accepts. */
export interface ListQuery {
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_dir?: "asc" | "desc";
  search?: string;
}
