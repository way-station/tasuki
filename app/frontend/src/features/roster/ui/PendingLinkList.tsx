import type { PendingLinkView } from "../types/roster.ts";
import { formatDateTime } from "./labels.ts";

type PendingLinkListProps = {
  links: PendingLinkView[];
  busy: boolean;
  onApprove: (link: PendingLinkView) => void;
  onReject: (link: PendingLinkView) => void;
};

// M3 (docs/design/screens.md). Rejections are told outside the app.
export function PendingLinkList({ links, busy, onApprove, onReject }: PendingLinkListProps) {
  return (
    <section>
      <h1>Approvals</h1>
      {links.length === 0 ? (
        <p>Nobody is waiting for approval.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th scope="col">LINE name</th>
              <th scope="col">Roster member</th>
              <th scope="col">Logged in</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {links.map((link) => (
              <tr key={link.id}>
                <td>{link.lineDisplayName}</td>
                <td>{link.rosterMember.name}</td>
                <td>
                  <time dateTime={link.createdAt}>{formatDateTime(link.createdAt)}</time>
                </td>
                <td>
                  <button type="button" disabled={busy} onClick={() => onApprove(link)}>
                    Approve {link.lineDisplayName}
                  </button>
                  <button type="button" disabled={busy} onClick={() => onReject(link)}>
                    Reject {link.lineDisplayName}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
