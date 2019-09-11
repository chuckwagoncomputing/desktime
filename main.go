package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/widgets"
	"github.com/therecipe/qt/quickcontrols2"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
	"os"
	"fmt"
	"unsafe"
 "time"
)

type QmlBridge struct {
	core.QObject

	_ func() string            `slot:"getTime"`
	_ func()            `slot:"resetTime"`
	_ func(bool)            `slot:"setClockState"`
	_ func(int, bool)            `slot:"setDesktopState"`
	_ func(t string) `slot:"copyText"`
}

const (
	DesktopID = int(core.Qt__UserRole) + 1<<iota
	DesktopName
)

type DesktopModel struct {
	core.QAbstractListModel
	_ map[int]*core.QByteArray `property:"roles"`
	_ func()                   `constructor:"init"`

	_ []*Desktop     `property:"desktops"`
}

type Desktop struct {
	DesktopID int
	DesktopName string
}

var lastDesktop uint
var desktopModel *DesktopModel
var xconn *xgbutil.XUtil
var qmlBridge *QmlBridge
var clockEnabled bool
var clockRunning bool
var clockStart time.Time
var clockStop time.Time
var desktopStates []bool
var app *widgets.QApplication

func main() {
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)
	app = widgets.NewQApplication(len(os.Args), os.Args)
	quickcontrols2.QQuickStyle_SetStyle("material")
	view := qml.NewQQmlApplicationEngine(nil)
	qmlBridge = NewQmlBridge(nil)
	desktopModel = NewDesktopModel(nil)

 // Connect to X server
 x, err := xgbutil.NewConn()
	if err != nil {
		fmt.Println("Error", err)
  return
	}
	// This is a bodge because Go complains about xconn not
 //  being used if we create it when we connect.
 xconn = x

 eventFilter := core.NewQAbstractNativeEventFilter()
 eventFilter.ConnectNativeEventFilter(filterEvents)
	app.InstallNativeEventFilter(eventFilter)

	resetTime()

 dc := getDesktopCount()
	desktopStates = make([]bool, dc)
	ds := make([]*Desktop, dc)
 ns := getDesktopNames()
	for i, _ := range ds {
  ds[i] = &Desktop{DesktopID: i, DesktopName: ns[i]}
	}
 desktopModel.SetDesktops(ds)

	checkCurrentDesktop()

	qmlBridge.ConnectGetTime(getTime)
	qmlBridge.ConnectResetTime(resetTime)
	qmlBridge.ConnectSetClockState(setClockState)
	qmlBridge.ConnectSetDesktopState(setDesktopState)
	qmlBridge.ConnectCopyText(copyText)

	view.RootContext().SetContextProperty("DesktopModel", desktopModel)
	view.RootContext().SetContextProperty("QmlBridge", qmlBridge)

	view.Load(core.NewQUrl3("qrc:///qml/main.qml", 0))
	gui.QGuiApplication_Exec()
}

func filterEvents(t *core.QByteArray, m unsafe.Pointer, r *int) bool {
	checkCurrentDesktop()
 return false
}

func getDesktopCount() int {
 rw := xconn.RootWin()
 id, err := xprop.PropValNum(xprop.GetProperty(xconn, rw, "_NET_NUMBER_OF_DESKTOPS"))
	if err != nil {
		fmt.Println("Error", err)
  return 0
	}
 return int(id)
}

func getDesktopNames() []string {
 rw := xconn.RootWin()
 names, err := xprop.PropValStrs(xprop.GetProperty(xconn, rw, "_NET_DESKTOP_NAMES"))
	if err != nil {
		fmt.Println("Error", err)
  return nil
	}
 return names
}

func checkCurrentDesktop() {
 id := getCurrentDesktop()
	if (id != lastDesktop) {
  desktopsSwitched(id)
 }
}

func getCurrentDesktop() uint {
 rw := xconn.RootWin()
 id, err := xprop.PropValNum(xprop.GetProperty(xconn, rw, "_NET_CURRENT_DESKTOP"))
	if err != nil {
		fmt.Println("Error", err)
  return 0
	}
 return id + 1
}

func desktopsSwitched(n uint) {
	if (clockEnabled) {
	 if (desktopStates[n-1] && !clockRunning) {
			startClock()
	 } else if (!desktopStates[n-1] && clockRunning) {
			stopClock()
		}
	}
 lastDesktop = n
}

func getTime() string {
 var t time.Time
 if (clockRunning) {
  t = time.Now()
	} else {
	 t = clockStop
 }
 d := t.Sub(clockStart)
 return d.String()
}

func resetTime() {
	t := time.Now()
 clockStart = t
 clockStop = t
}

func setClockState(e bool) {
	clockEnabled = e
 if (!clockEnabled && clockRunning) {
		stopClock()
	} else {
	 desktopsSwitched(getCurrentDesktop())
	}
}

func setDesktopState(n int, e bool) {
 desktopStates[n] = e
 desktopsSwitched(uint(n)+1)
}

func startClock() {
	t := time.Now()
fmt.Println("foobar start", clockStart, "stop", clockStop)
	clockStart = t.Add(clockStart.Sub(clockStop))
fmt.Println("and", clockStart.Sub(clockStop).String())
fmt.Println("and", (clockStart.Sub(clockStop)).String())
	clockRunning = true
}

func stopClock() {
	clockStop = time.Now()
	clockRunning = false
}

func copyText(t string) {
	clipboard := app.Clipboard()
	clipboard.SetText(t, gui.QClipboard__Clipboard)
}

func (m *DesktopModel) init() {
	m.SetRoles(map[int]*core.QByteArray{
		DesktopID: core.NewQByteArray2("desktopId", len("desktopId")),
		DesktopName:      core.NewQByteArray2("desktopName", len("desktopName")),
	})
	m.ConnectData(m.data)
	m.ConnectRowCount(m.rowCount)
	m.ConnectRoleNames(m.roleNames)
}

func (m *DesktopModel) roleNames() map[int]*core.QByteArray {
	return m.Roles()
}

func (m *DesktopModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if !index.IsValid() {
		return core.NewQVariant()
	}
	if index.Row() >= len(m.Desktops()) {
		return core.NewQVariant()
	}

	i := m.Desktops()[index.Row()]

	switch role {
	case DesktopID:
		return core.NewQVariant5(i.DesktopID)
	case DesktopName:
		return core.NewQVariant12(i.DesktopName)
	default:
		return core.NewQVariant()
	}
}

func (m *DesktopModel) rowCount(parent *core.QModelIndex) int {
	return len(m.Desktops())
}
