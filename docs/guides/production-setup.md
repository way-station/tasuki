# Setting up production

- Decisions: [ADR-0018](../adr/0018-hosting.md) (hosting), [ADR-0021](../adr/0021-production-configuration-and-deployment.md) (configuration, deployment, costs)
- Issues: #200 (Google Cloud), #25 (Neon), #198 (MongoDB Atlas), #199 (LINE), #28 (this guide), #29 (the deploy workflow)
- Workflow: [`.github/workflows/deploy.yml`](../../.github/workflows/deploy.yml)

This guide creates production by hand, once, in the order of ADR-0021. Afterwards every merge to `main` deploys by itself. Everything runs in Docker: the gcloud commands run in the `gcloud` container of `infra/` (`infra/README.md`), the database commands in the backend's Dev Container.

## 0. Names and values

Open a shell in the gcloud container and keep it open for the whole guide. Run this in `infra/`:

```sh
docker compose -f compose.yaml -f compose.dev.yaml --profile infra run --rm --entrypoint bash gcloud
```

In that shell, set these. The names are what this guide uses; change them if you like, but use the same ones everywhere, including the GitHub variables in step 9.

```sh
export PROJECT_ID=...                  # a new, globally unique project ID
export BILLING_ACCOUNT_ID=...          # from step 1
export REGION=asia-southeast1          # Singapore (ADR-0018)
export REPOSITORY=tasuki               # Artifact Registry repository
export SERVICE=tasuki                  # Cloud Run service
export RUNTIME_SA=tasuki-run@$PROJECT_ID.iam.gserviceaccount.com
```

## 1. Google Cloud and billing (#200)

1. Create a Google Cloud account in the console, and a **billing account in Japanese yen**. Its currency cannot be changed later (ADR-0021).
2. Use this billing account only for tasuki, because free tiers are counted per billing account.
3. Sign in and create the project:

```sh
gcloud auth login --no-launch-browser
gcloud projects create "$PROJECT_ID"
gcloud billing projects link "$PROJECT_ID" --billing-account="$BILLING_ACCOUNT_ID"
gcloud config set project "$PROJECT_ID"
gcloud services enable run.googleapis.com artifactregistry.googleapis.com secretmanager.googleapis.com \
  iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com
```

Do not enable Container Scanning (`containerscanning.googleapis.com`), Cloud Build, Cloud NAT, or a load balancer; each costs money (ADR-0021).

## 2. Budgets (#200)

In the console, under Billing → Budgets & alerts:

1. **Spend cap budget**: this project, the Cloud Run service, ¥100 a month. Its alerts are fixed at 50%, 80%, and 100%.
2. **Alerts-only budget**: the whole billing account, ¥100 a month, alerts at 50%, 90%, and 100%.

## 3. Artifact Registry (#200)

```sh
gcloud artifacts repositories create "$REPOSITORY" --repository-format=docker --location="$REGION"
gcloud artifacts repositories set-cleanup-policies "$REPOSITORY" --location="$REGION" \
  --policy=artifact-registry/cleanup-policy.json --no-dry-run
```

The policy keeps the latest 10 images and deletes the rest, so storage stays within the free 0.5 GiB. It takes effect within about a day.

## 4. Databases and LINE (#25, #198, #199)

Create these in each service's console, and keep the values for the next steps. Never paste them into an issue, a PR, or a commit.

| From | Value | Used as |
| --- | --- | --- |
| Neon (`aws-ap-southeast-1`) | The **direct** connection string (not `-pooler`, ADR-0015), with `sslmode=require` and `pool_max_conns` added | `DATABASE_URL` |
| Neon | The same direct connection string **without** `pool_max_conns`: golang-migrate would send it to PostgreSQL as a setting, which fails | Migrations: step 6 and `MIGRATION_DATABASE_URL` in step 9 |
| Neon | `SHOW max_connections;` | Choosing the maximum instances in step 7 |
| Atlas (M0, AWS `ap-southeast-1`) | The connection string of a user with `readWrite` on `tasuki` only; IP access list `0.0.0.0/0` | `MONGODB_URI` |
| LINE Developers (LINE Login channel, `openid` and `profile`) | Channel ID, channel secret | `LINE_CHANNEL_ID`, `LINE_CHANNEL_SECRET` |
| LINE Developers | Your LINE user ID (Basic settings) | `ORGANIZER_LINE_USER_IDS` |

