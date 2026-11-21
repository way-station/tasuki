import {
  card,
  cards,
  cardStack,
  content,
  dangerButton,
  group,
  groupHeading,
  primaryButton,
  regularButton,
} from "../../../shared/ui/classes.ts";
import type { PendingLinkView } from "../types/roster.ts";
import { formatDateTime } from "./labels.ts";

type PendingLinkListProps = {
  links: PendingLinkView[];
  busy: boolean;
  onApprove: (link: PendingLinkView) => void;
  onReject: (link: PendingLinkView) => void;
};

// M3 (docs/design/screens.md, wireframes/m3.html). The roster name and the LINE
// name side by side, so the organizer can tell it is the right person.
// Rejections are told outside the app.
export function PendingLinkList({ links, busy, onApprove, onReject }: PendingLinkListProps) {
  return (
    <section>
      <header className="border-b border-slate-200 bg-white p-4">
        <h1 className="text-large-title font-bold">承認</h1>
      </header>
      <div className={content}>
        {links.length === 0 ? (
          <p className="text-subheadline text-slate-600">承認待ちはありません</p>
        ) : (
          <section aria-label="承認待ち" className={group}>
            <h2 className={groupHeading}>承認待ち（{links.length}）</h2>
            <ul className={cards}>
              {links.map((link) => {
                const who = `${link.rosterMember.name}（${link.lineDisplayName}）`;
                return (
                  <li key={link.id} className={`${card} ${cardStack}`}>
                    <div className="flex flex-col gap-0.5">
                      <span className="text-headline font-semibold">{link.rosterMember.name}</span>
                      <span className="text-subheadline text-slate-600">
                        LINE の名前：{link.lineDisplayName}
                      </span>
                      <span className="text-subheadline text-slate-600">
                        <time dateTime={link.createdAt}>{formatDateTime(link.createdAt)}</time>{" "}
                        にログイン
                      </span>
                    </div>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        disabled={busy}
                        aria-label={`${who}を承認`}
                        onClick={() => onApprove(link)}
                        className={`${primaryButton} ${regularButton} flex-1`}
                      >
                        承認
                      </button>
                      <button
                        type="button"
                        disabled={busy}
                        aria-label={`${who}を却下`}
                        onClick={() => onReject(link)}
                        className={`${dangerButton} ${regularButton} flex-1`}
                      >
                        却下
                      </button>
                    </div>
                  </li>
                );
              })}
            </ul>
          </section>
        )}
      </div>
    </section>
  );
}
