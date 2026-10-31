import type { HealthView } from "../types/health.ts";

export type HealthStatusProps =
  | { state: "loading" }
  | { state: "unreachable" }
  | { state: "loaded"; health: HealthView };

// Shows whether the backend and its stores are reachable.
export function HealthStatus(props: HealthStatusProps) {
  return (
    <section aria-label="Backend health">
      <h2>Backend health</h2>
      {props.state === "loading" && <p>Checking…</p>}
      {props.state === "unreachable" && <p>The backend cannot be reached.</p>}
      {props.state === "loaded" && (
        <>
          <p>{props.health.healthy ? "Healthy" : "Unhealthy"}</p>
          <dl>
            {props.health.stores.map((store) => (
              <div key={store.name}>
                <dt>{store.name}</dt>
                <dd>{store.reachable ? "Reachable" : "Unreachable"}</dd>
              </div>
            ))}
          </dl>
        </>
      )}
    </section>
  );
}
