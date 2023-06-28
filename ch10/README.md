# 第10章 ファイルシステムの最深部を扱うGo言語の関数

この章ではアプリケーションから見たファイルシステムまわりの最深部をたどる。扱う話題は
- ファイルの変更監視
- ファイルロック
- ファイルのメモリへのマッピング
- 同期・非同期とブロッキング・ノンブロッキング
- select 属のシステムコールによるI/O 多重化
- FUSE を使った自作のファイルシステムの作成


## 10.1 ファイルの変更監視（syscall.Inotify*）

ファイルの変更監視には以下の2種類が考えられる

- passive: 監視したいファイルをOS側に通知しておいて、変更があったら教えてもらう
  - OSごとにAPIはあるけど実装が変わる (linux なら [ionotify](https://linuxjm.osdn.jp/html/LDP_man-pages/man7/inotify.7.html))
  - 低負荷
- active: タイマーなどで定期的にフォルダを走査し、os.Stat() などを使って変更を探しに行く
  - 自分で実装する場合、こっちの方が簡単
  - 監視対象が増えるとCPU負荷やIO負荷が上がる

パッシブな方法として、ここでは `gopkg.in/fsnotify.v1` を利用する:
- API が使いやすい
- 複数ディレクトリの監視がしやすい
- マルチプラットフォームが実現できる
  - Linux: ionotify
  - BSD / macOS: kqueue
  - Windows: ReadDirectoryChangesW

他のパッケージとしては以下がある:
- https://github.com/rjeczalik/notify
- https://github.com/golang/exp/

サンプル
- [ch10_1_fsnotify.go](./ch10_1_fsnotify.go)
  - プログラムを実行して、そのディレクトリで touch とか mv とかするとイベントが起きる
  - 4回検知したら終わり

## 10.2 ファイルのロック（syscall.Flock()）

- ファイルのロックは、複数のプロセス間で同じリソースを同時に変更しないようにするために他のプロセスに伝える手法のひとつ。
- もっとも単純な方法は、ロックファイルを作る事。
  - 大昔のCGIとか。
  - ロックファイルはポータブルだけど確実ではない。
    - お互いそのロックファイルを知らないといけないとか。
    - 確実に消せないといけないとか。
- 確実な方法として、ファイルをロックするシステムコールを利用する方法がある。
- Go では POSIX 系なら `syscall.Flock` が使える。
  - 通常の入出力のシステムコールではロックの有無は確認されない
    - `Flock` で確認しないといけない。
  - こういうロックをアドバイザリーロック (勧告ロック) と呼ぶ
- [flock](https://linuxjm.osdn.jp/html/LDP_man-pages/man2/flock.2.html) は実は POSIX API ではない
  - が、多くの POSIX 系 OS で利用可能
  - POSIX で規定されてるのは [fcntl](https://linuxjm.osdn.jp/html/LDP_man-pages/man2/fcntl.2.html)
    - 強制ロックも含めファイルに対する操作を色々できる
  - X/Open System Interface では [lockf](https://linuxjm.osdn.jp/html/LDP_man-pages/man3/lockf.3.html) が規定されてる
- メモ: fcntl の 強制ロックについて
  - https://linuxjm.osdn.jp/html/LDP_man-pages/man2/fcntl.2.html
    - あるプロセスが強制ロックが適用されたファイル領域に対してアクセスを実行しようとした場合、アクセスの結果はファイルを開くときの O_NONBLOCK フラグが有効になっているかで決まる。
      - O_NONBLOCK フラグが有効に **なっていない** ときは、ロックが削除されるかロックがアクセスと互換性のあるモードに変換されるまで、システムコールは停止 (block) される。 
      - O_NONBLOCK フラグが有効に **なっている** ときは、システムコールはエラー EAGAIN で失敗する。
- Windows では `LockFileEx` 関数を使う。これは強制ロック。
  - `syscall.Flock()` は利用できない。

### 10.2.1 syscall.Flock() によるPOSIX系OSでのファイルロック

- 共有ロックと排他ロック
  - 共有ロックは、複数のプロセスから同じリソースに対していくつも同時にかけられる
  - 排他ロックでは他のプロセスからの共有ロックがブロックされる
  - メモ: ReadLock と WriteLock みたいな言いかたもすると思う。
    - https://docs.oracle.com/en/java/javase/20/docs/api/java.base/java/util/concurrent/locks/ReadWriteLock.html
    - https://doc.rust-lang.org/std/sync/struct.RwLock.html
- モードフラグ
  - | フラグ  | 説明                                                                               |
    |---------|------------------------------------------------------------------------------------|
    | LOCK_SH | 共有ロック。他のプロセスからも共有ロックなら可能だが、排他ロックは同時には行えない |
    | LOCK_EX | 排他ロック。他のプロセスからは共有ロックも排他ロックも行えない                     |
    | LOCK_UN | ロック解除。ファイルをクローズしても解除になる                                     |
    | LOCK_NB | ノンブロッキングモード                                                             |
- syscall.Flock() によるロックでは、排他ロックが外されるまで待たされる
  - tryLock みたいな事を実現するためにノンブロッキングモードを利用する
  - go の場合は goroutine と channel で回避できるからブロッキングモードだけが用意されることもよくある


### 10.2.2 LockFileEx() によるWindows でのファイルロック
- LockFileEx()でも、排他ロックと共有ロックをフラグで使い分ける
- ロックの解除はUnlockFileEx()という別のAPI になっている
- ロックするファイルの範囲を指定することもできる
- メモ: サンプルそのままでは動かなかったので https://pkg.go.dev/golang.org/x/sys/windows を使うように書き換えた。

### 10.2.3 FileLock構造体の使い方

- サンプル
  - [ch10_2_filelock/](./ch10_2_filelock/)
    - 指定したファイルのロックを取得して10 秒後に解放する。
    - 10 秒以内に他のコンソールから同じプログラムを実行すると、最初のプロセスが終了するまで、他のプロセスが待たされる。
    - Build Constraints を使って、同じソースツリーで unix 系と Windows の両方でビルドできるようになってる。
- メモ:
  - unix のサンプルは FileLock 構造体を使うようになってなかったから使うように書き換えた
  - windows のサンプルは、そのままでは動かなかったので https://pkg.go.dev/golang.org/x/sys/windows を使うように書き換えた。
- マルチプラットフォームを実現するための手段
  - Build Constraints
    - コード先頭に `//go:build` に続けてビルド対象のプラットフォームを列挙したり、ファイル名に `_windows.go` のようなサフィックスを付ける
      - Go 1.17 以前は `// +build` で指定
    - POSIX 以外ではファイル名にサフィックスを付けるのが一般的
    - POSIX では `//go:build unix` を指定する
      - `_unix.go` というファイル名はすでに広く使われていて、後方互換性のためにコメントだけでしか使えない
  - `runtime.GOOS` 定数を使って実行時に処理を分岐
    - API 自体がプラットフォームによって異なる場合にはリンクエラーが発生する
  - メモ: この手のコメント (directive) はまとまったドキュメントがない気がする
    - build: 
      - https://pkg.go.dev/cmd/go#hdr-Build_constraints
      - https://go.googlesource.com/proposal/+/master/design/draft-gobuild.md
      - `go help buildconstraint`
    - compiler directive: 
      - https://pkg.go.dev/cmd/compile#hdr-Compiler_Directives
    - cgo:
      - https://pkg.go.dev/cmd/cgo@go1.20.5
    - generate: 
      - https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
      - `go help generate`
    - embed:
      - https://pkg.go.dev/embed

## 10.3 ファイルのメモリへのマッピング（syscall.Mmap()）
- これまで利用していた `os.File` は `io.Seeker` を満たしてるからランダムアクセスをする事はできる。が、読み込み位置を移動する必要がある
- これを解消するには `syscall.Mmap()` システムコールを使う
- これを使うと
  - ファイルの中身をそのままメモリ上に展開できる
  - メモリ上で書き換えた内容をそのままファイルに書き込むこともできる
- マッピングという名前のとおり、ファイルとメモリの内容を同期させる
- メモリマップドファイルとも呼ばれる
- メモ: ↓が分かりやすかった
  - http://funini.com/kei/mmap/mmap_api.shtml
  - http://funini.com/kei/mmap/mmap_implementation.shtml
- Windows でも `CreateFileMapping()`,  `MapViewOfFile()` API の組で同じことが実現できる
- クロスプラットフォームなライブラリも何種類かある
  - https://pkg.go.dev/golang.org/x/exp/mmap
    - シンプル
    - `io.ReaderAt` を満たす
  - https://pkg.go.dev/github.com/edsrzf/mmap-go
    - より柔軟で POSIX に近いインターフェイス
    - マッピングしたメモリを表すスライスをそのまま返すから、Goの文法を使って自由にデータにアクセスできる
    - この節ではこれを使う
- [ch10_3_mmap.go](./ch10_3_mmap.go)
  - mmap-go では以下のような関数が使える
    - `mmap.Map()`: 指定したファイルの内容をメモリ上に展開
    - `mmap.Unmap()`: メモリ上に展開された内容を削除して閉じる
    - `mmap.Flush()`: 書きかけの内容をファイルに保存する
    - `mmap.Lock()`: 開いているメモリ領域をロックする
    - `mmap.Unlock()`: メモリ領域をアンロックする
  - `mmap.Map()`
    - mmap-go ではオフセットとサイズを指定しなくても良い
    - オフセットとサイズを調整して一部だけを読み込みたい場合は、本来の `mmap()` システムコールに近い挙動の `mmap.MapRegion()` を使う
    - 第3引数 (`flags`)
      - 特殊なフラグで、`mmap.ANON` を渡すとファイルをマップせずメモリ領域だけ確保する
        - 「第16章 Go言語のメモリ管理」で再度登場する
          - Go がヒープを確保するときは mmap してるらしい
        - malloc の内側で使われてるらしい
          - 参考
            - https://www.slideshare.net/kosaki55tea/glibc-malloc#47
            - https://qiita.com/kaityo256/items/9e78b507940b2292bf79
            - http://funini.com/kei/mmap/mmap_api.shtml
          - こんな感じらしい
            - malloc は大きなヒープ (arena) を確保してそこから要求されたサイズのヒープを返す
            - [brk](https://linuxjm.osdn.jp/html/LDP_man-pages/man2/brk.2.html) システムコールでヒープを確保できるんだけど
              - プロセス末尾の連続した領域しか伸ばせないから、限度がある
              - 昔は遅かったらしい
            - ので、一定以上のサイズや確保できない等の場合は mmap を使うらしい
              - mmap の場合どこかからメモリ持ってきてくれるそう
          - malloc が何故 mmap だけではなく brk を使っているのかの質問と回答
            - https://stackoverflow.com/questions/55768549/in-malloc-why-use-brk-at-all-why-not-just-use-mmap
    - 第2引数 (`prot`)
      - `mmap.RDONLY`: 読み込み専用
      - `mmap.RDWR`: 読み書き可能
      - `mmap.EXEC`: 実行可能にする
      - `mmap.COPY`: コピーオンライト
        - 複数プロセスが同じファイルを mmap している時に、変更があった場所だけ新しいメモリを割り当てる
        - ファイルへの書き戻しはできない
- [ch10_3_mmap_multi.go](./ch10_3_mmap_multi.go)
  - 複数プロセスで mmap するとどうなるか確認する用に作ってみたもの
  - 結果として
    - `RDWR`: 
      - 他で書き込んだ状態が見える
      - 自分が書き込んだ状態が他でも見える
    - `RDONLY`: 
      - 他で書き込んだ状態は見える
      - 自分が書こうとしたら segv した
    - `COPY`: 
      - 自分が書くまでは他が書き込んだ状態が見える
      - 自分が書いたら、
        - そこから先他が書き込んだ状態が見えなくなる
        - 自分が書いたのも他から見えない
        - ページ単位で cow してるらしく、まとめてこの状態になる

### 10.3.1 mmapの実行速度
- `File.Read()` より mmap が速いかというとケースバイケース
- 逐次処理なら `File.Read()` でも十分速い
- データベースのように、全部メモリ上に読み込んでランダムアクセスが必要なら mmap が有利な事もある
  - アクセスがあるところを先に読み込むなどの最適化もカーネルが行う
  - 一度に多くのメモリを確保しなければならないため、ファイルサイズが大きくなるとI/O 待ちが長くなる可能性はある
  - メモ: Apache Lucene の MMapDirectory が多分そのケース
    - https://lucene.apache.org/core/9_6_0/core/org/apache/lucene/store/MMapDirectory.html
    - Java の場合ヒープ外のメモリを利用できる (GC の対象にならない) っていう利点もある
- コピーオンライト機能を使う場合や、確保したメモリの領域にアセンブリ命令が格納されていて実行を許可する必要がある場合には、mmap 一択
- メモリ共有の仕組みとしても使える。
  - ファイルI/O 以外の話なので、16.1.3「実行時の動的なメモリ確保：ヒープ」で改めて取り上げる。

## 10.4 同期・非同期／ブロッキング・ノンブロッキング
TODO

### 10.4.1 同期・ブロッキング処理
### 10.4.2 同期・ノンブロッキング処理
### 10.4.3 非同期・ブロッキング処理
### 10.4.4 非同期・ノンブロッキング処理
### 10.4.5 Go言語でさまざまなI/O モデルを扱う手法

## 10.5 select属のシステムコールによるI/O多重化
TODO

## 10.6 FUSEを使った自作のファイルシステムの作成
TODO

メモ: Go の中で完結するなら https://pkg.go.dev/io/fs なんてものもある。

### 10.7 本章のまとめと次章予告


TODO


