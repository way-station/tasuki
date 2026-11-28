# ADR-0021: Production configuration and deployment — Secret Manager, migrations from CI, open database networks, keyless CI, capped costs

- Status: Accepted
- Type: technical
- Date: 2026-11-25
- Issue: #28

## Context

Production runs as one container on Cloud Run in Singapore, next to Neon and MongoDB Atlas (ADR-0018). Before the first deployment, five questions are open. They are asked in #28, #198, and #200, and #29 depends on all of them:

1. Where the production secrets live.
2. How migrations reach the production PostgreSQL. ADR-0015 says the deployment applies them over Neon's direct connection, but not where they run.
3. How Cloud Run reaches MongoDB Atlas and Neon over the network.
4. How GitHub Actions signs in to Google Cloud to deploy.
5. How the monthly cost is kept small and stopped from growing.

The forces:

- **Costs should stay as close to zero as possible.** The goal of ADR-0018 is to stay within free tiers.
- **The repository is public.** Anything committed, including workflow files, can be read by anyone.
- **The backend reads only environment variables** (`app/backend/TECH.md`). Three of them are secrets: `DATABASE_URL`, `MONGODB_URI`, and `LINE_CHANNEL_SECRET`. The rest are not secret but some are private, such as `ORGANIZER_LINE_USER_IDS`.
- **Cloud Run has no fixed outbound IP address by default.** A fixed one needs a VPC, a Cloud Router, a reserved IP address, and a Cloud NAT gateway ([Static outbound IP](https://docs.cloud.google.com/run/docs/configuring/static-outbound-ip)).
- **Neither free database can restrict access by IP address in a cheap way.**
  - Atlas Free clusters support neither network peering nor private endpoints ([Free cluster limits](https://www.mongodb.com/docs/atlas/reference/free-shared-limitations/)). Only the IP access list is left.
  - Neon's IP Allow is available only on the Scale plan ([IP Allow](https://neon.com/docs/introduction/ip-allow)).
- **Neon Free and Atlas Free cannot charge.** Neither needs a card. When a limit is reached, Neon suspends the compute and Atlas slows down ([Neon pricing](https://neon.com/pricing), [Free cluster limits](https://www.mongodb.com/docs/atlas/reference/free-shared-limitations/)).
- **Google Cloud can charge.** Cloud Run needs a billing account.
  - Its free tier covers CPU, memory, and requests, but internet egress is free only within North America. Responses from Singapore to users in Japan are charged from the first byte ([Cloud Run pricing](https://cloud.google.com/run/pricing)).
  - Free tiers are counted per billing account, not per project.
  - A budget alert only sends a notice. It does not stop spending ([Budgets](https://docs.cloud.google.com/billing/docs/how-to/budgets)).
  - A spend cap budget pauses Cloud Run when the cap is reached, until the next period or until it is lifted by hand. It does not stop charges for stored resources ([Spend cap budgets](https://docs.cloud.google.com/billing/docs/how-to/budgets-spend-caps)).
  - Cost data reaches the budget with a delay, so spending can pass the amount before the budget reacts.

## Decision

**Secrets in Secret Manager**

- `DATABASE_URL`, `MONGODB_URI`, and `LINE_CHANNEL_SECRET` are stored in Secret Manager. Cloud Run reads them as environment variables, which are resolved when an instance starts ([Configure secrets](https://docs.cloud.google.com/run/docs/configuring/services/secrets)).
- The service's identity gets only `roles/secretmanager.secretAccessor`.
- Old secret versions are disabled or destroyed after a change. This keeps the count within the free tier of 6 active versions and 10,000 accesses a month ([Secret Manager pricing](https://cloud.google.com/secret-manager/pricing)).
- The other variables (`TEAM_ID`, `FRONTEND_ORIGIN`, `FRONTEND_DIR`, `ORGANIZER_LINE_USER_IDS`, `LINE_CHANNEL_ID`, `LINE_REDIRECT_URI`) are set once on the Cloud Run service by hand.
- The deployment changes only the image, so these values never appear in the public repository.

**Migrations from GitHub Actions, before the deployment**

- The deploy workflow (#29) runs golang-migrate against Neon's direct connection, then deploys the new image. If the migration fails, nothing is deployed.
- The direct connection string is stored as a GitHub Actions secret. It never appears in logs.
- Each migration must work with the image that is still running until the deployment finishes. A change that removes or renames something is split into two deployments: add first, remove later.

**Database networks open, protected by credentials and TLS**

- Atlas allows `0.0.0.0/0` in its IP access list. Neon stays at its default, which accepts connections from anywhere.
- Both are protected by the rest:
  - TLS on every connection.
  - Long random passwords that live only in Secret Manager and GitHub Actions secrets.
  - Database users that have only the access the backend needs. In Atlas, that is `readWrite` on the tasuki database only.

**GitHub Actions signs in with Workload Identity Federation**

- The deploy workflow uses `google-github-actions/auth` with direct Workload Identity Federation, the method it recommends ([google-github-actions/auth](https://github.com/google-github-actions/auth)).
- No service account key exists.
- The provider accepts only tokens from this repository's `main` branch.
- IAM, which includes it, has no charge ([IAM billing questions](https://docs.cloud.google.com/iam/docs/billing-questions)).

**A small, capped monthly cost**

- A small monthly cost for Cloud Run's internet egress is accepted. Everything else stays within free tiers.
- The tasuki project gets a spend cap budget on Cloud Run of ¥100 a month. Budget alerts are sent by email at 50%, 90%, and 100% of the same amount.
- The billing account is used only for tasuki, so no other project uses up its free tier.
- Cloud Run keeps the settings of ADR-0018: request-based billing, minimum instances 0, and a limited maximum.
- These are not turned on, because each costs money: Cloud NAT, a VPC connector, a load balancer, Cloud CDN, a reserved IP address, and Artifact Registry vulnerability scanning.
- The image is built in GitHub Actions, which is free for public repositories, not in Cloud Build.
- Artifact Registry has a cleanup policy that keeps only the latest 10 images. This keeps storage within the free 0.5 GiB ([Artifact Registry pricing](https://cloud.google.com/artifact-registry/pricing)).

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Secrets as plain Cloud Run environment variables | Anyone who can view the service in the console sees them, and they appear in the service's revision history. Secret Manager costs nothing at this size |
| Secrets in GitHub Actions only, passed to `gcloud run deploy` on each deployment | Every deployment would carry the secrets through the workflow. A mistake in a public workflow file could print them |
| Migrations at backend startup | Several instances can start at once and race. Every cold start would also pay for the check. Cloud Run often starts instances, because minimum instances are 0 (ADR-0018) |
| Migrations in a Cloud Run job | Needs a second image or entry point, and the direct connection string in Secret Manager as well. It adds setup but brings no benefit over a step in the workflow, which also costs nothing |
| Migrations by hand | Easy to forget. The image and the schema could drift apart |
| Fixed outbound IP (VPC, Cloud NAT, reserved IP) with Atlas allowing only that IP | Costs every hour. Cloud NAT charges per gateway hour and per GiB processed, and the reserved IP costs per hour as well ([Cloud NAT pricing](https://cloud.google.com/nat/pricing)). It also protects only Atlas, because Neon could not use the IP without the Scale plan |
| Paid Atlas cluster with private endpoints, or Neon Scale with IP Allow | A monthly fee for the MVP |
| Service account key (JSON) in GitHub Actions secrets | A long-lived credential that must be treated like a password and rotated. Google recommends avoiding user-managed keys ([Best practices for service account keys](https://docs.cloud.google.com/iam/docs/best-practices-for-managing-service-account-keys)) |
| Budget alerts only, without a spend cap | An alert does not stop spending. A sudden rise in traffic would keep costing until someone acts |
| Disable billing automatically when the budget is reached (Pub/Sub and a function) | Google warns that resources might be irretrievably deleted ([Disable billing with notifications](https://docs.cloud.google.com/billing/docs/how-to/disable-billing-with-notifications)). It also reacts late, because cost data is delayed |
| Serve the frontend from a free static host to avoid egress | Breaks the one origin of ADR-0018, which the session cookie and the `Origin` check rely on (ADR-0014). The API responses would still be charged |

## Consequences

- Production costs nothing extra for secrets, migrations, network access, or CI sign-in while usage stays within the free tiers.
- **Production is not strictly free.** Cloud Run's egress to Japan costs a small amount each month. The real amount is checked in the billing report after launch.
- **When the spend cap is reached, tasuki stops for everyone** until the next month or until the cap is lifted by hand. Stored images in Artifact Registry keep being charged if they pass the free tier.
- Because cost data is delayed, the month's cost can pass ¥100 a little before Cloud Run is paused.
- **The databases can be reached from any IP address.** A leaked connection string is enough to get in, so it must never be logged, committed, or pasted. If one leaks, change the password at once. This choice should be revisited when tasuki pays for a plan that can restrict networks.
- Changing a secret means adding a new version and deploying a new revision, because environment-variable secrets are read at startup. The old version is then disabled.
- The service's non-secret configuration lives only on the Cloud Run service until it is written as code. Infrastructure as code comes after the manual setup.
- Migrations must stay compatible with the image that is still running. This rule belongs in the migration guide written in #28.
- The order of the setup issues becomes: #200 (Google Cloud and Workload Identity Federation), #25 and #198 (the databases), #199 (LINE), #28 (the secrets and the written procedure), then #29 (the deploy workflow).

---

<details>
<summary>日本語</summary>

# ADR-0021: 本番の設定とデプロイ — Secret Manager、CI からのマイグレーション、開いた DB のネットワーク、鍵のない CI、上限のある費用

- Status: Accepted
- Type: technical
- Date: 2026-11-25
- Issue: #28

## Context（背景）

本番は、シンガポールの Cloud Run で 1 つのコンテナとして動き、Neon と MongoDB Atlas もシンガポールに置く（ADR-0018）。最初のデプロイの前に、5 つのことが決まっていない。#28・#198・#200 で問われていて、#29 はそのすべてに依存する。

1. 本番のシークレットをどこに置くか。
2. マイグレーションをどうやって本番の PostgreSQL に流すか。ADR-0015 は「デプロイで Neon の直接接続を使って流す」とだけ決めていて、どこで流すかは決めていない。
3. Cloud Run から MongoDB Atlas と Neon へ、ネットワークでどう届くか。
4. GitHub Actions が、デプロイのためにどう Google Cloud にサインインするか。
5. 毎月の費用をどう小さく保ち、膨らまないようにするか。

考慮したこと：

- **費用はできるだけ 0 に近づけたい。** ADR-0018 の目標は無料枠に収めること。
- **リポジトリは公開されている。** コミットしたものは、ワークフローのファイルも含めて誰でも読める。
- **backend は環境変数だけを読む**（`app/backend/TECH.md`）。そのうち 3 つがシークレット：`DATABASE_URL`、`MONGODB_URI`、`LINE_CHANNEL_SECRET`。残りはシークレットではないが、`ORGANIZER_LINE_USER_IDS` のように公開したくないものもある。
- **Cloud Run には、既定では決まった送信元 IP アドレスがない。** 決まったアドレスにするには、VPC、Cloud Router、予約した IP アドレス、Cloud NAT ゲートウェイが要る（[Static outbound IP](https://docs.cloud.google.com/run/docs/configuring/static-outbound-ip)）。
- **どちらの無料の DB も、安い方法では IP アドレスで接続を絞れない。**
  - Atlas の Free クラスタは、ネットワークピアリングにもプライベートエンドポイントにも対応していない（[Free cluster limits](https://www.mongodb.com/docs/atlas/reference/free-shared-limitations/)）。残るのは IP アクセスリストだけ。
  - Neon の IP Allow は Scale プランだけで使える（[IP Allow](https://neon.com/docs/introduction/ip-allow)）。
- **Neon Free と Atlas Free は請求できない。** どちらもカードが要らない。上限に達すると、Neon はコンピュートを止め、Atlas は速度を絞る（[Neon pricing](https://neon.com/pricing)、[Free cluster limits](https://www.mongodb.com/docs/atlas/reference/free-shared-limitations/)）。
- **Google Cloud は請求できる。** Cloud Run には請求先アカウントが要る。
  - 無料枠は CPU・メモリ・リクエストにあるが、インターネットへの送信が無料なのは北米の中だけである。シンガポールから日本の利用者への応答は、最初の 1 バイトから料金がかかる（[Cloud Run pricing](https://cloud.google.com/run/pricing)）。
  - 無料枠は、プロジェクトごとではなく、請求先アカウントごとに数えられる。
  - 予算アラートは知らせるだけで、支出を止めない（[Budgets](https://docs.cloud.google.com/billing/docs/how-to/budgets)）。
  - spend cap の予算は、上限に達すると Cloud Run を止める。止まるのは、次の期間になるか、手で解除するまでである。保存しているリソースの料金は止まらない（[Spend cap budgets](https://docs.cloud.google.com/billing/docs/how-to/budgets-spend-caps)）。
  - 費用のデータは遅れて予算に届くので、予算が反応する前に金額を超えることがある。

## Decision（決定）

**シークレットは Secret Manager に置く**

- `DATABASE_URL`、`MONGODB_URI`、`LINE_CHANNEL_SECRET` は Secret Manager に置く。Cloud Run はこれらを環境変数として読む。環境変数のシークレットは、インスタンスが起動するときに読まれる（[Configure secrets](https://docs.cloud.google.com/run/docs/configuring/services/secrets)）。
- サービスの ID には `roles/secretmanager.secretAccessor` だけを与える。
- 変えた後は、古いシークレットのバージョンを無効にするか破棄する。これで無料枠（有効なバージョン 6 個、アクセス月 10,000 回）に収まる（[Secret Manager pricing](https://cloud.google.com/secret-manager/pricing)）。
- 残りの変数（`TEAM_ID`、`FRONTEND_ORIGIN`、`FRONTEND_DIR`、`ORGANIZER_LINE_USER_IDS`、`LINE_CHANNEL_ID`、`LINE_REDIRECT_URI`）は、Cloud Run のサービスに手で 1 回設定する。
- デプロイはイメージだけを変えるので、これらの値が公開リポジトリに出ることはない。

**マイグレーションは GitHub Actions から、デプロイの前に流す**

- デプロイのワークフロー（#29）は、Neon の直接接続に golang-migrate を流し、その後で新しいイメージをデプロイする。マイグレーションが失敗したら、何もデプロイしない。
- 直接接続の文字列は GitHub Actions のシークレットに置く。ログには出さない。
- 各マイグレーションは、デプロイが終わるまで動いている古いイメージでも動くようにする。消すか名前を変える変更は 2 回のデプロイに分ける：先に足し、後で消す。

**DB のネットワークは開け、認証情報と TLS で守る**

- Atlas の IP アクセスリストには `0.0.0.0/0` を入れる。Neon は既定のままで、どこからでも接続を受ける。
- どちらも、次のもので守る：
  - すべての接続に TLS を使う。
  - 長くランダムなパスワードを使い、Secret Manager と GitHub Actions のシークレットにだけ置く。
  - DB のユーザーには、backend に必要な権限だけを与える。Atlas では、tasuki のデータベースへの `readWrite` だけ。

**GitHub Actions は Workload Identity Federation でサインインする**

- デプロイのワークフローは、`google-github-actions/auth` を直接の Workload Identity Federation で使う。このアクションが推奨している方法である（[google-github-actions/auth](https://github.com/google-github-actions/auth)）。
- サービスアカウントの鍵は作らない。
- プロバイダは、このリポジトリの `main` ブランチからのトークンだけを受け付ける。
- これを含む IAM に料金はかからない（[IAM billing questions](https://docs.cloud.google.com/iam/docs/billing-questions)）。

**毎月の費用は少額で、上限を付ける**

- Cloud Run のインターネットへの送信にかかる、毎月の少額の費用は受け入れる。それ以外はすべて無料枠に収める。
- tasuki のプロジェクトには、Cloud Run に月 ¥100 の spend cap の予算を付ける。予算アラートは、同じ金額の 50%、90%、100% でメールを送る。
- 請求先アカウントは tasuki だけに使う。ほかのプロジェクトが無料枠を使ってしまわないようにするためである。
- Cloud Run は ADR-0018 の設定のままにする：リクエストごとの課金、最小インスタンス数 0、上限を付けた最大インスタンス数。
- 次のものは、どれも費用がかかるので有効にしない：Cloud NAT、VPC コネクタ、ロードバランサ、Cloud CDN、予約した IP アドレス、Artifact Registry の脆弱性スキャン。
- イメージは Cloud Build ではなく、公開リポジトリなら無料の GitHub Actions でビルドする。
- Artifact Registry にはクリーンアップポリシーを付け、最新の 10 個のイメージだけを残す。これで保存量が無料の 0.5 GiB に収まる（[Artifact Registry pricing](https://cloud.google.com/artifact-registry/pricing)）。

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| シークレットを Cloud Run のただの環境変数にする | コンソールでサービスを見られる人には値が見え、サービスのリビジョンの履歴にも残る。この規模なら Secret Manager に費用はかからない |
| シークレットを GitHub Actions にだけ置き、デプロイのたびに `gcloud run deploy` に渡す | デプロイのたびに、シークレットがワークフローを通る。公開しているワークフローのファイルを間違えると、値が出てしまう |
| backend の起動時にマイグレーションを流す | 複数のインスタンスが同時に起動すると競合する。最小インスタンス数が 0 なので Cloud Run はよくインスタンスを起動し（ADR-0018）、起動のたびに確認の時間もかかる |
| Cloud Run のジョブでマイグレーションを流す | 2 つ目のイメージか入口が要り、直接接続の文字列も Secret Manager に置く必要がある。準備が増えるのに、ワークフローの 1 ステップ（これも費用がかからない）と比べた利点がない |
| マイグレーションを手で流す | 忘れやすい。イメージとスキーマがずれるおそれがある |
| 送信元 IP を固定し（VPC、Cloud NAT、予約した IP）、Atlas でその IP だけを許す | 時間ごとに費用がかかる。Cloud NAT はゲートウェイの時間と処理した GiB ごとに、予約した IP も時間ごとに料金がかかる（[Cloud NAT pricing](https://cloud.google.com/nat/pricing)）。しかも守れるのは Atlas だけで、Neon は Scale プランなしではその IP を使えない |
| 有料の Atlas クラスタでプライベートエンドポイントを使う、または Neon Scale で IP Allow を使う | MVP に月額がかかる |
| サービスアカウントの鍵（JSON）を GitHub Actions のシークレットに置く | パスワードと同じように扱い、入れ替える必要がある長期の認証情報になる。Google は、ユーザーが管理する鍵を避けることを推奨している（[Best practices for service account keys](https://docs.cloud.google.com/iam/docs/best-practices-for-managing-service-account-keys)） |
| spend cap を付けず、予算アラートだけにする | アラートは支出を止めない。アクセスが急に増えると、誰かが対応するまで費用がかかり続ける |
| 予算に達したら請求先を自動で外す（Pub/Sub と関数） | Google は、リソースが復元できない形で消えることがあると警告している（[Disable billing with notifications](https://docs.cloud.google.com/billing/docs/how-to/disable-billing-with-notifications)）。費用のデータが遅れるので、反応も遅い |
| 送信の料金を避けるため、frontend を無料の静的ホスティングから配る | ADR-0018 の 1 つのオリジンが崩れる。セッションの Cookie と `Origin` の確認はそれに頼っている（ADR-0014）。しかも API の応答には料金がかかったままである |

## Consequences（影響）

- 無料枠に収まる限り、シークレット、マイグレーション、ネットワーク、CI のサインインに追加の費用はかからない。
- **本番は厳密には無料ではない。** 日本への Cloud Run の送信に、毎月少額がかかる。実際の金額は、公開した後に請求レポートで確かめる。
- **spend cap に達すると、tasuki は全員にとって止まる。** 止まるのは、次の月になるか、手で上限を解除するまでである。Artifact Registry に保存したイメージは、無料枠を超えていれば料金がかかり続ける。
- 費用のデータが遅れるので、Cloud Run が止まる前に、その月の費用が ¥100 を少し超えることがある。
- **DB にはどの IP アドレスからでも届く。** 接続文字列が漏れれば、それだけで入られる。ログに出したり、コミットしたり、貼り付けたりしてはいけない。漏れたら、すぐにパスワードを変える。ネットワークを絞れるプランに払うようになったら、この決定を見直す。
- シークレットを変えるには、新しいバージョンを足し、新しいリビジョンをデプロイする。環境変数のシークレットは起動時に読まれるからである。その後で古いバージョンを無効にする。
- シークレットではないサービスの設定は、コードにするまで Cloud Run のサービスの上にだけある。インフラのコード化は、手での準備が終わった後に行う。
- マイグレーションは、動いている古いイメージと両立しなければならない。このルールは、#28 で書くマイグレーションの手順に入れる。
- 準備の issue の順番は次のとおりになる：#200（Google Cloud と Workload Identity Federation）、#25 と #198（DB）、#199（LINE）、#28（シークレットと手順書）、#29（デプロイのワークフロー）。

</details>
