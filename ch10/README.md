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
  - OSごとにAPIはあるけど実装が変わる (linux なら [inotify](https://linuxjm.osdn.jp/html/LDP_man-pages/man7/inotify.7.html))
  - 低負荷
- active: タイマーなどで定期的にフォルダを走査し、os.Stat() などを使って変更を探しに行く
  - 自分で実装する場合、こっちの方が簡単
  - 監視対象が増えるとCPU負荷やIO負荷が上がる

パッシブな方法として、ここでは `gopkg.in/fsnotify.v1` を利用する:
- API が使いやすい
- 複数ディレクトリの監視がしやすい
- マルチプラットフォームが実現できる
  - Linux: inotify
  - BSD / macOS: kqueue
  - Windows: ReadDirectoryChangesW

他のパッケージとしては以下がある:
- https://github.com/rjeczalik/notify
- https://github.com/golang/exp/tree/master/inotify
  - 今はもう存在してない

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
          - こんな感じらしい
            - malloc は大きなヒープ (arena) を確保してそこから要求されたサイズのヒープを返す
            - [brk](https://linuxjm.osdn.jp/html/LDP_man-pages/man2/brk.2.html) システムコールでヒープを確保できるんだけど
              - プロセス末尾の連続した領域しか伸ばせないから、限度がある
              - 昔は遅かったらしい
            - ので、一定以上のサイズや確保できない等の場合は mmap を使うらしい
              - mmap の場合どこかからメモリ持ってきてくれるそう
          - malloc が何故 mmap だけではなく brk を使っているのかの質問と回答
            - https://stackoverflow.com/questions/55768549/in-malloc-why-use-brk-at-all-why-not-just-use-mmap
          - 参考
            - https://www.slideshare.net/kosaki55tea/glibc-malloc#47
            - https://qiita.com/kaityo256/items/9e78b507940b2292bf79
            - http://funini.com/kei/mmap/mmap_api.shtml
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
  - [mmapのほうがreadより速いという迷信について - kazuhoのメモ置き場](https://kazuhooku.hatenadiary.org/entry/20131010/1381403041)
    > 正しいコードを書けば、シーケンシャルアクセスを行うケースにおいて read(2) が mmap(2) より大幅に遅いということは、まず起こらない。むしろ、read(2) のほうが mmap(2) よりも速くなるというケースも実際には多い。
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
- IO は CPU内部の処理に比べると劇的に遅いタスク
- 入出力が関係するプログラムではこの「重い処理」でプログラム全体が遅くならないようにする仕組みが必要
- その仕組みをシステムコールで整備するためのモデルが以下の分類
  - 同期処理と非同期処理
    - 実データを取りにいくのか、通知をもらうのかで区別
      - OS に I/O タスクを投げて結果を貰うときに
        - 同期処理: アプリケーションに処理が返ってくる
        - 非同期処理: アプリケーションが通知をもらう
  - ブロッキング処理とノンブロッキング処理
    - タスクの結果の受け取り方によって区別
      - お願いしたI/O タスクの結果が返ってくるのを
       - ブロッキング処理: できるまで待つ（自分は停止）
       - ノンブロッキング処理: できるのを待たない（自分は停止しない）
  - ここで結果は、全部完了せずバッファが一杯になるのも含む

- システムコールは以下の4種類に分けられる
  - |        | ブロッキング | ノンブロッキング         |
    |--------|--------------|--------------------------|
    | 同期   | read, write  | read, write (O_NONBLOCK) |
    | 非同期 | select, poll | AIO                      |
  - この分類は、「シングルスレッドで動作するプログラム」で「効率よくI/O をさばくため」の指標
  - がマルチプロセスやマルチスレッドで動作するときはまたちょっと別

### 10.4.1 同期・ブロッキング処理
- 読み込み・書き込み処理が完了するまでの間、何もせずに待つ
- 重い処理があれば、そこですべての処理は止まる
- 性能は良くないけどコードはシンプル


### 10.4.2 同期・ノンブロッキング処理
- API を呼ぶとAPI を呼ぶと、即座に完了フラグと、現在準備ができているデータが得られる
- 処理を完遂するには完了するまでポーリングする
- ファイルオープン時にノンブロッキングのフラグを付与して実現する
- メモ: read, write の man page によると、デバイスが処理中の時は `EAGAIN` が返るとある。
  - https://linuxjm.osdn.jp/html/LDP_man-pages/man2/read.2.html
    - ファイルディスクリプター fd がソケット以外のファイルを参照していて、 非停止 (nonblocking) モード (O_NONBLOCK) に設定されており、読み込みを行うと停止する状況にある
  - https://linuxjm.osdn.jp/html/LDP_man-pages/man2/write.2.html
    - ファイルディスクリプター fd がソケット以外のファイルを参照していて、 非停止 (nonblocking) モード (O_NONBLOCK) に設定されており、書き込みを行うと停止する状況にある
  - つまり、読んだり書いたりできるまで (`EAGAIN` が返らなくなるまで) read, write を何度も呼ぶという事をポーリングと呼んでいる。


### 10.4.3 非同期・ブロッキング処理

- 準備が完了したものがあれば通知してもらう、というイベント駆動モデル
- I/O 多重化（I/O マルチプレクサー）とも呼ばれる
- 通知の受けとりには select 属と呼ばれるシステムコールを使う
  - POSIX: select, poll,
  - Linux: epoll
  - BSD: kqueue
  - Solaris: Event ports
  - Windows: I/O Completion Port

### 10.4.4 非同期・ノンブロッキング処理
- メインプロセスのスレッドとは完全に別のスレッドでタスクを行い、完了したらその通知だけを受け取るモデル
- POSIX の非同期I/O (`aio_*`) が有名
  - Linux では成熟してないし、Linus もあまり興味もってないらしい
  - Node.js でも何度か検討されたけど却下されてる
  - Nginx も当初は使ってたけど、マルチスレッドに変更してスループットが9倍になった
- 非同期IO API の進化
  - Linux では、今までの非同期IO の欠点を解消した `io_uring` が2019年の Linux 5.1 に搭載
    - 5.8.2 項でも紹介した
    - アプリケーションとカーネルが送受信用の2つのキューを使ってやり取りする
    - システムコール呼び出しの待ち時間のロスがない
    - アプリケーションとカーネルの間のメモリのコピーも少なくて済む

### 10.4.5 Go言語でさまざまなI/O モデルを扱う手法

- Go ではベースとなるファイルI/OやネットワークI/Oは、シンプルな同期・ブロッキングインタフェースになっている
- Go言語で同期・ブロッキング以外のI/O モデルを実現するには、ベースの言語機能（goroutine、チャネル、select）をアプリケーションで組み合わせる
  - goroutine をたくさん実行し、それぞれに同期・ブロッキングI/O を担当させると、**非同期・ノンブロッキング**
  - goroutine で並行化させたI/Oの入出力でチャネルを使うことで、他のgoroutine とやり取りする箇所のみの同期ができる
  - チャネルにバッファがあれば、チャネルの書き込み側も **ノンブロッキング**
  - これらのチャネルに select 構文を使うと **非同期・ブロッキングのI/O 多重化**
  - select 構文にdefault 節があると、読み込みを **ノンブロッキング** で行える
- このように、基本シンタックスの組み合わせで、同期で実装したコードをラップし、さまざまなI/O のスタイルを実現できる点がGo言語の強み
  - 他の言語だと、非同期をあとから足そうとするとプログラムの多くの箇所で修正が必要になることがある
  - 特に、次節で扱う非同期・ブロッキングの追加では、プログラムが構造から変わってくることすらあります
    - メモ async / await の事を言ってるかな？


## 10.5 select属のシステムコールによるI/O多重化
- 非同期・ブロッキングは1スレッドでたくさんの入出力を効率よく扱うための手法で、I/O 多重化とも呼ばれる
- これを効率よく実現する API を本書では **select族** と総称する。
- select 属はC10K 問題と呼ばれる、万の単位の入出力を効率よく扱うための手法として有効
- ネットワークについては、Go言語のランタイム内部に組み込まれていて、サーバーを実装したときに効率よくスケジューラーが働くようになっている
- POSIX の select と poll
  - select は、扱うディスクリプタの数が増えたときの性能に問題
  - pollは、多少はましだけど移植性に欠ける
    - MEMO: POSIX なのに？古いカーネルだと使えないとか？
- select 属のAPI を使う目的は、大量のI/O をさばくときのパフォーマンス向上
  - 中途半端なAPI では使う意味がない
  - よって、パフォーマンスも移植性も落ちる poll に相当するものは、Go言語のsyscallパッケージに含まれてない
- go の runtime で使われているのは [10.4.3](#1043-非同期ブロッキング処理) の一覧で書いたものと同じ
- サンプルコード
  - [ch10_4_kqueue.go](ch10_4_kqueue.go)
    - `./test` フォルダの変更を監視してイベントを待つ
    - メモ
      - やってる事は fsnotify の時と大体一緒っぽい
      - fsnotify の実装が kqueue ってあったしね

  - [ch10_4_iouring.go](ch10_4_iouring.go)
    - https://github.com/Iceber/iouring-go が見つかったから、それのサンプルに少しコメント付けたやつ
- Linux の epoll はネットワークにしか使えない
- 環境も、サポートされるイベントやオプションの種類も膨大だから、詳細は省いている
- 具体的なコードが見たい場合は [evio](https://github.com/tidwall/evio) の実装を見るとよい
  - Goの標準ライブラリよりも高速な、ネットワーク特化のイベントループのライブラリ
  - internal ディレクトリを見ると kqueue, epoll を使ったコードが読める
- OS の select はブロックしうる複数のI/O システムコールをまとめて登録して準備ができたものを教えてもらうのが役割
  - Go の select と考え方は同じ
  - 内部実装も Linux カーネルと似てるらしい
    - [Goをカンストさせる話](https://www.slideshare.net/moriyoshi/go-73631497)
    - メモ: Linux カーネル関係ないけど↑のスライドによると
      - [select に 65535 以上の case を書くと](https://www.slideshare.net/moriyoshi/go-73631497#11)
      - [too many cases で死ぬ](https://www.slideshare.net/moriyoshi/go-73631497#22)
        - ちゃんと panic してて偉い

## 10.6 FUSEを使った自作のファイルシステムの作成
- ファイルシステムの集大成としてファイルシステムを Go で自作する
- AWS S3, GCP Storage, Azure Blob Storage をマウントして使える読み込み専用の FS を作る
- FUSE を使うとユーザーランドで動作する FS が簡単に作れる
  - FUSE は FS を操作するシステムコールを、ユーザーランド上で動作しているプロセスに転送する
  - 実際のファイル操作は、カーネル内部ではなくユーザーランドで行われているが、ファイルを操作するアプリケーションからは、まるで本物のファイルシステムがあるかのように見える
  - https://ja.wikipedia.org/wiki/Filesystem_in_Userspace
  - https://www.kernel.org/doc/html/latest/filesystems/fuse.html
- 本節ではWindows を含む多くのプラットフォームに対応していてAPI も簡便な [github.com/billziss-gh/cgofuse](https://pkg.go.dev/github.com/billziss-gh/cgofuse) を利用する
  - これを利用するには以下のライブラリが必要
    - Linux: libfuse
    - macOS: [FUSE for macOS](https://osxfuse.github.io/)
    - Windows: [Windows File System Proxy](https://github.com/winfsp/winfsp)
      - https://github.com/dokan-dev/dokany みたいなやつ
  - `Open()` や `Read()` など33通りのコードを `FileSystemInterface` インターフェースに実装する
    - `fuse.FileSystemBase` 構造体を自分の構造体に埋め込む事で簡単に実装できる
      - メモ: Go でよくある実装パターン
  - メモ: https://github.com/hanwen/go-fuse の方がstar多いんだけど、cgofuse を選んだのは Windows 対応させたかったのかな。
    - ファイルで抽象化されてるから、こっちの方が使いやすそうに見える
- メモ: Go の中で完結するなら https://pkg.go.dev/io/fs なんてものもある。
  - これを fuse にマップすると色々やれそうな気がする。というかありそう。


### 10.6.1 クラウドのストレージサービスへのアクセス
- クラウドサービスのアクセスには [Go Cloud Development Kit](https://gocloud.dev/) を使う
- `s3://<bucket>`, `gs://<bucket>`, `azblob://<bucket>` のように使える
  - https://gocloud.dev/howto/blob/
  - Azure の場合は `AZURE_STORAGE_ACCOUNT`, `AZURE_STORAGE_KEY`, `AZURE_STORAGE_SAS_TOKEN` の指定がいる。めんどい。
    - ひとまず direnv で指定してる
- サンプルコード:
  - [ch10_6_cloud_storage.go](ch10_6_cloud_storage.go)
    - 開いて stdout に流すだけ


### 10.6.2 ファイルシステム作成の最初の一歩
- 難しい事もなく、cgofuse を使ってマウント出来るようにしてるだけ
- サンプルコード:
  - [ch10_6_cloudfs/main.go](ch10_6_cloudfs/main.go)
- `-d` オプションを付けて実行するとデバッグ出力が得られる


### 10.6.3 ディレクトリ内の一覧を取得できるようにする
- 10.6.2 のコードだとマウントはできるけど他に何もできない
- `-d` を付けて実行すると `getattr` がないと言われているからこれを実装する
  - [ch10_6_cloudfs/readattr.go](ch10_6_cloudfs/readattr.go)
- 次は `readdir` が無いと言われるからこれも実装する
  - [ch10_6_cloudfs/readdir.go](ch10_6_cloudfs/readdir.go)
- これで ls ができるようになる

### 10.6.4 ファイルの読み込み
- `Read` を実装して読み込みができるようにする
  - [ch10_6_cloudfs/read.go](ch10_6_cloudfs/read.go)
  

### 10.6.5 実用的なファイルシステムを実装するには
- シンプルさを優先してナイーブな実装例になってるから遅い
  - `Read()` に渡されるバッファは 130KB 程度だから沢山分割される
  - `Open()` を実装してファイルディスクリプタを発行できれば連続した read の場合は同じ `io.ReadCloser` が使えるようになる
    - メモ: `Write()` を実装するのにも必要
      - blob に対して offset を指定して書き込めないから
  - また、クラウドサービスがファイルのハッシュを持ってるから、キャッシュする事もできる

### 10.7 本章のまとめと次章予告
PDF を読みましょう。

