# mngproj - Monorepo Polyglot Manager

**Status:** 🚧 In Development (Prototype Phase)

## 1. プロジェクト概要 (Project Overview)

`mngproj` は、多言語（Go, Python, Rust, Node.js 等）が混在するモノレポ開発において、ビルド・実行環境を抽象化・統一するためのCLIツールです。
各プロジェクトディレクトリに配置された設定ファイル (`mngproj.toml`) と、言語/ツールごとに定義された **プリセット設定 (Presets)** を組み合わせることで、複雑なビルド手順や環境依存の問題を解決します。

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

### 2.3 Isolation (Sandboxing)
パッケージマネージャによるインストールがグローバル環境を汚染しないよう、`mngproj` は自動的にローカルディレクトリ（例: `.libs`, `.npm-global`）へのインストールを強制します。

---

## 3. 設定ファイル構成 (Configuration)

### 3.1 プロジェクト設定 (`mngproj.toml`)

```toml
[project]
name = "my-web-service"
description = "A full-stack web application"

# (Optional) ロールごとの優先順位をカスタマイズ
[resolution.role_priority]
tool = 100 # toolをframeworkより優先させる例

# --- Component Definition ---
[[components]]
name = "api"
# 複数のプリセットを組み合わせる (Mixin)
types = ["python", "uv", "django"]
path = "./backend"

[components.env]
PORT = "8000"

# ユーザー定義スクリプト (最高優先度)
[components.scripts]
db_migrate = "python manage.py migrate"
```

---

## 4. コマンド体系 (Commands)

| コマンド | 引数例 | 説明 |
| :--- | :--- | :--- |
| **`init`** | `(なし)` | カレントディレクトリに `mngproj.toml` の雛形を生成します。 |
| **`run`** | `[comp] [args...]` | コンポーネントを実行します。(例: `mngproj run api`) |
| **`build`** | `[comp] [args...]` | コンポーネントをビルドします。 |
| **`install`** | `[comp] [pkgs...]` | パッケージを **ローカル環境に** インストールします。(例: `mngproj install api requests`) |
| **`remove`** | `[comp] [pkgs...]` | パッケージを削除します。 |
| **`ls`** | `(なし)` | 現在のプロジェクト内のコンポーネント一覧を表示します。 |
| **`lsproj`** | `(なし)` | カレントディレクトリ以下の **全てのプロジェクト** (`mngproj.toml`) を再帰的に検索・表示します。 |
| **`info`** | `(なし)` | 現在のプロジェクト情報や読み込まれているプリセットパスを表示します。 |
| **`query`** | `(なし)` | コンポーネント情報をJSON形式で出力します (CI/CD連携用)。 |

---

## 5. ディレクトリ構造

```text
~/Work/monorepo/
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

---

## 6. クイックスタート (Quick Start)

```bash
# 1. ビルド
go build -o mngproj cmd/mngproj/main.go

# 2. 初期化
./mngproj init

# 3. コンポーネント定義 (mngproj.tomlを編集)
# [[components]]
# name = "app"
# types = ["python", "uv"]
# path = "."

# 4. パッケージ追加 (ローカルインストール)
./mngproj install app flask

# 5. 実行
./mngproj run app
```