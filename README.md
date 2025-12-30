# mngproj - Monorepo Polyglot Manager

**Status:** 🚧 In Development (Prototype Phase)

## 1. プロジェクト概要 (Project Overview)

`mngproj` は、多言語（Go, Python, Rust, Node.js 等）が混在するモノレポ開発において、ビルド・実行環境を抽象化・統一するためのCLIツールです。
各プロジェクトディレクトリに配置された設定ファイル (`mngproj.toml`) と、言語/ツールごとに定義された **プリセット設定 (Presets)** を組み合わせることで、複雑なビルド手順や環境依存の問題を解決します。
さらに、スクリプトの柔軟な実行、依存関係管理、並列実行、ホットリロードなどのワークフロー機能を提供します。

---

## 2. コア機能 (Core Features)

### 2.1 Project & Components
1つのプロジェクト（`mngproj.toml` が置かれた単位）は、**複数のコンポーネント** を持つことができます。
例: `backend` (Rust), `frontend` (React), `docs` (Python/MkDocs) が一つのリポジトリに共存。

### 2.2 Presets & Roles
標準的な技術スタックはプリセット (`presets/*.toml`) として定義されています。
各プリセットは以下の **Role** を持ち、競合時の優先順位を決定します。

1.  **Framework** (Score: 30): `django`, `react`, `nextjs` ...
2.  **Tool** (Score: 20): `docker`, `make` ...
3.  **Package Manager** (Score: 10): `uv`, `pip`, `npm`, `cargo` ...
4.  **Language** (Score: 0): `python`, `go`, `rust`, `node` ...

同じコマンド（例: `run`）が複数のプリセットで定義されている場合、よりスコアの高いRoleのコマンドが自動選択されます。

### 2.3 Dependency Management (Add & Sync)
`mngproj` は各コンポーネントの依存関係を `mngproj.toml` で宣言的に管理し、対応するマニフェストファイル（`requirements.txt` など）を自動生成・同期します。

### 2.4 Parallel Execution & Aggregated Logs (Up)
複数のコンポーネントを並列で起動し、それぞれのログをコンポーネント名でプレフィックス付けして統一的に表示できます。モノレポでの開発体験を向上させます。

### 2.5 Hot Reloading (Watch)
ファイルの変更を検知し、自動的にコンポーネントを再起動するホットリロード機能を提供します。開発中の迅速なフィードバックサイクルを実現します。

### 2.6 Isolation (Sandboxing)
パッケージマネージャによるインストールがグローバル環境を汚染しないよう、`mngproj` は自動的にローカルディレクトリ（例: `.libs`, `.npm-global`）へのインストールを強制します。

### 2.7 Multi-Platform Support
Windows, Linux, macOS での動作をサポートしています。OSに応じたシェルの切り替え（WindowsではPowerShell）や、OS固有のプリセット読み込みを自動的に行います。

---

## 3. 設定ファイル構成 (Configuration)

### 3.1 プロジェクト設定 (`mngproj.toml`)

```toml
[project]
name = "my-web-service"
description = "A full-stack web application"
root = "../" # (Optional) プロジェクトのルートディレクトリを明示的に指定。mngproj.tomlがあるディレクトリからの相対パス、または絶対パス。

# (Optional) ロールごとの優先順位をカスタマイズ
[resolution.role_priority]
tool = 100 # toolをframeworkより優先させる例

# --- Component Definition ---
[[components]]
name = "api"
# 複数のプリセットを組み合わせる (Mixin)
types = ["python", "uv", "django"]
path = "./backend"
# 任意のグループ名を付与してまとめて操作可能 (例: mngproj up backend)
groups = ["backend", "core"]

# コンポーネントが依存するパッケージ一覧
dependencies = ["flask==2.3.0", "requests"]

[components.env]
PORT = "8000"

# ユーザー定義スクリプト (最高優先度)
[components.scripts]
# スクリプト内で引数や環境変数をテンプレートとして利用可能
# 例: mngproj build api production -> echo building api for production
build = "echo building {{.Name}} for {{index .Args 0}}" 
deploy = "file:scripts/deploy.sh" # 外部シェルスクリプトファイルを指定

```
#### スクリプトのテンプレート機能と外部ファイル (Script Templating & External Files)
`components.scripts` 内のコマンド定義では、Goの `text/template` 構文を利用できます。
`{{.Args}}`: コマンドに渡された引数のスライス。`{{index .Args 0}}` で個別にアクセス可能。
`{{.Env.VAR_NAME}}`: コンポーネントの環境変数にアクセス。

また、`file:` プレフィックスを使用すると、外部ファイルに記述されたスクリプトを実行できます。
例: `deploy = "file:scripts/deploy.sh"` とすると、`project.root` または `mngproj.toml` のあるディレクトリからの相対パスで `scripts/deploy.sh` を探します。

