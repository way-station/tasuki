# ADR-0018: Hosting — one container on Cloud Run in Singapore

- Status: Accepted
- Type: technical
- Date: 2026-10-24
- Issue: #178

## Context

Production needs a place for the backend and the frontend. The databases are already chosen: Neon for PostgreSQL and MongoDB Atlas (`app/backend/TECH.md`). Related: #24.

The forces:

- **The frontend and the API must be on the same site**, because the session is an HttpOnly cookie with `SameSite=Lax` and requests that change data are checked against the frontend's `Origin` (ADR-0014).
- **Neon has no region in Japan.** Its nearest region is Singapore (`aws-ap-southeast-1`), and a project's region cannot be changed later. MongoDB Atlas's free cluster appears to be available in Singapore (AWS `ap-southeast-1`).
- **One screen makes several database queries.** Each query pays the network round trip between the backend and the database, while the user pays the round trip to the backend once per request.
- **The team is one relay team using tasuki occasionally** (ADR-0005), so the app is idle most of the time. Costs should stay within free tiers.
- The backend already builds a small container (`infra/docker/backend/Dockerfile`, distroless). The backend runs a retry worker for the outbox (ADR-0016).

## Decision

**The backend and the built frontend run as one container on Google Cloud Run, in Singapore (`asia-southeast1`), next to Neon and MongoDB Atlas in Singapore.**

**One container, one origin**

- The Go backend serves the built frontend files and the API from the same origin: the API under `/api`, and everything else returns the frontend (with `index.html` for client-side routes).
- The production image builds the frontend in a Node stage and copies the result into the backend image. There is no separate frontend container.
- Until a custom domain is chosen, the Cloud Run URL (`*.run.app`) is used. It already serves the frontend and the API from one origin, so the cookie rules of ADR-0014 hold. The LINE Login callback is registered for that URL.

**Region**

| Part | Where |
| --- | --- |
| Backend and frontend | Cloud Run, `asia-southeast1` (Singapore) |
| PostgreSQL | Neon, `aws-ap-southeast-1` (Singapore) |
| MongoDB | MongoDB Atlas, AWS `ap-southeast-1` (Singapore); free cluster availability is confirmed when the cluster is created |

**Scaling and cost**

- Request-based billing, with **minimum instances 0** (the service stops when idle).
- **Maximum instances are limited** so that maximum instances × `MaxConns` stays below Neon's `max_connections` (ADR-0015).
- The goal is to stay within the free tiers of Cloud Run, Neon, and MongoDB Atlas.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Frontend on a CDN (Cloudflare Pages, Vercel) with `/api` forwarded to the backend | Two places to deploy and configure. If the forwarding is set up wrong, the cookie and the `Origin` check stop working |
| Cloud Run in Tokyo (`asia-northeast1`) | Closer to the users, but every database query would cross between Tokyo and Singapore, and one screen makes several |
| Fly.io in Singapore (always on) | Costs every month. The backend would not stop, but Neon's free plan still suspends an idle database, so the first request after a pause would still be slow |
| Render's free tier | Stops after 15 minutes idle and takes about a minute to wake up; members opening a link from LINE would wait too long |
| Railway | A monthly fee, for no benefit tasuki needs over Cloud Run |
| Move everything to Tokyo with a PostgreSQL provider that has a Tokyo region | Reverses the choice of Neon |

## Consequences

- One deployment ships the backend and the frontend together; a frontend-only change still deploys the container.
- The same-site requirement of ADR-0014 is met by construction.
- Users in Japan reach Singapore once per request; queries between the backend and the databases stay in the same region.
- After an idle period, the first request is slower: Cloud Run starts an instance, and Neon resumes its compute. Later requests are fast.
- The outbox retry worker (ADR-0016) runs only while an instance is running. Remaining steps normally finish in the request; leftovers are retried when the next request starts an instance.
- Singapore is a Tier 2 Cloud Run region, with a smaller free allowance than Tier 1 regions; usage is checked against it after launch.
- The production image now needs both `app/frontend` and `app/backend`, so its build context changes. #27 (a separate frontend image) is no longer needed and is closed.
- Secrets (database connection strings, the LINE channel secret) and applying migrations are set up in #28; deployment on merge in #29. Logs go to Cloud Logging as JSON (`log/slog`), where outbox rows that keep failing can be noticed.
- In local development nothing changes: the frontend's dev server forwards `/api` to the backend (ADR-0014).

---

<details>
<summary>日本語</summary>

# ADR-0018: ホスティング — シンガポールの Cloud Run で 1 つのコンテナ

- Status: Accepted
- Type: technical
- Date: 2026-10-24
- Issue: #178

## Context（背景）

本番で backend と frontend を動かす場所が要る。DB はすでに決まっている：PostgreSQL は Neon、MongoDB は MongoDB Atlas（`app/backend/TECH.md`）。関連：#24。

考慮したこと：

