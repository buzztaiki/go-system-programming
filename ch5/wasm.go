package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("button clicked")
		cb.Release()
		return nil
	})
	js.Global().Get("document").
		Call("getElementById", "myButton").
		Call("addEventListener", "click", cb)
}
