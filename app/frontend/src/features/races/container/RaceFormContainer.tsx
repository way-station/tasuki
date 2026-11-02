import { useNavigate, useParams } from "@tanstack/react-router";
import { useState } from "react";
import { useDraft } from "../../../shared/draft/useDraft.ts";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useMe } from "../../auth/hooks/useMe.ts";
import { raceFormErrors } from "../hooks/raceFormInvalid.ts";
import { useCreateRace } from "../hooks/useCreateRace.ts";
import { useRaceFormSource } from "../hooks/useRaceFormSource.ts";
import { useUpdateRace } from "../hooks/useUpdateRace.ts";
import type { RaceFormErrors } from "../types/race.ts";
import { RaceFormView } from "../ui/RaceFormView.tsx";
import { RaceMessage } from "../ui/RaceMessage.tsx";

// R3 (docs/design/screens.md), at /races/new to register and at
// /races/$raceId/edit to edit. Every edit is kept as a draft until it is saved
// (ADR-0013).
export function RaceFormContainer() {
  // Registering a new race when the URL has no race.
  const { raceId } = useParams({ strict: false });
  const me = useMe();
  const { source, error } = useRaceFormSource(raceId);
  const draft = useDraft(
    me.data?.userId,
    raceId === undefined ? "race-new" : `race-${raceId}`,
    source,
  );
  const createRace = useCreateRace();
  const updateRace = useUpdateRace(raceId ?? "");
  const { notify } = useNotifications();
  const navigate = useNavigate();
  const [errors, setErrors] = useState<RaceFormErrors>({});
  const [conflict, setConflict] = useState(false);

  if (error !== null) {
    if (errorCode(error) === "not_found") {
      return <RaceMessage title="Race not found" />;
    }
    return <RaceMessage title="Something went wrong" message="Reload the page." />;
  }
  if (draft.value === undefined) {
    return <RaceMessage title="Loading…" />;
  }
  const form = draft.value;

  const save = async () => {
    setErrors({});
    try {
      const saved =
        raceId === undefined
          ? await createRace.mutateAsync(form)
          : await updateRace.mutateAsync({ form, version: draft.version ?? 0 });
      draft.clear();
      notify("success", raceId === undefined ? "Race registered." : "Race saved.");
      void navigate({ to: "/races/$raceId", params: { raceId: saved.id } });
    } catch (caught) {
      const fieldErrors = raceFormErrors(caught);
      if (fieldErrors !== undefined) {
        setErrors(fieldErrors);
        return;
      }
      switch (errorCode(caught)) {
        case "version_conflict":
          // The draft stays; the user chooses to load the latest (ADR-0013).
          setConflict(true);
          notify("error", "Someone else saved this race first.");
          return;
        case "validation_failed":
          notify("error", "The race could not be saved. Check the inputs.");
          return;
        default:
          notify("error", "The race could not be saved. Try again.");
      }
    }
  };

  const loadLatest = () => {
    draft.clear();
    setErrors({});
    setConflict(false);
  };

  return (
    <RaceFormView
      title={raceId === undefined ? "Register race" : "Edit race"}
      form={form}
      errors={errors}
      saving={createRace.isPending || updateRace.isPending}
      conflict={conflict}
      onChange={draft.edit}
      onSave={() => void save()}
      onLoadLatest={loadLatest}
    />
  );
}
