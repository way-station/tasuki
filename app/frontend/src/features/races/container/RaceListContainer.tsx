import { useNavigate } from "@tanstack/react-router";
import { useMe } from "../../auth/hooks/useMe.ts";
import { useRaces } from "../hooks/useRaces.ts";
import { RaceList } from "../ui/RaceList.tsx";
import { RaceMessage } from "../ui/RaceMessage.tsx";

// R1 (docs/design/screens.md). The route guard has already loaded the user.
export function RaceListContainer() {
  const me = useMe();
  const races = useRaces(me.data?.isOrganizer ?? false);
  const navigate = useNavigate();

  if (races.isPending) {
    return <RaceMessage title="Loading…" />;
  }
  if (races.isError) {
    return <RaceMessage title="Something went wrong" message="Reload the page." />;
  }
  return (
    <RaceList
      view={races.data}
      onOpen={(raceId) => void navigate({ to: "/races/$raceId", params: { raceId } })}
      onRegister={() => void navigate({ to: "/races/new" })}
    />
  );
}
