// Every query key, in one place, so a mutation can invalidate every query it
// affects (ADR-0013).
export const queryKeys = {
  health: ["health"],
} as const;
