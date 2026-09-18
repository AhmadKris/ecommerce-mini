/** Pagination metadata every list endpoint responds with. */
export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ApiSuccessEnvelope<T> {
  success: true;
  data: T;
  message?: string;
  meta?: PaginationMeta;
}

/** Shape of a list endpoint's unwrapped data: `{success, data: {items, meta}}`. */
export interface PaginatedResponse<T> {
  items: T[];
  meta: PaginationMeta;
}

export interface ApiErrorEnvelope {
  success: false;
  error: {
    code: string;
    message: string;
    details: string[] | null;
  };
}

/**
 * Normalized shape every failed apiClient call rejects with — components
 * and React Query's onError never need to reach into AxiosError internals.
 */
export class ApiError extends Error {
  code: string;
  details: string[];
  status?: number;

  constructor(message: string, code: string, details: string[] = [], status?: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.details = details;
    this.status = status;
  }
}
