import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Window 2.0

import io.github.greyvar.launcher 1.0

ApplicationWindow {
	id: page
	width: 320
	height: 180
	title: "Greyvar"
	visible: true
	x: Screen.desktopAvailableWidth / 2 - width / 2
	y: Screen.desktopAvailableHeight / 2 - height / 2

	BackEnd {
		id: backend

		onStatusChanged: {
			txtStatus.text = backend.status

			if (backend.canPlay) {
				txtStatus.color = "green"

				btnPlay.enabled = true
			} else {
				txtStatus.color = "red"

				btnPlay.enabled = false
			}
		}
	}

	Component.onCompleted: {
		backend.checkStatus()
	}

	Text {
		id: txtTitle
		text: "Greyvar"
		y: 30
		anchors.horizontalCenter: parent.horizontalCenter
		font.pointSize: 24;
		font.bold: true
	}

	Flow {
		y: 80
		anchors.horizontalCenter: parent.horizontalCenter

		Text {
			id: txtStatusPrefix
			text: "Status: "
		}


		Text {
			id: txtStatus
			color: "red"
			font.bold: true
			text: "waiting for backend"
		}
	}

    Button {
        id: btnPlay
        text: "Play"
		enabled: false
		anchors.horizontalCenter: parent.horizontalCenter
        y: 140

		onClicked: {
			console.log("Clicked play!")
		}
    }
}
