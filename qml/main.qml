import Qt.labs.platform 1.1
import QtQml 2.13

SystemTrayIcon {
	visible: true
	icon.source: "qrc:/qml/images/clock.png"

 onActivated: {
		menu.open()
	}

	menu: Menu {
		id: statusMenu

  Instantiator {
   model: DesktopModel
		 onObjectAdded: statusMenu.insertItem( index, object )
   onObjectRemoved: statusMenu.removeItem( object )
   delegate: MenuItem {
				property bool desktopEnabled: false
   	text: (desktopEnabled ? "✓" : " ") + model.desktopName
    onTriggered: {
					desktopEnabled = !desktopEnabled
					QmlBridge.setDesktopState(model.desktopId, desktopEnabled)
    }
  	}
  }

		MenuSeparator {}
  MenuItem{
			id: enabledItem
			property bool clockEnabled: false
			text: clockEnabled ? "Disable" : "Enable"
			onTriggered: {
				clockEnabled = !clockEnabled
				QmlBridge.setClockState(clockEnabled)
			}
		}
  MenuItem{
			id: elapsedTime
			text: QmlBridge.getTime()
			onTriggered: {
				QmlBridge.copyText(text)
			}
		}
  MenuItem{
			text: "Reset"
			onTriggered: {
				QmlBridge.resetTime()
			}
		}
		MenuItem {
   text: "Quit"
   onTriggered: Qt.quit()
  }

		onAboutToShow: {
			elapsedTime.text = QmlBridge.getTime()
		}
 }
}
