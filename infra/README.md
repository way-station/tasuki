# tasuki infra

The Docker setup of tasuki: the containers for local development, the production image, and the tools for setting up production. Everything here runs in Docker, so nothing else is installed on your machine.

## What is here

| Path | What it is |
| --- | --- |
| [`compose.yaml`](compose.yaml) | MongoDB, PostgreSQL, and the production image (`backend`). The Compose project is named `tasuki` |
| [`compose.dev.yaml`](compose.dev.yaml) | Added on top of `compose.yaml`: the Dev Containers (`backend-dev`, `frontend-dev`) and, under the `infra` profile, the production tools (`gcloud`, `terraform`) |
| [`docker/backend/Dockerfile`](docker/backend/Dockerfile) | The backend's Dev Container (`dev` stage) and the production image, which builds the frontend and serves it from the backend (ADR-0018) |
| [`docker/frontend/Dockerfile`](docker/frontend/Dockerfile) | The frontend's Dev Container. There is no production frontend image |
| [`docker/postgres/init/`](docker/postgres/init) | Runs once when the PostgreSQL volume is created; it adds the test database |
| [`.env.example`](.env.example) | Host ports. Copy it to `infra/.env` if a port is already in use |
| [`Makefile`](Makefile) | The production tools: gcloud and Terraform |

The Dev Containers are started from each app: see [`app/backend/README.md`](../app/backend/README.md) and [`app/frontend/README.md`](../app/frontend/README.md).

## Run the production image on your machine

In `app/backend`:

```sh
make up     # builds the production image and starts it with MongoDB and PostgreSQL
make down
```

It uses the same databases as the Dev Container, and the image does not apply migrations (ADR-0021). Run `make db-migrate` and `make db-seed` in the backend's Dev Container first.

Open `http://localhost:8080` (or your `BACKEND_PORT`). The image serves the API and the built frontend from one origin, as in production.

## Production tools

Run these in `infra`. Each target starts a container for one command and removes it afterwards. They are used to set up production (ADR-0021).

| Command | What it does |
| --- | --- |
| `make gcloud-login` | Signs gcloud in. Open the printed URL in your browser, then paste the code back |
| `make gcloud ARGS="..."` | Runs any gcloud command, for example `make gcloud ARGS="run services list"` |
| `make adc-login` | Creates the Application Default Credentials that Terraform uses, in the same way |
| `make tf-init` / `tf-plan` / `tf-apply` | Runs Terraform in `TF_DIR` (default `terraform`, relative to `infra`) |
| `make gcloud-forget` | Deletes the sign-in volume, so no credentials stay on your machine |

There are no Terraform files yet: infrastructure as code comes after the manual setup.

## Be careful

- **Sign-in lives in the `tasuki_gcloud-config` volume**, shared by both tools, never in your `~/.config`. Remove it with `make gcloud-forget` when you are done.
- **Do not run `docker compose down --volumes`** for this project. The volumes belong to the same Compose project, so it also deletes the MongoDB and PostgreSQL data.
- **Never commit Terraform state or `.tfvars`.** They can hold secrets and private values, and the repository is public. [`.gitignore`](.gitignore) ignores them.

---

<details>
<summary>日本語</summary>

# tasuki infra

tasuki の Docker の構成である。ローカルの開発のためのコンテナ、本番のイメージ、本番を準備するための道具がある。すべて Docker で動くので、手元のマシンには、ほかに何も入れない。

## ここにあるもの

| パス | 中身 |
| --- | --- |
| [`compose.yaml`](compose.yaml) | MongoDB、PostgreSQL、本番のイメージ（`backend`）。Compose のプロジェクト名は `tasuki` |
| [`compose.dev.yaml`](compose.dev.yaml) | `compose.yaml` に重ねて使う。Dev Container（`backend-dev`、`frontend-dev`）と、`infra` プロファイルの本番の道具（`gcloud`、`terraform`） |
| [`docker/backend/Dockerfile`](docker/backend/Dockerfile) | backend の Dev Container（`dev` ステージ）と本番のイメージ。本番のイメージは frontend をビルドし、backend から配る（ADR-0018） |
| [`docker/frontend/Dockerfile`](docker/frontend/Dockerfile) | frontend の Dev Container。本番の frontend のイメージはない |
| [`docker/postgres/init/`](docker/postgres/init) | PostgreSQL のボリュームを作るときに 1 回だけ動き、テスト用のデータベースを足す |
| [`.env.example`](.env.example) | ホストのポート。使われているポートがあれば、`infra/.env` にコピーして変える |
| [`Makefile`](Makefile) | 本番の道具：gcloud と Terraform |

Dev Container は、それぞれのアプリから起動する。[`app/backend/README.md`](../app/backend/README.md) と [`app/frontend/README.md`](../app/frontend/README.md) を見ること。

## 本番のイメージを手元で動かす

`app/backend` で：

```sh
make up     # 本番のイメージを作り、MongoDB と PostgreSQL と一緒に起動する
make down
```

Dev Container と同じデータベースを使い、イメージはマイグレーションを流さない（ADR-0021）。先に backend の Dev Container で `make db-migrate` と `make db-seed` を実行する。

`http://localhost:8080`（または `BACKEND_PORT`）を開く。本番と同じく、API とビルドした frontend を 1 つのオリジンから配る。

## 本番の道具

`infra` で実行する。どのターゲットも、1 つのコマンドのためにコンテナを起動し、終わったら消す。本番の準備（ADR-0021）に使う。

| コマンド | すること |
| --- | --- |
| `make gcloud-login` | gcloud にサインインする。表示された URL をブラウザで開き、コードを貼り戻す |
| `make gcloud ARGS="..."` | 好きな gcloud のコマンドを実行する。例：`make gcloud ARGS="run services list"` |
| `make adc-login` | Terraform が使う Application Default Credentials を、同じ方法で作る |
| `make tf-init` / `tf-plan` / `tf-apply` | `TF_DIR`（既定は `terraform`、`infra` からの相対パス）で Terraform を実行する |
| `make gcloud-forget` | サインインのボリュームを消し、認証情報を手元に残さない |

Terraform のファイルはまだない。インフラのコード化は、手での準備の後に行う。

## 気をつけること

- **サインインの情報は `tasuki_gcloud-config` ボリュームにある。** 2 つの道具で共有し、`~/.config` には置かない。使い終わったら `make gcloud-forget` で消す。
- **このプロジェクトで `docker compose down --volumes` を実行しない。** ボリュームは同じ Compose プロジェクトに属するので、MongoDB と PostgreSQL のデータまで消える。
- **Terraform の state と `.tfvars` はコミットしない。** シークレットや公開したくない値が入ることがあり、リポジトリは公開されている。[`.gitignore`](.gitignore) で無視している。

</details>
