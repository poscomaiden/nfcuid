package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
)

var (
	serviceCtx    context.Context
	serviceCancel context.CancelFunc
	running       bool
	statusMenu    *systray.MenuItem
	deviceMenu    *systray.MenuItem
	currentFlags  Flags
)

func onReady() {
	// Set initial icon (stopped state)
	setIconForStatus("stopped")
	systray.SetTooltip("NFC UID Reader - Stopped")
	systray.SetTitle("NFC UID")

	// Status menu item
	statusMenu = systray.AddMenuItem("Status: Stopped", "Current status")
	statusMenu.Disable()

	// Device menu item
	deviceMenu = systray.AddMenuItem("Device: Not selected", "Selected device")
	deviceMenu.Disable()

	systray.AddSeparator()

	// Start/Stop menu item
	mToggle := systray.AddMenuItem("Start", "Start reading NFC tags")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Create context for service
	serviceCtx, serviceCancel = context.WithCancel(context.Background())

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				if !running {
					startService()
					mToggle.SetTitle("Stop")
					mToggle.SetTooltip("Stop reading NFC tags")
				} else {
					stopService()
					mToggle.SetTitle("Start")
					mToggle.SetTooltip("Start reading NFC tags")
				}
			case <-mQuit.ClickedCh:
				stopService()
				systray.Quit()
				return
			}
		}
	}()

	// Handle system signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		stopService()
		systray.Quit()
	}()
}

func onExit() {
	stopService()
}

func startService() {
	if running {
		return
	}
	running = true
	statusMenu.SetTitle("Status: Running...")
	setIconForStatus("running")
	systray.SetTooltip("NFC UID Reader - Running...")
	
	go func() {
		service := NewService(currentFlags)
		service.StartAsync(serviceCtx)
	}()
}

func stopService() {
	if !running {
		return
	}
	running = false
	if serviceCancel != nil {
		serviceCancel()
	}
	statusMenu.SetTitle("Status: Stopped")
	setIconForStatus("stopped")
	systray.SetTooltip("NFC UID Reader - Stopped")
}

// setIconForStatus sets the icon based on the current status
// Status options: "stopped", "running"
func setIconForStatus(status string) {
	iconData := getIconDataForStatus(status)
	if iconData != nil {
		systray.SetIcon(iconData)
	}
}

// getIconDataForStatus returns icon data for the specified status
// Supports .ico files (preferred) and .png files as fallback
func getIconDataForStatus(status string) []byte {
	// Try .ico files first (native Windows format)
	icoFiles := map[string]string{
		"stopped": "icon_stopped.ico",
		"running": "icon_running.ico",
	}

	// Try status-specific .ico file
	if filename, ok := icoFiles[status]; ok {
		if data, err := ioutil.ReadFile(filename); err == nil {
			return data
		}
	}

	// Fallback: try .png files
	pngFiles := map[string]string{
		"stopped": "icon_stopped.png",
		"running": "icon_running.png",
	}

	if filename, ok := pngFiles[status]; ok {
		if data, err := ioutil.ReadFile(filename); err == nil {
			return data
		}
	}

	// Fallback: try generic icon.ico or icon.png
	if status == "stopped" {
		if data, err := ioutil.ReadFile("icon.ico"); err == nil {
			return data
		}
		if data, err := ioutil.ReadFile("icon.png"); err == nil {
			return data
		}
	}

	// Final fallback: return default icon
	return getDefaultIcon()
}


// getDefaultIcon returns a minimal default icon as a fallback
func getDefaultIcon() []byte {
	// This is a minimal 16x16 PNG icon - a simple blue square with "N"
	// You can replace this with your own icon bytes or use the file loading method above
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0xF3, 0xFF, 0x61, 0x00, 0x00, 0x00,
		0x19, 0x74, 0x45, 0x58, 0x74, 0x53, 0x6F, 0x66, 0x74, 0x77, 0x61, 0x72,
		0x65, 0x00, 0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x49, 0x6D, 0x61, 0x67,
		0x65, 0x52, 0x65, 0x61, 0x64, 0x79, 0xCC, 0x35, 0x8C, 0x59, 0x00, 0x00,
		0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00,
		0x00, 0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00,
		0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

