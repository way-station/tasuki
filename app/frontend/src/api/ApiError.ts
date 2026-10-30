import type { components } from "./schema.gen.ts";

export type Problem = components["schemas"]["Problem"];
export type ProblemCode = Problem["code"];

// An error response from the backend. The backend answers errors with RFC 9457
// problem+json and a `code` (docs/design/api.md); `code` is undefined only when
// something else answered, such as a proxy returning 502.
export class ApiError extends Error {
  readonly status: number;
  readonly code: ProblemCode | undefined;
  readonly problem: Problem | undefined;

  constructor(status: number, problem?: Problem) {
    super(problem ? `${status} ${problem.code}: ${problem.title}` : `${status}`);
    this.name = "ApiError";
    this.status = status;
    this.code = problem?.code;
    this.problem = problem;
  }
}

export function toApiError(response: Response, body: unknown): ApiError {
  const isProblem =
    response.headers.get("Content-Type")?.startsWith("application/problem+json") &&
    typeof body === "object" &&
    body !== null &&
    "code" in body;
  return new ApiError(response.status, isProblem ? (body as Problem) : undefined);
}
