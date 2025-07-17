package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
)

const watchDir = "/Users/alien/Vault/Messages/"
const empty = "[ ]"

var lastFile string

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Ошибка получения пути к бинарнику:", err)
		return
	}
	lastMod := getModTime(exePath)

	// Запускаем проверку обновлений в фоне
	go func() {
		for {
			time.Sleep(2 * time.Second)
			if isUpdated(exePath, lastMod) {
				fmt.Println("Бинарник обновлён — перезапуск...")
				exec.Command(exePath).Start()
				os.Exit(0)
			}
		}
	}()

	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetTitle(empty)
	go watch()
}

func watch() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	err = filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Ошибка Walk:", err)
		return
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create != 0 {
				processFile(event.Name)
			}
		case err := <-watcher.Errors:
			fmt.Println("Ошибка watcher:", err)
		}
	}

}

func processFile(file string) {

	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		return
	}
	if file == lastFile {
		return
	}
	lastFile = file
	xtype := getXAttr(file, "type")
	if xtype != "code" {
		return
	}
	summary := getXAttr(file, "summary")
	if summary == "" {
		return
	}
	systray.SetTitle("[ " + summary + " ]")
	copyToClipboard(summary)
	go func() {
		time.Sleep(10 * time.Second)
		systray.SetTitle(empty)
	}()

}

func getXAttr(file, attr string) string {

	out, err := exec.Command("xattr", "-p", attr, file).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func copyToClipboard(text string) {

	cmd := exec.Command("pbcopy")
	in, _ := cmd.StdinPipe()
	_ = cmd.Start()
	_, _ = in.Write([]byte(text))
	_ = in.Close()
	_ = cmd.Wait()

}

func getModTime(path string) time.Time {

	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}

	return info.ModTime()

}

func isUpdated(path string, last time.Time) bool {

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.ModTime().After(last)

}
