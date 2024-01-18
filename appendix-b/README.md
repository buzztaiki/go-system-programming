# 付録B デバッガーのお仕事

## B.1 デバッグ対象のプログラムに接続する

SysProcAttr構造体においてPtraceフラグをオン
https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/proc_linux.go#L100-L109

```
func Launch(cmd []string, wd string, flags proc.LaunchFlags, debugInfoDirs []string, tty string, stdinPath string, stdoutOR proc.OutputRedirect, stderrOR proc.OutputRedirect) (*proc.TargetGroup, error) {
...
		process = exec.Command(cmd[0])
		process.Args = cmd
		process.Stdin = stdin
		process.Stdout = stdout
		process.Stderr = stderr
		process.SysProcAttr = &syscall.SysProcAttr{
			Ptrace:     true,
			Setpgid:    true,
			Foreground: foreground,
		}
```


アタッチ

https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/ptrace_linux.go#L10C1-L12
```
func ptraceAttach(pid int) error {
	return sys.PtraceAttach(pid)
}
```


## B.2 プログラムを止めたり進めたりする

ステップオーバー
https://speakerdeck.com/aarzilli/internal-architecture-of-delve?slide=53
> - Set a breakpoint on every line of the current function
>   – condition: stay on the same goroutine & stack frame
> - Set a breakpoint on the return address of the current frame
> – condition: stay on the same goroutine & previous stack frame
>   - Set a breakpoint on the most recently deferred function
> – condition: stay on the same goroutine & check that it was called through a panic
>   -  Call Continue

ハードウェアブレークポイント
https://speakerdeck.com/aarzilli/internal-architecture-of-delve?slide=40

> The target layer overwrites the instruction at 0x452e60 with an instruction that, when executed, stops execution of the thread and makes the OS notify the debugger.
> – In intel amd64 it’s the instruction INT 3 which is encoded as 0xCC


### B.2.1 ハードウェアブレークポイント

sys.Pwait4
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/proc_linux.go#L615
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/threads_linux.go#L61

WaitForDebugEvent
- https://github.com/go-delve/delve/blob/v1.8.0/pkg/proc/native/proc_windows.go#L248


PTRACE_POKEUSR
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/threads_linux_amd64.go#L76
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/hwbreak_amd64.go#L8
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/hwbreak_amd64.go#L14

writeSoftwareBreakpoint
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/proc.go#L422
- https://github.com/go-delve/delve/blob/1b201c3184ca33fc64b302c74cc82b6d7318b000/pkg/proc/native/threads.go#L66

