package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/pkg/xattr"
)

type Config struct {
	Folder string `json:"folder"`
}

var watchDir string

const empty = "[ ]"

var lastFile string

func main() {

	cfg, err := loadConfig()
	if err != nil || len(cfg.Folder) == 0 {
		log.Fatalln("Ошибка загрузки config.json:", err)
		return
	}
	watchDir = cfg.Folder
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Ошибка получения пути к бинарнику:", err)
		return
	}
	lastMod := getModTime(exePath)
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
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					if err := addDirRecursive(watcher, event.Name); err != nil {
						fmt.Println("Ошибка рекурсивного добавления директории:", err)
					}
				} else {
					processFile(event.Name)
				}
			}
		case err := <-watcher.Errors:
			fmt.Println("Ошибка watcher:", err)
		}
	}

}

func addDirRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if err := w.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
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
	summary = truncate(summary, 20)
	systray.SetTitle("[ " + summary + " ]")
	copyToClipboard(summary)
	go func() {
		time.Sleep(10 * time.Second)
		systray.SetTitle(empty)
	}()

}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func getXAttr(file, attr string) string {
	b, err := xattr.Get(file, attr)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func copyToClipboard(text string) {
	if err := clipboard.WriteAll(text); err != nil {
		log.Printf("clipboard error: %v\n", err)
	}
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

func loadConfig() (Config, error) {

	exePath, err := os.Executable()
	if err != nil {
		return Config{}, err
	}
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)

	return cfg, err
}
