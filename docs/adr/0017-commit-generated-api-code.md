# ADR-0017: Commit generated API code

- Status: Proposed
- Type: technical
- Date: 2026-10-24
- Issue: #177

## Context

The API is described spec-first in `app/backend/api/openapi.yaml` (ADR-0010). From it, oapi-codegen generates the Go types and the chi server interface that the controllers implement (ADR-0011), and openapi-typescript generates the TypeScript types the frontend uses (ADR-0012). ADR-0010 left to the implementation whether the generated code is committed and how CI checks that it is up to date.

The forces:

- `go build` does not run code generators. If the generated Go code is not in the repository, a fresh clone does not build, and the editor shows errors, until someone runs the generator.
- The generated code is what the controllers and the frontend compile against. A change to the document changes it, and that change is what reviewers need to see to judge the impact of an API change.
- Generated code that is older than the document would let the backend and the frontend disagree with the document without any error.
- All toolchains run in containers, not on the host (`app/backend/TECH.md`, `app/frontend/TECH.md`).

## Decision

**The generated files are committed, and CI fails when they are out of date.**

| | Generator | Generated file |
| --- | --- | --- |
| Backend | oapi-codegen (types and chi server) | `app/backend/internal/presentation/api/openapi.gen.go` |
| Frontend | openapi-typescript | `app/frontend/src/api/schema.gen.ts` |

- Generated files have `.gen.` in their names and start with a "generated, do not edit" header. They are never edited by hand; a change is made in `openapi.yaml` and the files are generated again.
- The generators' versions are pinned in the repository: oapi-codegen as a Go tool in `go.mod`, openapi-typescript in `package.json`. Generation runs in the Dev Containers, with one command per side (for example `go generate ./...` and `pnpm generate:api`).
- **CI regenerates both files and fails if `git diff --exit-code` shows a difference.** The backend and the frontend workflows also run when `app/backend/api/openapi.yaml` changes.
- A change to `openapi.yaml` and the regenerated files go in the same pull request.
- Linters and formatters skip the generated files. GitHub marks them as generated (`.gitattributes`, `linguist-generated`), so they are collapsed in pull request diffs but can still be opened.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Generate in every build and do not commit | Every place that builds (Dev Containers, CI, Docker images) needs the generators and the generation step. A fresh clone does not build, and the editor shows errors, until the generator runs. The effect of an API change on the generated code is not visible in the pull request |
| Commit without a CI check | Generated code could fall behind the document without any error, which is the drift spec-first is meant to prevent (ADR-0010) |

## Consequences

- A fresh clone builds, and the editor works, without running a generator first.
- An API change shows its effect on the Go and TypeScript code in the pull request. The generated part is collapsed so it does not hide the hand-written changes.
- Pull requests that change the API are larger, because they include the regenerated files.
- Forgetting to regenerate fails CI instead of failing silently.
- Two pull requests that both change `openapi.yaml` can conflict in the generated files; the fix is to merge the document and generate again, not to resolve the generated files by hand.
- `.github/workflows/frontend.yml` must add `app/backend/api/openapi.yaml` to its paths; the generation commands and the CI check are set up in #30.

---

<details>
<summary>日本語</summary>

# ADR-0017: 生成した API のコードをコミットする

- Status: Proposed
- Type: technical
- Date: 2026-10-24
- Issue: #177

## Context（背景）

API は `app/backend/api/openapi.yaml` にスペックファーストで書く（ADR-0010）。そこから oapi-codegen が、コントローラが実装する Go の型と chi のサーバーの interface を生成し（ADR-0011）、openapi-typescript が frontend の使う TypeScript の型を生成する（ADR-0012）。生成したコードをコミットするか、それが最新であることを CI でどう確かめるかは、ADR-0010 で実装に任せていた。

考慮したこと：

- `go build` はコードの生成を実行しない。生成した Go のコードがリポジトリになければ、clone した直後はビルドが通らず、誰かが生成するまでエディタにエラーが出る
- 生成したコードは、コントローラと frontend がコンパイルするときの相手。ドキュメントを変えるとこれが変わり、レビュアーが API の変更の影響を判断するのに見る必要があるのも、この変化
- 生成したコードがドキュメントより古いと、backend と frontend がエラーなしにドキュメントと食い違う
- ツールはすべてホストではなくコンテナで動かす（`app/backend/TECH.md`、`app/frontend/TECH.md`）

## Decision（決定）

**生成したファイルはコミットし、古くなっていたら CI を失敗させる。**

| | 生成するツール | 生成するファイル |
| --- | --- | --- |
| backend | oapi-codegen（型と chi のサーバー） | `app/backend/internal/presentation/api/openapi.gen.go` |
| frontend | openapi-typescript | `app/frontend/src/api/schema.gen.ts` |

- 生成したファイルは名前に `.gen.` を含み、先頭に「生成したもの。編集しない」という見出しを持つ。手で編集しない。変えるときは `openapi.yaml` を変えて生成し直す
- 生成するツールのバージョンはリポジトリで固定する：oapi-codegen は `go.mod` の Go の tool として、openapi-typescript は `package.json` で。生成は Dev Container の中で、それぞれ 1 つのコマンドで行う（例：`go generate ./...` と `pnpm generate:api`）
- **CI は両方のファイルを生成し直し、`git diff --exit-code` で差分が出たら失敗する。** backend と frontend のワークフローは、`app/backend/api/openapi.yaml` が変わったときにも動く
- `openapi.yaml` の変更と、生成し直したファイルは同じ PR に入れる
- linter と formatter は生成したファイルを対象にしない。GitHub では生成したものとして扱う（`.gitattributes` の `linguist-generated`）ので、PR の差分では折りたたまれるが、開いて見ることはできる

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| ビルドのたびに生成し、コミットしない | ビルドするすべての場所（Dev Container、CI、Docker のイメージ）に生成するツールと生成の手順が要る。clone した直後はビルドが通らず、生成するまでエディタにエラーが出る。API の変更が生成したコードに与える影響が PR に出ない |
| コミットするが CI で確かめない | 生成したコードがエラーなしにドキュメントより古くなりうる。スペックファーストが防ぎたい食い違いそのもの（ADR-0010） |

## Consequences（影響）

- clone した直後から、生成しなくてもビルドでき、エディタも使える
- API の変更が Go と TypeScript のコードに与える影響が PR に出る。生成した部分は折りたたまれるので、手で書いた変更が埋もれない
- API を変える PR は、生成し直したファイルを含むので大きくなる
- 生成し忘れると、黙って食い違うのではなく CI が失敗する
- `openapi.yaml` を変える PR が 2 つあると、生成したファイルで衝突することがある。直し方は、生成したファイルを手で解消するのではなく、ドキュメントをマージしてから生成し直すこと
- `.github/workflows/frontend.yml` の paths に `app/backend/api/openapi.yaml` を足す必要がある。生成のコマンドと CI の確認は #30 で用意する

</details>
