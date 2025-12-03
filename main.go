package main

import (
	"fmt"
	"image/color"
	"machine"
	"machine/usb/hid/keyboard"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

// --- CONFIGURATION ---

// MATRIX PINS
var rowPins = []machine.Pin{
	machine.GP10, machine.GP11, machine.GP12, machine.GP13,
}
var colPins = []machine.Pin{
	machine.GP14, machine.GP15, machine.GP16,
}

// DISPLAY PINS (I2C0 Default)
// Connect OLED SDA to GP0, SCL to GP1
var (
	sdaPin = machine.GP4
	sclPin = machine.GP5
	keyMap = map[string]string{
		"2-0": "kubectl get pods -A",
		"2-1": "kubectl get nodes -A",
		"2-2": "kubectl get svc -A",
		"2-3": "kubectl describe pod",

		"1-0": "kubectl logs -f",
		"1-1": "kubectl get all -A",
		"1-2": "kubectl top nodes",
		"1-3": "kubectl version",

		"0-0": "kubectl apply -f .",
		"0-1": "kubectl delete pod",
		"0-2": "kubectl config view",
		"0-3": "clear",
	}
)

func main() {
	display := setupOLED(sdaPin, sclPin)

	setupKeyboardMatrix()

	kb := keyboard.New()

	// Clear screen initially
	display.ClearDisplay()
	tinyfont.WriteLine(display, &freemono.Regular12pt7b, 10, 20, "Kubepad Active!", color.RGBA{255, 255, 255, 255})
	display.Display()
	time.Sleep(1 * time.Second)

	// 3. MAIN LOOP
	for {
		for cIndex, col := range colPins {
			col.High()
			for rIndex, row := range rowPins {
				if row.Get() {
					time.Sleep(200 * time.Millisecond)
					// 1. Identify Key
					keyID := fmt.Sprintf("%d-%d", cIndex, rIndex)

					// 2. Lookup Command
					cmd, exists := keyMap[keyID]

					display.ClearDisplay()
					if exists {
						// Type the command
						// We iterate purely to type strings easily in TinyGo
						for _, char := range cmd {
							kb.Write([]byte{byte(char)})
						}
						// Press Enter
						kb.Write([]byte("\n"))

						tinyfont.WriteLine(display, &freemono.Regular12pt7b, 0, 15, cmd, color.RGBA{255, 255, 255, 255})

					} else {
						// Key defined in hardware but not in map
						tinyfont.WriteLine(display, &freemono.Regular12pt7b, 0, 30, "NO MAP:", color.RGBA{255, 255, 255, 255})
						tinyfont.WriteLine(display, &freemono.Regular12pt7b, 0, 50, keyID, color.RGBA{255, 255, 255, 255})
					}
					display.Display()

					// Debounce
					time.Sleep(300 * time.Millisecond)
				}
			}
			col.Low()
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func setupOLED(sda, sdk machine.Pin) *ssd1306.Device {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * 1000,
		SDA:       sdaPin,
		SCL:       sclPin,
	})

	// Initialize SSD1306 (Standard 128x64 size)
	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C, // Standard address, try 0x3D if this fails
	})

	return display
}

func setupKeyboardMatrix() {
	for _, p := range colPins {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
		p.Low()
	}
	for _, p := range rowPins {
		p.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
}
