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
- unix のサンプルは FileLock 構造体を使うようになってなかったから使うように書き換えた
  - EINTR は https://blog.lufia.org/entry/2020/02/29/162727 を参照
- windows のサンプルは、そのままでは動かなかったので https://pkg.go.dev/golang.org/x/sys/windows を使うように書き換えた。




### 10.7 本章のまとめと次章予告


TODO

Go の中で完結するなら https://pkg.go.dev/io/fs なんてものもある。
