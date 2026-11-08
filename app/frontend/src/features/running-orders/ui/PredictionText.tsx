import type { PredictionView } from "../types/runningOrder.ts";

type PredictionTextProps = {
  prediction: PredictionView | null;
};

// The prediction, which leaves empty legs out (docs/design/screens.md).
export function PredictionText({ prediction }: PredictionTextProps) {
  if (prediction === null) {
    return <p>No prediction yet: no leg has a runner.</p>;
  }
  return (
    <p>
      Prediction:{" "}
      {prediction.type === "distance" ? prediction.totalTime : `${prediction.laps} laps`}
      {prediction.hasEmptyLegs && " (some legs are empty and left out)"}
    </p>
  );
}
