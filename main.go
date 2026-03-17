package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var Version = "0.1.1"

type WorldInfo struct {
	folder        string
	name          string
	lastSaved     time.Time
	size          uint64
	sizeFormatted string
	path          string
}

var worlds []WorldInfo
var worldList *widget.List
var searchEntry *widget.Entry
var statusLabel *widget.Label

func getWorldsDir() string {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		appdata = "."
	}
	return filepath.Join(appdata, "MinecraftPC_Netease_PB", "minecraftWorlds")
}

func formatSize(bytes uint64) string {
	const kb = 1024
	const mb = kb * 1024
	const gb = mb * 1024

	if bytes < kb {
		return fmt.Sprintf("%.2f B", float64(bytes))
	} else if bytes < mb {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	} else if bytes < gb {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
}

func getFolderSize(path string) uint64 {
	var size uint64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	return size
}

func loadWorlds() error {
	worldsDir := getWorldsDir()

	if _, err := os.Stat(worldsDir); os.IsNotExist(err) {
		return fmt.Errorf("存档目录不存在: %s", worldsDir)
	}

	worlds = nil

	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		if strings.HasPrefix(folderName, "+++") {
			continue
		}

		folderPath := filepath.Join(worldsDir, folderName)
		levelnamePath := filepath.Join(folderPath, "levelname.txt")

		if _, err := os.Stat(levelnamePath); os.IsNotExist(err) {
			continue
		}

		nameBytes, err := os.ReadFile(levelnamePath)
		if err != nil {
			continue
		}
		worldName := strings.TrimSpace(string(nameBytes))
		if worldName == "" {
			worldName = "未知"
		}

		leveldatPath := filepath.Join(folderPath, "level.dat")
		var lastSaved time.Time
		if info, err := os.Stat(leveldatPath); err == nil {
			lastSaved = info.ModTime()
		}

		size := getFolderSize(folderPath)

		worlds = append(worlds, WorldInfo{
			folder:        folderName,
			name:          worldName,
			lastSaved:     lastSaved,
			size:          size,
			sizeFormatted: formatSize(size),
			path:          folderPath,
		})
	}

	sort.Slice(worlds, func(i, j int) bool {
		return worlds[i].lastSaved.After(worlds[j].lastSaved)
	})

	return nil
}

func openFolder(path string) {
	cmd := "explorer"
	args := []string{path}
	if os.Getenv("OS") == "Windows_NT" {
		cmd = "cmd"
		args = []string{"/c", "start", "", path}
	}
	execCmd := &exec.Cmd{
		Path: cmd,
		Args: args,
	}
	execCmd.Start()
}

func createListPanel(win fyne.Window) *fyne.Container {
	worldList = widget.NewList(
		func() int { return len(worlds) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("World Name"),
				container.NewHBox(
					widget.NewLabel("Folder: "),
					widget.NewLabel("path"),
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(worlds) {
				return
			}
			w := worlds[id]
			vbox := item.(*fyne.Container)
			label := vbox.Objects[0].(*widget.Label)
			label.SetText(w.name)
			hbox := vbox.Objects[1].(*fyne.Container)
			folderLabel := hbox.Objects[0].(*widget.Label)
			folderLabel.Text = "文件夹: "
			pathLabel := hbox.Objects[1].(*widget.Label)
			pathLabel.Text = w.folder
		},
	)

	worldList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(worlds) {
			w := worlds[id]
			dialog.ShowInformation("存档信息",
				fmt.Sprintf("文件夹: %s\n世界名: %s\n大小: %s\n路径: %s",
					w.folder, w.name, w.sizeFormatted, w.path),
				win)
		}
	}

	return container.NewBorder(
		nil, nil, nil, nil,
		worldList,
	)
}

func createToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			if err := loadWorlds(); err != nil {
				statusLabel.SetText("错误: " + err.Error())
			} else {
				statusLabel.SetText(fmt.Sprintf("共 %d 个存档", len(worlds)))
				worldList.Refresh()
			}
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			openFolder(getWorldsDir())
		}),
	)
}

func createStatusBar() *fyne.Container {
	statusLabel = widget.NewLabel(fmt.Sprintf("共 %d 个存档", len(worlds)))
	return container.NewHBox(
		statusLabel,
		layout.NewSpacer(),
		widget.NewLabel(fmt.Sprintf("存档目录: %s", getWorldsDir())),
	)
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("网易MC存档管理 - Fyne")
	w.Resize(fyne.NewSize(900, 600))

	if err := loadWorlds(); err != nil {
		fmt.Fprintf(os.Stderr, "加载存档失败: %v\n", err)
	}

	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("搜索存档...")

	toolbar := createToolbar()
	worldListContainer := createListPanel(w)
	statusBar := createStatusBar()

	mainContent := container.NewBorder(
		container.NewVBox(
			toolbar,
			searchEntry,
		),
		statusBar,
		nil, nil,
		worldListContainer,
	)

	w.SetContent(mainContent)

	w.ShowAndRun()
}