- **frontend と API は同じサイトにある必要がある。** セッションは `SameSite=Lax` の HttpOnly Cookie で、データを変えるリクエストは frontend の `Origin` と照らし合わせるため（ADR-0014）
- **Neon には日本のリージョンがない。** 一番近いのはシンガポール（`aws-ap-southeast-1`）で、プロジェクトのリージョンは後から変えられない。MongoDB Atlas の無料のクラスタは、シンガポール（AWS `ap-southeast-1`）で使えるとみられる
- **1 つの画面で DB に何回も問い合わせる。** 問い合わせのたびに backend と DB の間の往復がかかる。一方、ユーザーから backend への往復はリクエストごとに 1 回
- **使うのは 1 つのリレーチームで、使う頻度は高くない**（ADR-0005）。アプリはほとんどの時間、使われていない。費用は無料枠に収めたい
- backend はすでに小さなコンテナを作れる（`infra/docker/backend/Dockerfile`、distroless）。backend は outbox の再試行の仕組みを動かす（ADR-0016）

## Decision（決定）

**backend とビルドした frontend を 1 つのコンテナにして、Google Cloud Run のシンガポール（`asia-southeast1`）で動かす。Neon と MongoDB Atlas もシンガポールに置く。**

**1 つのコンテナ、1 つのオリジン**

- Go の backend が、ビルドした frontend のファイルと API を同じオリジンから出す：API は `/api` の下、それ以外は frontend を返す（クライアント側のルートには `index.html` を返す）
- 本番のイメージは、Node のステージで frontend をビルドし、その結果を backend のイメージにコピーする。frontend だけのコンテナは作らない
- 独自ドメインを決めるまでは、Cloud Run の URL（`*.run.app`）を使う。この URL でも frontend と API は 1 つのオリジンから出るので、ADR-0014 の Cookie のルールは成り立つ。LINE ログインのコールバックはこの URL で登録する

**リージョン**

| もの | 場所 |
| --- | --- |
| backend と frontend | Cloud Run、`asia-southeast1`（シンガポール） |
| PostgreSQL | Neon、`aws-ap-southeast-1`（シンガポール） |
| MongoDB | MongoDB Atlas、AWS `ap-southeast-1`（シンガポール）。無料のクラスタが使えるかは、作るときに確かめる |

**台数と費用**

- リクエスト単位の課金で、**最小インスタンス数は 0**（使われていないときは止まる）
- **最大インスタンス数を制限する。** 最大インスタンス数 × `MaxConns` が Neon の `max_connections` を下回るようにする（ADR-0015）
- Cloud Run、Neon、MongoDB Atlas の無料枠に収めることを目標にする

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| frontend を CDN（Cloudflare Pages、Vercel）に置き、`/api` を backend に転送する | デプロイと設定の場所が 2 つになる。転送の設定を間違えると、Cookie と `Origin` の確認が効かなくなる |
| 東京の Cloud Run（`asia-northeast1`） | ユーザーには近いが、DB への問い合わせのたびに東京とシンガポールの間を往復する。1 つの画面で何回も問い合わせる |
| シンガポールの Fly.io（常に動かす） | 毎月費用がかかる。backend は止まらないが、Neon の無料プランは使われていない DB を止めるので、しばらく使わなかった後の最初のリクエストはやはり遅い |
| Render の無料枠 | 15 分使わないと止まり、起きるのに約 1 分かかる。LINE からリンクを開いたメンバーが待ちきれない |
| Railway | 月額がかかり、Cloud Run と比べて tasuki に必要な利点がない |
| 東京リージョンのある PostgreSQL のサービスに変え、すべてを東京に置く | Neon という選択を覆す |

## Consequences（影響）

- 1 回のデプロイで backend と frontend を一緒に出す。frontend だけを変えても、コンテナごとデプロイする
- ADR-0014 の「同じサイト」という条件は、構成そのもので満たされる
- 日本のユーザーからシンガポールへの往復はリクエストごとに 1 回。backend と DB の間の問い合わせは同じリージョンの中で済む
- しばらく使われなかった後の最初のリクエストは遅い：Cloud Run がインスタンスを起動し、Neon が計算資源を再開するため。その後のリクエストは速い
- outbox の再試行の仕組み（ADR-0016）は、インスタンスが動いている間だけ働く。残りの手順はふつうリクエストの中で終わり、残ったものは次のリクエストがインスタンスを起動したときに再試行される
- シンガポールは Cloud Run の Tier 2 のリージョンで、Tier 1 より無料枠が小さい。公開した後に、使った量を無料枠と照らし合わせる
- 本番のイメージは `app/frontend` と `app/backend` の両方を必要とするので、ビルドのコンテキストが変わる。#27（frontend だけのイメージ）は不要になるので閉じる
- シークレット（DB の接続文字列、LINE のチャネルシークレット）とマイグレーションの適用は #28 で、マージしたときのデプロイは #29 で用意する。ログは JSON（`log/slog`）で Cloud Logging に送られ、失敗し続ける outbox の行はそこで気づける
- ローカル開発は変わらない：frontend の開発サーバーが `/api` を backend に転送する（ADR-0014）

</details>