## 5. Secrets and the service's identity (#28)

```sh
gcloud iam service-accounts create tasuki-run --display-name="tasuki on Cloud Run"
for secret in database-url mongodb-uri line-channel-secret; do
  echo "Paste the value of $secret, then press Ctrl-D twice without Enter:"
  gcloud secrets create "$secret" --replication-policy=automatic --data-file=-
  gcloud secrets add-iam-policy-binding "$secret" \
    --member="serviceAccount:$RUNTIME_SA" --role=roles/secretmanager.secretAccessor
done
```

Pressing Enter before Ctrl-D adds a newline to the value. The service's identity gets only `secretAccessor` (ADR-0021).

## 6. Migrations and the team's row (#28)

In the backend's Dev Container (`app/backend`), with the **direct** connection string. A leading space keeps the line out of the shell history in most shells.

```sh
 export DATABASE_URL='<Neon direct connection string>'
export TEAM_ID=$(cat /proc/sys/kernel/random/uuid)   # keep it for step 7
export TEAM_NAME='<your team name>'
make db-migrate
make db-seed
unset DATABASE_URL
```

## 7. The Cloud Run service (#200)

Choose the maximum instances so that maximum instances × `pool_max_conns` stays below Neon's `max_connections` (ADR-0015, ADR-0018). The first deployment uses Google's sample image; the deploy workflow replaces it.

```sh
export MAX_INSTANCES=...
export TEAM_ID=...                     # from step 6
gcloud run deploy "$SERVICE" --region="$REGION" \
  --image=us-docker.pkg.dev/cloudrun/container/hello \
  --service-account="$RUNTIME_SA" --allow-unauthenticated \
  --min-instances=0 --max-instances="$MAX_INSTANCES" \
  --set-secrets=DATABASE_URL=database-url:1,MONGODB_URI=mongodb-uri:1,LINE_CHANNEL_SECRET=line-channel-secret:1 \
  --set-env-vars=TEAM_ID="$TEAM_ID",MONGODB_DATABASE=tasuki
export URL=$(gcloud run services describe "$SERVICE" --region="$REGION" --format='value(status.url)')
gcloud run services update "$SERVICE" --region="$REGION" \
  --update-env-vars="^;^FRONTEND_ORIGIN=$URL;LINE_REDIRECT_URI=$URL/api/v1/auth/line/callback;LINE_CHANNEL_ID=<channel ID>;ORGANIZER_LINE_USER_IDS=<LINE user IDs, comma-separated>"
```

