import { ApiError, type ProblemCode } from "../../../api/ApiError.ts";

// Lets containers tell errors apart without importing the API client (ADR-0012).
export function errorCode(error: unknown): ProblemCode | undefined {
  return error instanceof ApiError ? error.code : undefined;
}

export function isUnauthenticated(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}