### 3.2 プリセット設定 (`presets/*.toml`)
`presets/*.toml` ファイルは、`mngproj.toml` のコンポーネント設定と同様に `scripts` と `env` を定義できます。
さらに、`[metadata]` セクションには `manifest_file`, `required_tools`, `gitignore` を指定できます。

---

## 4. コマンド体系 (Commands)

| コマンド | 引数例 | 説明 |
| :--- | :--- | :--- |
| **`init`** | `[type]` | カレントディレクトリに `mngproj.toml` の雛形と `.gitignore` を生成します。`type` で言語を指定可能（例: `go`, `python`, `node`）。 |
| **`run`** | `[comp] [args...]` | コンポーネントを実行します。(例: `mngproj run api`) |
| **`build`** | `[comp] [args...]` | コンポーネントをビルドします。 |
| **`add`** | `[comp] [pkgs...]` | パッケージをコンポーネントの依存関係に追加し、`mngproj.toml` を更新、マニフェストファイルを同期します。(例: `mngproj add api flask`) |
| **`sync`** | `[comp]` | 指定された、または全てのコンポーネントのマニフェストファイルを更新し、依存関係を解決します。必要なツールのインストールチェックも行います。 |
| **`up`** | `[comp/group...]` | 指定されたコンポーネントまたはグループを並列で実行し、ログをプレフィックス付きで表示します。(例: `mngproj up api web`) |
| **`watch`** | `[comp...]` | コンポーネントのソースコード変更を監視し、自動的に再起動します。(例: `mngproj watch frontend`) |
| **`lfs`** | `[threshold_mb]` | 大容量ファイルを検出し、`.gitattributes` に Git LFS 設定を追加します。(例: `mngproj lfs 50`) |
| **`install-self`** | `(なし)` | 現在のソースコードから `mngproj` をビルドし、システムにインストールします。 |
| **`remove`** | `[comp] [pkgs...]` | パッケージをコンポーネントの依存関係から削除します。 |
| **`ls`** | `(なし)` | 現在のプロジェクト内のコンポーネント一覧を表示します。 |
| **`lsproj`** | `(なし)` | カレントディレクトリ以下の **全てのプロジェクト** (`mngproj.toml`) を再帰的に検索・表示します。 |
| **`info`** | `(なし)` | 現在のプロジェクト情報や読み込まれているプリセットパスを表示します。 |
| **`query`** | `(なし)` | コンポーネント情報をJSON形式で出力します (CI/CD連携用)。 |
| **`<script>`** | `<script> <comp> [args...]` | `mngproj.toml` で定義されたカスタムスクリプトを、指定されたコンポーネントで実行します。(例: `mngproj deploy api`) |

---

## 5. ディレクトリ構造 & プリセット

```text
~/Work/monorepo/
├── mngproj.toml        # プロジェクトルートに置かれる設定ファイル
├── presets/            # カスタムプリセット (Optional)
│   ├── languages/      # go.toml, python.toml ...
│   ├── frameworks/     # django.toml, react.toml ...
│   ├── managers/       # uv.toml, npm.toml ...
│   └── tools/          # docker.toml ...
└── services/
    ├── service-a/
    │   ├── mngproj.toml
    │   └── ...
    └── service-b/
        ├── mngproj.toml
        └── ...
```

### 利用可能なプリセット (Available Presets)
- **Languages:** `go`, `python`, `node`, `ts`, `rust`, `java`, `c++` (`clang`/`gcc`), `deno`, `bun`, `php`, `ruby`
- **Frameworks:** `nextjs`, `react`, `vuejs`, `svelte`, `flutter`
- **Managers:** `pip`, `uv`, `poetry`, `npm`, `maven`, `gradle`
- **Tools:** `docker`, `make`

---

## 6. クイックスタート (Quick Start)

```bash
# 1. ビルド
go build -o mngproj cmd/mngproj/main.go

# 2. 初期化
./mngproj init

# 3. コンポーネント定義 (mngproj.tomlを編集)
# [project]
# name = "my-mono-repo"
# root = "." # このディレクトリをルートとする

# [[components]]
# name = "backend"
# types = ["python", "uv"]
# path = "services/backend"
# dependencies = ["flask==2.3.0"] # 依存関係を直接記述することも可能

# [[components]]
# name = "frontend"
# types = ["node", "react"]
# path = "services/frontend"

# 4. パッケージ追加 (mngproj.tomlに記録し、自動同期)
./mngproj add backend requests

# 5. 全コンポーネネントの依存関係を同期
./mngproj sync

# 6. バックエンドとフロントエンドを並列起動 (mngproj.tomlのrunスクリプトが使われる)
./mngproj up backend frontend

# 7. フロントエンドのファイル変更を監視して自動リロード
./mngproj watch frontend
```