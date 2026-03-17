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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var Version = "0.1.9"

type WorldInfo struct {
	folder        string
	name          string
	lastSaved     time.Time
	size          uint64
	sizeFormatted string
	path          string
}

var worlds []WorldInfo
var worldTable *widget.Table
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
	template := container.NewHBox(
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel(""),
	)

	worldTable = widget.NewTable(
		func() (int, int) { return len(worlds), 5 },
		func() fyne.CanvasObject {
			t := template.MinSize()
			template.Resize(fyne.NewSize(t.Width, 30))
			return template
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row >= len(worlds) {
				return
			}
			w := worlds[id.Row]
			hbox := cell.(*fyne.Container)
			label := hbox.Objects[id.Col].(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(w.name)
			case 1:
				label.SetText(w.folder)
			case 2:
				label.SetText(w.sizeFormatted)
			case 3:
				label.SetText("📋")
			case 4:
				label.SetText("📁")
			}
		},
	)
	worldTable.SetColumnWidth(0, 200)
	worldTable.SetColumnWidth(1, 150)
	worldTable.SetColumnWidth(2, 80)
	worldTable.SetColumnWidth(3, 40)
	worldTable.SetColumnWidth(4, 40)

	header := container.NewHBox(
		widget.NewLabel("世界名"),
		widget.NewLabel(""),
		widget.NewLabel("文件夹"),
		widget.NewLabel(""),
		widget.NewLabel("大小"),
		widget.NewLabel(""),
		widget.NewLabel("操作"),
		widget.NewLabel(""),
	)

	return container.NewVBox(
		header,
		worldTable,
	)
}

func createToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			if err := loadWorlds(); err != nil {
				statusLabel.SetText("错误: " + err.Error())
			} else {
				statusLabel.SetText(fmt.Sprintf("共 %d 个存档", len(worlds)))
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
