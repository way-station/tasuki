import { useHealth } from "../hooks/useHealth.ts";
import { HealthStatus } from "../ui/HealthStatus.tsx";

export function HealthContainer() {
  const health = useHealth();

  if (health.isPending) {
    return <HealthStatus state="loading" />;
  }
  if (health.isError) {
    return <HealthStatus state="unreachable" />;
  }
  return <HealthStatus state="loaded" health={health.data} />;
}
