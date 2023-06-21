# 第10章 ファイルシステムの最深部を扱うGo言語の関数

## 10.1 ファイルの変更監視（syscall.Inotify*）

ファイルの変更監視には以下の2種類が考えられる

- passive: 監視したいファイルをOS側に通知しておいて、変更があったら教えてもらう
  - OSごとにAPIはあるけど実装が変わる (linux なら [ionotify](https://linuxjm.osdn.jp/html/LDP_man-pages/man7/inotify.7.html))
  - 低負荷
- active: タイマーなどで定期的にフォルダを走査し、os.Stat() などを使って変更を探しに行く
  - 簡単
  - 監視対象が増えるとCPU負荷やIO負荷が上がる

ここでは `gopkg.in/fsnotify.v1` を利用する:
- API が使いやすい
- 複数ディレクトリの監視がしやすい
- マルチプラットフォームが実現できる
  - Linux: ionotify
  - BSD / macOS: kqueue
  - Windows: ReadDirectoryChangesW

他のパッケージとしては以下がある:
- https://github.com/rjeczalik/notify
- https://github.com/golang/exp/

## 10.2 ファイルのロック（syscall.Flock()）

- ファイルのロックは、複数のプロセス間で同じリソースを同時に変更しないようにするために他のプロセスに伝える手法のひとつ。
- もっとも単純な方法は、ロックファイルを作る事。大昔のCGIとか。
- ロックファイルはポータブルだけど確実ではない。お互いそのロックファイルを知らないといけないとか。
- 確実な方法として、ファイルをロックするシステムコールを利用する方法がある。
- Go では POSIX 系なら `syscall.Flock` が使える。
  - 通常の入出力のシステムコールではロックの有無は確認されない
  - こういうロックをアドバイザリーロック (勧告ロック) と呼ぶ
- [flock](https://linuxjm.osdn.jp/html/LDP_man-pages/man2/flock.2.html) は実は POSIX API ではない
  - が、多くの POSIX 系 OS で利用可能
  - POSIX で規定されてるのは [fcntl](https://linuxjm.osdn.jp/html/LDP_man-pages/man2/fcntl.2.html)
    - 強制ロックも含めファイルに対する操作を色々できる
  - X/Open System Interface では [lockf](https://linuxjm.osdn.jp/html/LDP_man-pages/man3/lockf.3.html) が規定されてる
- 強制ロックについて
  - https://linuxjm.osdn.jp/html/LDP_man-pages/man2/fcntl.2.html
    - デフォルトでは昔からある (プロセスに関連付けられる) レコードロックもオープンファイル記述のレコードロックもアドバイザリーロック。 アドバイザリーロックに強制力はなく、協調して動作するプロセス間でのみ有効。
    - 両方のタイプのロックを強制ロックにすることができる。 強制ロックは全てのプロセスに対して効果がある。 
    - あるプロセスが強制ロックが適用されたファイル領域に対してアクセスを実行しようとした場合、アクセスの結果はそのファイルのオープンファイル記述で O_NONBLOCK フラグが有効になっているかにより決まる。
      - O_NONBLOCK フラグが有効になっていないときは、ロックが削除されるかロックがアクセスと互換性のあるモードに変換されるまで、システムコールは停止 (block) される。 
      - O_NONBLOCK フラグが有効になっているときは、システムコールはエラー EAGAIN で失敗する。
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
- syscall.Flock() によるロックでは、ロックが外されるまで待たされる
  - tryLock みたいな事を実現するためにノンブロッキングモードを利用する
  - go の場合は goroutine と channel で回避できるからブロッキングモードだけが用意されることもよくある
    - メモ: でもロックが取れなかったらあきらめるとか難しそうな気もするけど

### 10.2.2 LockFileEx() によるWindows でのファイルロック
- LockFileEx()でも、排他ロックと共有ロックをフラグで使い分ける
- ロックの解除はUnlockFileEx()という別のAPI になっている
- ロックするファイルの範囲を指定することもできる
- メモ: サンプルそのままでは動かなかったので https://pkg.go.dev/golang.org/x/sys/windows を使うように書き換えた。

### 10.2.3 FileLock構造体の使い方
- メモ:
  - unix のサンプルは FileLock 構造体を使うようになってなかったから使うように書き換えた
    - EINTR は https://blog.lufia.org/entry/2020/02/29/162727 を参照
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
- Windows でも `CreateFileMapping()`,  `MapViewOfFile()` API の組で同じことが実現できる
- クロスプラットフォームなライブラリも何種類かある
  - https://pkg.go.dev/golang.org/x/exp/mmap
    - シンプル
    - `io.ReaderAt` を満たす
  - https://pkg.go.dev/github.com/edsrzf/mmap-go
    - より柔軟で POSIX に近いインターフェイス
    - マッピングしたメモリを表すスライスをそのまま返すから、Goの文法を使って自由にデータにアクセスできる
    - この節ではこれを使う

- http://funini.com/kei/mmap/mmap_implementation.shtml が分かりやすかった

TODO

- メモ: Apache Lucene に MMapDirectory なんてものがありましたね
  - https://lucene.apache.org/core/9_6_0/core/org/apache/lucene/store/MMapDirectory.html
  - Java の場合ヒープ外のメモリを利用できる (GC の対象にならない) っていう利点もある


### 10.7 本章のまとめと次章予告


TODO

Go の中で完結するなら https://pkg.go.dev/io/fs なんてものもある。
