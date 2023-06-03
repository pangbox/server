// Copyright (C) 2018-2023, John Chadwick <john@jchw.io>
//
// Permission to use, copy, modify, and/or distribute this software for any purpose
// with or without fee is hereby granted, provided that the above copyright notice
// and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
// OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
// THIS SOFTWARE.
//
// SPDX-FileCopyrightText: Copyright (c) 2018-2023 John Chadwick
// SPDX-License-Identifier: ISC

//go:build windows

package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/pangbox/server/cmd/minibox/lang/dict"
	"github.com/pangbox/server/minibox"
	"github.com/pangbox/server/res"
	log "github.com/sirupsen/logrus"
	"github.com/xo/dburl"
	"golang.org/x/sys/windows"
)

var logs bytes.Buffer
var mw *walk.MainWindow
var ni *walk.NotifyIcon
var lang *dict.Dict

//go:embed status-healthy.png
var statusHealthyPNG []byte

//go:embed status-unhealthy.png
var statusUnhealthyPNG []byte

//go:embed status-offline.png
var statusOfflinePNG []byte

func mustDecodeToBitmap(data []byte) *walk.Bitmap {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		fatalErrorf("error decoding PNG icon resource: %v", err)
	}
	bmp, err := walk.NewBitmapFromImageForDPI(img, 96*2)
	if err != nil {
		fatalErrorf("error creating bitmap from PNG icon: %v", err)
	}
	return bmp
}

var (
	statusHealthy   = mustDecodeToBitmap(statusHealthyPNG)
	statusUnhealthy = mustDecodeToBitmap(statusUnhealthyPNG)
	statusOffline   = mustDecodeToBitmap(statusOfflinePNG)

	pangboxIcon = mustDecodeToBitmap(res.PangboxPNG)
)

var pangyaRegions = []string{
	"",
	"us",
	"jp",
	"th",
	"eu",
	"id",
	"kr",
}

func tr(source string) string {
	if translation := walk.TranslationFunc(); translation != nil {
		return translation(source)
	}
	return source
}

func regionIndex(opt string) int {
	for i, n := range pangyaRegions {
		if n == opt {
			return i
		}
	}
	return 0
}

type statusController struct {
	image  *walk.ImageView
	button *walk.PushButton
}

func (s *statusController) SetRunning(running bool) {
	if s.image == nil || s.button == nil {
		// UI has not loaded yet.
		return
	}
	if running {
		s.button.SetText(tr("Stop"))
		s.image.SetImage(statusHealthy)
	} else {
		s.button.SetText(tr("Start"))
		s.image.SetImage(statusOffline)
	}
}

type ServiceController interface {
	Running() bool
	Start() error
	Stop() error
}

func pollingStatusController(ctx context.Context, server ServiceController) func(status *statusController) {
	return func(status *statusController) {
		go func() {
			t := time.NewTicker(time.Second / 5)
			for {
				select {
				case <-t.C:
					status.SetRunning(server.Running())
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func toggleServer(server ServiceController) func() {
	return func() {
		if server.Running() {
			server.Stop()
		} else {
			server.Start()
		}
	}
}

func serverStatus(name string, cb func(s *statusController), onClick func()) Widget {
	statusController := &statusController{}

	cb(statusController)

	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			ImageView{
				Image:    statusOffline,
				AssignTo: &statusController.image,
			},
			TextLabel{
				Text: name,
			},
			HSpacer{},
			PushButton{
				Text:      tr("Stop"),
				AssignTo:  &statusController.button,
				OnClicked: onClick,
			},
		},
	}
}

func patchStatus(ctx context.Context, patcher *minibox.RugburnPatcher) Widget {
	var unpatchButton *walk.PushButton
	var patchButton *walk.PushButton
	var label *walk.TextLabel
	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				// this may happen while still loading.
				if unpatchButton == nil || patchButton == nil || label == nil {
					continue
				}
				haveOrig := patcher.HaveOriginal()
				ver, err := patcher.RugburnVersion()

				patchButton.SetEnabled(haveOrig)
				unpatchButton.SetEnabled(haveOrig)

				if err != nil {
					label.SetText(tr("Status: Error (Set pangya dir?)"))
				} else if ver == "unpatched" {
					label.SetText(tr("Status: Unpatched"))
				} else if ver == "unknown" && !haveOrig {
					label.SetText(tr("Status: Unknown ijl15 patch"))
				} else if ver == "unknown" && haveOrig {
					label.SetText(tr("Status: Unknown Rugburn version"))
				} else {
					label.SetText(fmt.Sprintf(tr("Status: Patched (ver: %s)"), strings.Trim(ver, "\000")))
					patchButton.SetText(tr("Re-&patch"))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			TextLabel{
				MinSize:  Size{Width: 250},
				Text:     tr("Checking..."),
				AssignTo: &label,
			},
			HSpacer{},
			PushButton{
				Text:     tr("&Unpatch"),
				Enabled:  false,
				AssignTo: &unpatchButton,
				OnClicked: func() {
					if err := patcher.Unpatch(); err != nil {
						walk.MsgBox(mainForm(), tr("Error"), fmt.Sprintf(tr("Unpatching failed: %v"), err), walk.MsgBoxIconError)
					}
				},
			},
			PushButton{
				Text:     tr("&Patch"),
				Enabled:  false,
				AssignTo: &patchButton,
				OnClicked: func() {
					if err := patcher.Patch(); err != nil {
						walk.MsgBox(mainForm(), tr("Error"), fmt.Sprintf(tr("Patching failed: %v"), err), walk.MsgBoxIconError)
					}
				},
			},
		},
	}
}

func textOption(name string, option *string) Widget {
	var te *walk.LineEdit
	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			TextLabel{
				Text:    name,
				MinSize: Size{Width: 100},
			},
			LineEdit{
				AssignTo: &te,
				Text:     *option,
				OnTextChanged: func() {
					*option = te.Text()
				},
			},
		},
	}
}

