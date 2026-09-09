import { ApiError, type ApiErrorCode } from "./types";

type UnauthorizedHandler = () => void;

let onUnauthorized: UnauthorizedHandler | null = null;

/**
 * Registers the single callback fired whenever apiFetch sees a 401. Phase 2
 * wires this to redirect to /admin/login — apiFetch itself never touches
 * window.location so this module stays framework/navigation agnostic.
 */
export function setOnUnauthorized(handler: UnauthorizedHandler | null): void {
  onUnauthorized = handler;
}

interface ErrorResponseBody {
  error: {
    code: ApiErrorCode;
    message: string;
    fields?: Record<string, string>;
  };
}

function isErrorResponseBody(value: unknown): value is ErrorResponseBody {
  if (typeof value !== "object" || value === null) return false;
  const err = (value as { error?: unknown }).error;
  return typeof err === "object" && err !== null && "code" in err && "message" in err;
}

export interface ApiFetchInit extends Omit<RequestInit, "body"> {
  /** Encoded to JSON automatically; omit for GET/DELETE with no body. */
  body?: unknown;
}

/**
 * Fetch wrapper for the /api/v1/admin/* endpoints proxied through Next's
 * rewrites(). Always sends credentials (the HttpOnly sun_admin_token
 * cookie), JSON-encodes a provided body, and JSON-decodes the response.
 * Throws ApiError on any non-2xx response.
 */
export async function apiFetch<T>(path: string, init: ApiFetchInit = {}): Promise<T> {
  const { body, headers: initHeaders, ...rest } = init;
  const headers = new Headers(initHeaders);
  headers.set("Accept", "application/json");

  let encodedBody: BodyInit | undefined;
  if (body !== undefined) {
    headers.set("Content-Type", "application/json");
    encodedBody = JSON.stringify(body);
  }

  const response = await fetch(path, {
    ...rest,
    headers,
    body: encodedBody,
    credentials: "include",
  });

  if (response.status === 401) {
    onUnauthorized?.();
  }

  if (!response.ok) {
    throw await toApiError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

async function toApiError(response: Response): Promise<ApiError> {
  let body: unknown = null;
  try {
    body = await response.json();
  } catch {
    // Non-JSON error body (e.g. a proxy/network failure page) — fall
    // through to the generic ApiError below.
  }

  if (isErrorResponseBody(body)) {
    return new ApiError(response.status, body.error.code, body.error.message, body.error.fields);
  }

  return new ApiError(response.status, "internal", response.statusText || "Request failed");
}
