import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Hyprland
import Quickshell.Widgets

PanelWindow {
    id: bar
    anchors {
        top: true
        left: true
        right: true
    }
    height: 40

    color: "transparent"

    Rectangle {
        anchors.fill: parent
        color: "#2b303b"
        opacity: 0.95
    }

    RowLayout {
        anchors.fill: parent
        anchors.margins: 8
        spacing: 8

        // Workspaces
        Repeater {
            model: HyprlandWorkspaces {
                id: workspaces
            }

            delegate: Button {
                required property var modelData
                property bool active: modelData.active

                text: modelData.name
                color: active ? "#0078d4" : "transparent"
                textColor: "white"

                onClicked: {
                    workspaces.switchTo(modelData.id)
                }
            }
        }

        Item { Layout.fillWidth: true }

        // System tray
        Text {
            text: "Tray"
            color: "white"
        }

        // Network
        Text {
            text: "Network"
            color: "white"
        }

        // Audio
        Text {
            text: "Audio"
            color: "white"
        }

        // CPU
        Text {
            text: "CPU"
            color: "white"
        }

        // Memory
        Text {
            text: "Memory"
            color: "white"
        }

        // Battery
        Text {
            text: "Battery"
            color: "white"
        }

        // Clock
        Text {
            text: new Date().toLocaleTimeString(Qt.locale(), "HH:mm")
            color: "white"
            Timer {
                interval: 1000
                running: true
                repeat: true
                onTriggered: parent.text = new Date().toLocaleTimeString(Qt.locale(), "HH:mm")
            }
        }
    }
}