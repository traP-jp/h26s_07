# Getting Started

## 開発を始める前に

1. [mise](https://mise.jdx.dev/) をインストール
   [インストール方法](https://mise.jdx.dev/installing-mise.html)
   面倒な人のための一例(bash)
   ```sh
   $ sudo apt install -y extrepo
   $ sudo extrepo enable mise
   $ sudo apt update
   $ sudo apt install -y mise
   $ echo 'eval "$(mise activate bash)"' >> ~/.bashrc
   ```
2. リポジトリをクローン
    ```sh
    $ git clone git@github.com:traP-jp/h26_07.git
    ```
3. 環境をセットアップ
    ```sh
    $ mise trust
    $ mise install
    $ corepack enable
    $ mise run init
    ```
4. バージョンを確認
   ```sh
   $ node -v
   v24.17.0

   $ go version
   go version go1.26.4 ...

   $ cd frontend
   $ pnpm -v
   11.8.0
   ```

## 開発サーバー

```sh
$ mise run backend
$ mise run frontend
```

## バックエンド開発

```sh
$ cd backend
$ mise run dev # バックエンドサーバーを起動

$ mise run fmt # コードを整形

$ mise run test # テストを実行
```

## Codegen

```sh
$ mise run codegen # backend/frontend の生成コードを更新
```

backendだけ

```sh
$ cd backend && mise run gen:api
```
