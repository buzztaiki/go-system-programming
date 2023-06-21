package main

import (
	"gopkg.in/fsnotify/fsnotify.v1"
	"log"
)

func main() {
	counter := 0
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			// チャネルからブロッキングしてイベントを待つ
			case event := <-watcher.Events:
				log.Println("event:", event)
				// 変更ステータスはビットフラグになってる
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("created file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("removed file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("renamed file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					log.Println("chmod file:", event.Name)
					counter++
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
			// 4回変更を検知したら終わり
			if counter > 3 {
				done <- true
			}
		}
	}()

	// 監視対象を複数追加できる
	err = watcher.Add(".")

	if err != nil {
		log.Fatal(err)
	}
	<-done
}
