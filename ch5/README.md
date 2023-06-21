## syscall/js

以下にしないと `imports syscall/js: build constraints exclude all Go files in /usr/lib/go/src/syscall/js` になる

```
GOOS=js GOARCH=wasm go build wasm.go
```

ちなみに go run するとバイナリフォーマットが違うから以下のエラー

```
GOOS=js GOARCH=wasm go run wasm.go
fork/exec /tmp/go-build1888649922/b001/exe/wasm: exec format error
```

wasmer で実行しようとしたけど

```
❯❯ wasmer run wasm
error: failed to run `wasm`
╰─▶ 1: Error while importing "go"."debug": unknown import. Expected Function(FunctionType { params: [I32], results: [] })
```
