# Rules for app/

Every change under `app/` (the backend and the frontend) follows these rules, in addition to each app's `TECH.md`. Issue: #379.

## 1. Environment-dependent values come from environment variables

- A value that changes between environments is read from an environment variable and never hard-coded. This applies to both apps. Examples are URLs, origins, ports, hosts, credentials, IDs, and switches.
  - Backend: read it in `internal/config` (`app/backend/TECH.md`).
  - Frontend: the app calls the API at the relative path `/api` (ADR-0014, ADR-0018), so it needs no URL. If a value is needed, use a `VITE_` variable. Vite writes it into the built files, where anyone can read it, so a secret never goes there.
- If a hard-coded value is really needed, write an English comment next to it that says why it is needed, and why no alternative could be found.
- These are not covered: tests, and the local-only values in `infra/compose*.yaml`.

## 2. Constants say why they are constants

A constant is allowed. Each one has an English comment that says three things:

1. Why it is a constant.
2. What the alternative is, such as an environment variable or a setting.
3. Why the alternative was not chosen.

This covers named constants: Go `const` and package-level `var` that hold fixed values, and TypeScript module-level `const` that hold fixed values. It also covers a value with a meaning written inline, such as a timeout or a limit; such a value becomes a named constant with this comment. Unit conversions (`60`, `3600`) are not covered.

```go
// SessionLifetime is how long a session lasts (ADR-0014).
// Why a constant: it is a rule of the login design, the same in every environment.
// Alternative: a SESSION_LIFETIME environment variable.
// Not chosen: no environment needs a different value, and a wrong one would log users out early.
const SessionLifetime = 30 * 24 * time.Hour
```

## 3. Go names

- One-letter names are not allowed. The only exceptions are those in `app/backend/TECH.md` (Naming): `t` for `*testing.T`, `w` for `http.ResponseWriter`, and `i` for the index of a loop of a few lines.
- This includes parameters in interface methods and type parameters, which golangci-lint's `varnamelen` does not check.
- Every other name is meaningful and in lowerCamelCase. Exported names start with a capital letter, because Go exports only those. Initialisms keep one case, as in `userID` and `ServeHTTP` (Go Code Review Comments).

## 4. Tests run in Docker

- Tests run only in the Dev Containers, never with Go, Node, or pnpm installed on the host.
  - Backend: `make test` and `make test-integration` in `backend-dev`.
  - Frontend: `pnpm test` in `frontend-dev`.
- Each PR records the commands that were run in Docker and their results, under how it was checked.

## 5. Static analysis passes

- Backend: `make lint` (golangci-lint) and `go vet ./...` pass.
- Frontend: `pnpm lint` (oxlint), `pnpm format:check` (oxfmt), and `tsc -b` pass.
- Each app's `TECH.md` is followed.

---

<details>
<summary>日本語</summary>

# app/ のルール

`app/` の下（backend と frontend）のすべての変更は、それぞれのアプリの `TECH.md` に加えて、このルールに従う。Issue：#379。

## 1. 環境によって変わる値は環境変数から読む

- 環境ごとに変わる値は環境変数から読み、ハードコーディングしない。両方のアプリに当てはまる。URL、オリジン、ポート、ホスト、認証情報、ID、切り替えなどがそれである。
  - backend：`internal/config` で読む（`app/backend/TECH.md`）。
  - frontend：API は相対パス `/api` で呼ぶので（ADR-0014、ADR-0018）、URL は要らない。値が要るときは `VITE_` の変数を使う。Vite はその値をビルドしたファイルに書き込み、誰でも読めるので、シークレットは決して入れない。
- どうしてもハードコーディングが要るときは、その横に英語のコメントで、なぜ要るのか、なぜ代替案を出せなかったのかを書く。
- テストと、`infra/compose*.yaml` にあるローカル専用の値は対象外とする。

## 2. 定数には、定数にする理由を書く

定数は使ってよい。ただし、どの定数にも英語のコメントで次の 3 つを書く。

1. なぜ定数にするのか。
2. 代替案は何か（環境変数や設定など）。
3. 代替案を選ばなかった理由。

対象は名前の付いた定数である。Go の `const` と、決まった値を持つパッケージレベルの `var`、TypeScript の、決まった値を持つモジュールレベルの `const` がそれである。タイムアウトや上限のように、意味のある値をその場に直接書いたものも対象になる。そういう値は名前の付いた定数にして、このコメントを付ける。単位の換算（`60`、`3600`）は対象外とする。

## 3. Go の名前

- 1 文字の名前は使わない。例外は `app/backend/TECH.md`（Naming）にあるものだけである：`*testing.T` の `t`、`http.ResponseWriter` の `w`、数行のループの添え字の `i`。
- interface のメソッドの引数と型パラメータもこれに含む。golangci-lint の `varnamelen` はこれらを確かめない。
- ほかの名前は、意味のある lowerCamelCase にする。公開する名前は大文字で始める。Go は大文字で始まる名前しか公開しないからである。略語は大文字か小文字のどちらかにそろえる（`userID`、`ServeHTTP`。Go Code Review Comments）。

## 4. テストは Docker で実行する

- テストは Dev Container の中だけで実行する。ホストに入れた Go、Node、pnpm では実行しない。
  - backend：`backend-dev` で `make test` と `make test-integration`。
  - frontend：`frontend-dev` で `pnpm test`。
- 各 PR には、Docker で実行したコマンドとその結果を、確かめたこととして残す。

## 5. 静的解析を通す

- backend：`make lint`（golangci-lint）と `go vet ./...` を通す。
- frontend：`pnpm lint`（oxlint）、`pnpm format:check`（oxfmt）、`tsc -b` を通す。
- それぞれのアプリの `TECH.md` に従う。

</details>