func runMainWindow(ctx context.Context, minibox *minibox.Server) {
	var dbTE, pyTE *walk.LineEdit
	var curLang = language

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pakFiles, err := fs.Glob(os.DirFS("."), "*.pak")
	if err != nil {
		log.Printf("Unexpected error globbing pak files: %v", err)
	}

	mainGroup := GroupBox{
		Title:    tr("Main Configuration"),
		Layout:   VBox{},
		Children: []Widget{},
	}

	dbGroup := GroupBox{
		Title:    tr("Database Configuration"),
		Layout:   VBox{},
		Children: []Widget{},
	}

	netGroup := GroupBox{
		Title:  tr("Listen Addresses"),
		Layout: VBox{SpacingZero: true},
		Children: []Widget{
			textOption(tr("Web Server"), &opts.WebAddr),
			textOption(tr("Admin Server"), &opts.AdminAddr),
			textOption(tr("Login Server"), &opts.LoginAddr),
			textOption(tr("Game Server"), &opts.GameAddr),
			textOption(tr("Message Server"), &opts.MessageAddr),
			textOption(tr("QA Auth Server"), &opts.QAAuthAddr),
			VSpacer{},
		},
	}

	nameGroup := GroupBox{
		Title:  tr("Server Names"),
		Layout: VBox{SpacingZero: true},
		Children: []Widget{
			textOption(tr("Server Name"), &opts.GameServerName),
			textOption(tr("Channel Name"), &opts.GameChannelName),
			VSpacer{},
		},
	}

	dbGroup.Children = append(dbGroup.Children, TextLabel{
		Text:    tr("The database will be created automatically if it does not exist."),
		MinSize: Size{Width: 1, Height: 1},
	})

	// Database URI control
	dbGroup.Children = append(dbGroup.Children, Composite{
		Layout: HBox{},
		Children: []Widget{
			TextLabel{
				Text: tr("Database URI"),
			},
			LineEdit{
				AssignTo: &dbTE,
				Text:     dbOpts.DatabaseURI,
				OnTextChanged: func() {
					dbOpts.DatabaseURI = dbTE.Text()
				},
			},
			PushButton{
				Text: tr("Browse"),
				OnClicked: func() {
					dlg := new(walk.FileDialog)
					dlg.Filter = tr("SQLite3 Database (*.sqlite3)|*.sqlite3")
					dlg.Title = tr("Set SQLite3 Database")

					// Try to preload file dialog with current DSN.
					if currentUrl, err := dburl.Parse(dbOpts.DatabaseURI); err == nil {
						if absPath, err := filepath.Abs(currentUrl.DSN); err == nil {
							dlg.FilePath = absPath
						}
					}

					if ok, err := dlg.ShowOpen(mainForm()); err != nil {
						log.Printf("Unexpected error in folder dialog: %v", err)
					} else if ok {
						url := "sqlite://" + dlg.FilePath
						dbTE.SetText(url)
						dbOpts.DatabaseURI = url
					}
				},
			},
		},
	})

	dbGroup.Children = append(dbGroup.Children, VSpacer{})

	// Pak file hint
	if len(pakFiles) == 0 {
		mainGroup.Children = append(mainGroup.Children, TextLabel{
			Text:      tr("Note: Consider moving Minibox to your PangYa install directory and running from there."),
			TextColor: walk.RGB(255, 0, 0),
			Font:      Font{Bold: true},
			MinSize:   Size{Width: 1, Height: 1},
		})
	}

	var displayLanguageOptions = []string{
		tr("English"),
		tr("Japanese"),
	}

	var displayLanguages = []string{
		"en",
		"ja",
	}

	languageIndex := 0
	for i, n := range displayLanguages {
		if n == language {
			languageIndex = i
		}
	}

	var languageCB *walk.ComboBox
	mainGroup.Children = append(mainGroup.Children, Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			TextLabel{
				Text:    tr("Language"),
				MinSize: Size{Width: 100},
			},
			ComboBox{
				Model:        displayLanguageOptions,
				CurrentIndex: languageIndex,
				OnCurrentIndexChanged: func() {
					language = displayLanguages[languageCB.CurrentIndex()]
				},
				AssignTo: &languageCB,
			},
			HSpacer{},
		},
	})

	mainGroup.Children = append(mainGroup.Children, Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			TextLabel{
				Text:    tr("PangYa Path"),
				MinSize: Size{Width: 100},
			},
			LineEdit{
				AssignTo: &pyTE,
				Text:     opts.PangyaDir,
				OnTextChanged: func() {
					opts.PangyaDir = pyTE.Text()
				},
			},
			PushButton{
				Text: tr("Browse"),
				OnClicked: func() {
					dlg := new(walk.FileDialog)
					dlg.Title = tr("Set PangYa Path")
					if ok, err := dlg.ShowBrowseFolder(mainForm()); err != nil {
						log.Printf("Unexpected error in folder dialog: %v", err)
					} else if ok {
						pyTE.SetText(dlg.FilePath)
						opts.PangyaDir = dlg.FilePath
					}
				},
			},
		},
	})

	var pangyaRegionOptions = []string{
		tr("Auto-detect (slower)"),
		"US",
		"JP",
		"TH",
		"EU",
		"ID",
		"KR",
	}

	var regionCB *walk.ComboBox
	mainGroup.Children = append(mainGroup.Children, Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			TextLabel{
				Text:    tr("PangYa Region"),
				MinSize: Size{Width: 100},
			},
			ComboBox{
				Model:        pangyaRegionOptions,
				CurrentIndex: regionIndex(opts.PangyaRegion),
				OnCurrentIndexChanged: func() {
					opts.PangyaRegion = pangyaRegions[regionCB.CurrentIndex()]
				},
				AssignTo: &regionCB,
			},
			HSpacer{},
		},
	})

	mainGroup.Children = append(mainGroup.Children, VSpacer{})

	statusGroup := GroupBox{
		Title:  "Server Status",
		Layout: VBox{},
		Children: []Widget{
			serverStatus(tr("Topology Server"), pollingStatusController(ctx, minibox.Topology), toggleServer(minibox.Topology)),
			serverStatus(tr("Web Server"), pollingStatusController(ctx, minibox.Web), toggleServer(minibox.Web)),
			serverStatus(tr("Admin Server"), pollingStatusController(ctx, minibox.Admin), toggleServer(minibox.Admin)),
			serverStatus(tr("QA Auth Server"), pollingStatusController(ctx, minibox.QAAuth), toggleServer(minibox.QAAuth)),
			serverStatus(tr("Login Server"), pollingStatusController(ctx, minibox.Login), toggleServer(minibox.Login)),
			serverStatus(tr("Game Server"), pollingStatusController(ctx, minibox.Game), toggleServer(minibox.Game)),
			serverStatus(tr("Message Server"), pollingStatusController(ctx, minibox.Message), toggleServer(minibox.Message)),
			VSpacer{},
		},
	}

	var logTE *walk.TextEdit
	logView := TextEdit{
		ReadOnly: true,
		Text:     string(logs.Bytes()),
		AssignTo: &logTE,
	}

	t := time.NewTicker(time.Second / 5)
	defer t.Stop()
	go func() {
		for range t.C {
			if logTE != nil {
				logTE.SetText(string(logs.Bytes()))
			}
		}
	}()

	rugburnGroup := GroupBox{
		Title:  tr("Rugburn"),
		Layout: VBox{},
		Children: []Widget{
			TextLabel{
				MinSize: Size{Width: 200},
				Text:    tr("Patch your PangYa installation with Rugburn to bypass GameGuard and redirect network requests. Visit rugburn.gg for more information."),
			},
			LinkLabel{
				Alignment: AlignHNearVCenter,
				MaxSize:   Size{Width: 200},
				Text:      `<a href="https://rugburn.gg">https://rugburn.gg</a>`,
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					windows.ShellExecute(0, nil, windows.StringToUTF16Ptr(link.URL()), nil, nil, windows.SW_SHOWNORMAL)
				},
			},
			patchStatus(ctx, minibox.Rugburn),
		},
	}

	w := MainWindow{
		AssignTo: &mw,
		Title:    tr("Minibox - All-in-one PangYa Server"),
		Size:     Size{Width: 500, Height: 500},
		Layout:   VBox{},
		Children: []Widget{
			TabWidget{
				Pages: []TabPage{
					{Title: tr("Main"), Layout: VBox{}, Children: []Widget{mainGroup, statusGroup, rugburnGroup}},
					{Title: tr("Database"), Layout: VBox{}, Children: []Widget{dbGroup}},
					{Title: tr("Network"), Layout: VBox{}, Children: []Widget{netGroup, nameGroup}},
					{Title: tr("Logs"), Layout: VBox{}, Children: []Widget{logView}},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: tr("&Quit"),
						OnClicked: func() {
							// TODO:
							// - Should shut down gracefully
							// - If games are active, should prompt
							os.Exit(0)
						},
					},
					PushButton{
						Text: tr("&Save"),
						OnClicked: func() {
							minibox.ConfigureDatabase(dbOpts)
							minibox.ConfigureServices(opts)
							if err := saveConfiguration("minibox.json"); err != nil {
								msg := fmt.Sprintf(tr("Error saving to disk: %s; settings may not be retained next time you start Minibox."), err)
								walk.MsgBox(mainForm(), tr("Warning"), msg, walk.MsgBoxIconWarning)
							}
							// Restart window if language changed.
							if language != curLang {
								mw.Close()
								mw.Dispose()
								updateLang()
								return
							}
						},
					},
				},
			},
		},
	}

	w.Run()
}