- Secrets are pinned to version `1`, not `latest`, so a new version takes effect only with a new revision (ADR-0021).
- `^;^` makes `;` the separator, because `ORGANIZER_LINE_USER_IDS` contains commas.
- `DEV_LOGIN_ENABLED` is never set in production (ADR-0009). `FRONTEND_DIR` is set in the image.
- Register `$URL/api/v1/auth/line/callback` as the callback URL of the LINE Login channel, then publish the channel (#199).

## 8. Workload Identity Federation (#200)

```sh
export PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
gcloud iam workload-identity-pools create github --location=global --display-name=GitHub
gcloud iam workload-identity-pools providers create-oidc tasuki --location=global \
  --workload-identity-pool=github --issuer-uri=https://token.actions.githubusercontent.com \
  --attribute-mapping=google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.ref=assertion.ref \
  --attribute-condition="assertion.repository == 'way-station/tasuki' && assertion.ref == 'refs/heads/main'"
export PRINCIPAL="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/github/attribute.repository/way-station/tasuki"
gcloud artifacts repositories add-iam-policy-binding "$REPOSITORY" --location="$REGION" \
  --member="$PRINCIPAL" --role=roles/artifactregistry.writer
gcloud run services add-iam-policy-binding "$SERVICE" --region="$REGION" \
  --member="$PRINCIPAL" --role=roles/run.developer
gcloud iam service-accounts add-iam-policy-binding "$RUNTIME_SA" \
  --member="$PRINCIPAL" --role=roles/iam.serviceAccountUser
echo "projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/github/providers/tasuki"
```

The provider accepts only tokens from this repository's `main` branch, and no service account key exists (ADR-0021). The last line prints the value for `GCP_WORKLOAD_IDENTITY_PROVIDER`.

## 9. GitHub (#29)

In the repository's Settings → Secrets and variables → Actions:

| Kind | Name | Value |
| --- | --- | --- |
| Secret | `MIGRATION_DATABASE_URL` | The Neon direct connection string |
| Variable | `GCP_REGION` | `$REGION` |
| Variable | `ARTIFACT_REPOSITORY` | `$REPOSITORY` |
| Variable | `CLOUD_RUN_SERVICE` | `$SERVICE` |
| Variable | `GCP_WORKLOAD_IDENTITY_PROVIDER` | The line printed in step 8 |
| Variable | `GCP_PROJECT_ID` | `$PROJECT_ID`. **Set it last**: the deploy workflow runs only once it exists |

Then run the `deploy` workflow by hand from `main` (Actions → deploy → Run workflow). It ends by checking `/health`.

## 10. Afterwards

**Changing a secret.** Add a version, point the service at it, then disable the old one:

```sh
gcloud secrets versions add database-url --data-file=-
gcloud run services update "$SERVICE" --region="$REGION" --update-secrets=DATABASE_URL=database-url:2
gcloud secrets versions disable 1 --secret=database-url
```

Keep at most 6 active versions in all, the free tier of Secret Manager.

**Writing migrations.** The workflow applies migrations while the previous image is still serving. Every migration must work with that image: a change that removes or renames something is split into two deployments, adding first and removing later (ADR-0021).

**Signing out.** Leave the gcloud shell, then run `make gcloud-forget` in `infra/`.

---

<details>
<summary>日本語</summary>

# 本番を準備する

- 決定：[ADR-0018](../adr/0018-hosting.md)（ホスティング）、[ADR-0021](../adr/0021-production-configuration-and-deployment.md)（設定、デプロイ、費用）
- Issue：#200（Google Cloud）、#25（Neon）、#198（MongoDB Atlas）、#199（LINE）、#28（この手順書）、#29（デプロイのワークフロー）
- ワークフロー：[`.github/workflows/deploy.yml`](../../.github/workflows/deploy.yml)

この手順書は、ADR-0021 の順番で、本番を手で 1 回だけ作る。その後は、`main` への merge のたびに自動でデプロイされる。すべて Docker で動かす。gcloud のコマンドは `infra/` の `gcloud` コンテナ（`infra/README.md`）で、DB のコマンドは backend の Dev Container で実行する。

## 0. 名前と値

gcloud のコンテナでシェルを開き、手順の間ずっと開いておく。`infra/` で実行する：

```sh
docker compose -f compose.yaml -f compose.dev.yaml --profile infra run --rm --entrypoint bash gcloud
```

そのシェルで、英語の節の `export` を設定する。名前はこの手順書で使うもので、変えてもよいが、手順 9 の GitHub の変数も含めて、どこでも同じものを使う。

## 1. Google Cloud と請求先アカウント（#200）

1. コンソールで Google Cloud のアカウントと、**日本円の請求先アカウント**を作る。通貨は後から変えられない（ADR-0021）。
2. 無料枠は請求先アカウントごとに数えられるので、この請求先アカウントは tasuki だけに使う。
3. サインインし、英語の節のコマンドでプロジェクトを作り、API を有効にする。

Container Scanning（`containerscanning.googleapis.com`）、Cloud Build、Cloud NAT、ロードバランサは有効にしない。どれも費用がかかる（ADR-0021）。

## 2. 予算（#200）

コンソールの「お支払い → 予算とアラート」で：

1. **spend cap の予算**：このプロジェクト、Cloud Run のサービス、月 ¥100。アラートは 50%、80%、100% で固定。
2. **知らせるだけの予算**：請求先アカウント全体、月 ¥100、アラートは 50%、90%、100%。

## 3. Artifact Registry（#200）

英語の節のコマンドで、リポジトリを作り、クリーンアップポリシーを付ける。ポリシーは最新の 10 個のイメージを残して残りを消すので、保存量が無料の 0.5 GiB に収まる。反映には 1 日ほどかかる。

## 4. DB と LINE（#25、#198、#199）

それぞれのコンソールで作り、値を次の手順のために控える。issue、PR、commit には決して貼らない。何をどこから取り、何に使うかは、英語の節の表のとおりである。Neon は、backend もマイグレーションも**直接接続**（`-pooler` ではない、ADR-0015）を使う。ただし、マイグレーション用の文字列には `pool_max_conns` を付けない。golang-migrate がそれを PostgreSQL の設定として送り、失敗するからである。

## 5. シークレットとサービスの ID（#28）

英語の節のコマンドで、サービスアカウントを作り、3 つのシークレットを作って `secretAccessor` を与える。値を貼ったら、Enter を押さずに Ctrl-D を 2 回押す。先に Enter を押すと、値の末尾に改行が入る。

## 6. マイグレーションとチームの行（#28）

backend の Dev Container（`app/backend`）で、**直接接続**の文字列を使って、英語の節のコマンドを実行する。先頭に空白を入れると、多くのシェルでその行が履歴に残らない。`TEAM_ID` は手順 7 で使うので控える。

## 7. Cloud Run のサービス（#200）

最大インスタンス数は、最大インスタンス数 × `pool_max_conns` が Neon の `max_connections` を下回るように決める（ADR-0015、ADR-0018）。最初のデプロイには Google のサンプルのイメージを使い、デプロイのワークフローがそれを置き換える。コマンドは英語の節のとおり。

- シークレットは `latest` ではなくバージョン `1` に固定する。新しいバージョンは、新しいリビジョンでだけ効く（ADR-0021）。
- `ORGANIZER_LINE_USER_IDS` にカンマが入るので、`^;^` で区切り文字を `;` にする。
- 本番では `DEV_LOGIN_ENABLED` を決して設定しない（ADR-0009）。`FRONTEND_DIR` はイメージで設定されている。
- `$URL/api/v1/auth/line/callback` を LINE ログインのチャネルのコールバック URL に登録し、チャネルを公開する（#199）。

## 8. Workload Identity Federation（#200）

英語の節のコマンドで、プールとプロバイダを作り、権限を与える。プロバイダはこのリポジトリの `main` ブランチのトークンだけを受け付け、サービスアカウントの鍵は作らない（ADR-0021）。最後の行で、`GCP_WORKLOAD_IDENTITY_PROVIDER` の値が表示される。

## 9. GitHub（#29）

リポジトリの「Settings → Secrets and variables → Actions」で、英語の節の表のシークレットと変数を設定する。**`GCP_PROJECT_ID` は最後に設定する。** これがあるときだけ、デプロイのワークフローが動くからである。

その後、`main` から `deploy` ワークフローを手で実行する（Actions → deploy → Run workflow）。最後に `/health` を確かめる。

## 10. その後

- **シークレットを変える**：バージョンを足し、サービスをそれに向け、古いバージョンを無効にする（英語の節のコマンド）。有効なバージョンは、全部で 6 個までにする（Secret Manager の無料枠）。
- **マイグレーションを書く**：ワークフローは、前のイメージが動いている間にマイグレーションを流す。各マイグレーションは、そのイメージでも動くようにする。消すか名前を変える変更は、2 回のデプロイに分ける：先に足し、後で消す（ADR-0021）。
- **サインアウト**：gcloud のシェルを抜け、`infra/` で `make gcloud-forget` を実行する。

</details>