func mainForm() walk.Form {
	if mw != nil {
		return mw
	}
	return nil
}

func fatalError(args ...any) {
	msg := fmt.Sprint(args...)
	walk.MsgBox(mainForm(), tr("Fatal Error"), msg, walk.MsgBoxIconError)
	log.Fatal(msg)
}

func fatalErrorf(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	walk.MsgBox(mainForm(), tr("Fatal Error"), msg, walk.MsgBoxIconError)
	log.Fatal(msg)
}

func updateLang() {
	var err error
	lang, err = dict.NewDict(language)
	if err != nil {
		// Disable translations.
		lang = nil
		walk.SetTranslationFunc(nil)
		return
	}
	walk.SetTranslationFunc(lang.Translation)
	if err := ni.SetToolTip(tr("Minibox PangYa Server")); err != nil {
		fatalError(err)
	}
}

func main() {
	noGui := flag.Bool("nogui", false, "Disables the GUI.")
	flag.Parse()

	if *noGui {
		cliMain()
		return
	}

	log.SetOutput(&logs)

	ctx := context.Background()
	log := log.WithContext(ctx)

	minibox := minibox.NewServer(ctx, log)
	dummy, err := walk.NewMainWindow()
	if err != nil {
		fatalError(err)
	}

	if err := loadConfiguration("minibox.json"); err != nil && !errors.Is(err, os.ErrNotExist) {
		msg := fmt.Sprintf(tr("Error loading settings from disk: %s"), err)
		walk.MsgBox(mainForm(), tr("Warning"), msg, walk.MsgBoxIconWarning)
	}

	ni, err = walk.NewNotifyIcon(dummy)
	if err != nil {
		fatalError(err)
	}
	defer ni.Dispose()

	updateLang()

	// Configure concurrently to reduce startup delay.
	go func() {
		minibox.ConfigureDatabase(dbOpts)
		if err := minibox.ConfigureServices(opts); err != nil {
			msg := fmt.Sprintf(tr("Error configuring services: %s - Either move Minibox to your PangYa directory, OR set the PangYa directory to the location of your client."), err)
			walk.MsgBox(mainForm(), tr("Warning"), msg, walk.MsgBoxIconWarning)
		}
	}()

	if err := ni.SetIcon(pangboxIcon); err != nil {
		fatalError(err)
	}

	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		if mw != nil && !mw.IsDisposed() {
			mw.Close()
			mw.Dispose()
		} else {
			runMainWindow(ctx, minibox)
		}
	})

	exitAction := walk.NewAction()
	if err := exitAction.SetText(tr("E&xit")); err != nil {
		fatalError(err)
	}

	exitAction.Triggered().Attach(func() {
		// Nuclear option, to exit fast.
		// Should fix this later to shut down gracefully.
		ni.SetVisible(false)
		os.Exit(0)
	})

	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		fatalError(err)
	}

	if err := ni.SetVisible(true); err != nil {
		fatalError(err)
	}

	if err := ni.ShowInfo(tr("Minibox PangYa Server"), tr("Click the tray icon to configure. Right click the tray icon to exit.")); err != nil {
		fatalError(err)
	}

	dummy.Run()
}
