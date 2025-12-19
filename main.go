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

// MATRIX PINS (3 rows, 4 columns)
var rowPins = []machine.Pin{
	machine.GP10, machine.GP11, machine.GP12, machine.GP13,
}
var colPins = []machine.Pin{
	machine.GP14, machine.GP15, machine.GP16,
}

// DISPLAY PINS (I2C)
var (
	sdaPin = machine.GP4
	sclPin = machine.GP5
)

// MODE DEFINITIONS
type KeyMapping struct {
	Label   string
	Command string
	NoEnter bool // If true, don't auto-press Enter (useful for commands that need additional args)
}

type Mode struct {
	Name string
	Keys map[string]KeyMapping
}

var modes = []Mode{
	// Mode 0: Basic kubectl
	{
		Name: "kubectl",
		Keys: map[string]KeyMapping{
			"0-2": {Label: "pods", Command: "kubectl get pods -A"},
			"1-2": {Label: "nodes", Command: "kubectl get nodes"},
			"2-2": {Label: "svc", Command: "kubectl get svc -A"},
			"3-2": {Label: "desc", Command: "kubectl describe pod ", NoEnter: true},

			"0-1": {Label: "logs", Command: "kubectl logs -f ", NoEnter: true},
			"1-1": {Label: "all", Command: "kubectl get all -A"},
			"2-1": {Label: "top", Command: "kubectl top nodes"},
			"3-1": {Label: "version", Command: "kubectl version"},

			"0-0": {Label: "apply", Command: "kubectl apply -f ", NoEnter: true},
			"1-0": {Label: "delete", Command: "kubectl delete pod ", NoEnter: true},
			"2-0": {Label: "busybox", Command: "kubectl run busybox --image=busybox -it --rm -- sh"},
			"3-0": {Label: "MODE", Command: ""}, // Mode switch button
		},
	},
	// Mode 1: Helm
	{
		Name: "helm",
		Keys: map[string]KeyMapping{
			"0-2": {Label: "status", Command: "helm status ", NoEnter: true},
			"1-2": {Label: "history", Command: "helm history ", NoEnter: true},
			"2-2": {Label: "rollback", Command: "helm rollback ", NoEnter: true},
			"3-2": {Label: "template", Command: "helm template ", NoEnter: true},

			"0-1": {Label: "list", Command: "helm list -A"},
			"1-1": {Label: "repo", Command: "helm repo list"},
			"2-1": {Label: "search", Command: "helm search repo ", NoEnter: true},
			"3-1": {Label: "values", Command: "helm get values ", NoEnter: true},

			"0-0": {Label: "install", Command: "helm install ", NoEnter: true},
			"1-0": {Label: "upgrade", Command: "helm upgrade ", NoEnter: true},
			"2-0": {Label: "delete", Command: "helm delete ", NoEnter: true},
			"3-0": {Label: "MODE", Command: ""},
		},
	},
	// Mode 2: Debug
	{
		Name: "debug",
		Keys: map[string]KeyMapping{
			"0-2": {Label: "events", Command: "kubectl get events -A --sort-by='.lastTimestamp'"},
			"1-2": {Label: "pvc", Command: "kubectl get pvc -A"},
			"2-2": {Label: "ingress", Command: "kubectl get ingress -A"},
			"3-2": {Label: "drain", Command: "kubectl drain ", NoEnter: true},

			"0-1": {Label: "top pod", Command: "kubectl top pods -A"},
			"1-1": {Label: "api-res", Command: "kubectl api-resources"},
			"2-1": {Label: "explain", Command: "kubectl explain ", NoEnter: true},
			"3-1": {Label: "cordon", Command: "kubectl cordon ", NoEnter: true},

			"0-0": {Label: "curl", Command: "kubectl run curl --image=curlimages/curl -it --rm -- ", NoEnter: true},
			"1-0": {Label: "netshoot", Command: "kubectl run netshoot --image=nicolaka/netshoot -it --rm -- bash"},
			"2-0": {Label: "exec", Command: "kubectl exec -it ", NoEnter: true},
			"3-0": {Label: "MODE", Command: ""},
		},
	},
}

const modeSwitchKey = "3-0" // Bottom right key

// --- MAIN CODE ---

type Kubepad struct {
	display      *ssd1306.Device
	currentMode  int
	font         *tinyfont.Font
	lastKeyPress time.Time
}

var macropad = keyboard.New()

func main() {
	kp := &Kubepad{
		font:         &freemono.Bold9pt7b,
		currentMode:  0,
		lastKeyPress: time.Now(),
	}

	kp.display = setupOLED(sdaPin, sclPin)
	setupKeyboardMatrix()

	kp.showStartup()
	kp.run()
}

func (kp *Kubepad) showStartup() {
	kp.display.ClearDisplay()
	tinyfont.WriteLine(kp.display, kp.font, 2, 20, "Kubepad Ready!", color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(kp.display, kp.font, 2, 40, modes[kp.currentMode].Name, color.RGBA{150, 150, 150, 255})
	kp.display.Display()
	time.Sleep(1500 * time.Millisecond)
}

func (kp *Kubepad) run() {
	for {
		kp.scanKeys()
		time.Sleep(10 * time.Millisecond)
	}
}

func (kp *Kubepad) scanKeys() {
	for colIndex, col := range colPins {
		col.High()

		for rowIndex, row := range rowPins {
			if row.Get() {
				// Debounce
				if time.Since(kp.lastKeyPress) < 250*time.Millisecond {
					continue
				}

				kp.lastKeyPress = time.Now()
				keyID := fmt.Sprintf("%d-%d", rowIndex, colIndex)

				kp.handleKeyPress(keyID)

				// Wait for key release
				for row.Get() {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}

		col.Low()
	}
}

func (kp *Kubepad) handleKeyPress(keyID string) {
	// Check if mode switch button
	if keyID == modeSwitchKey {
		kp.switchMode()
		return
	}

	// Execute command for current mode
	currentMode := modes[kp.currentMode]
	if mapping, exists := currentMode.Keys[keyID]; exists {
		if mapping.Command != "" {
			kp.executeCommand(mapping)
		}
	} else {
		kp.showUnmapped(keyID)
	}
}

func (kp *Kubepad) switchMode() {
	kp.currentMode = (kp.currentMode + 1) % len(modes)

	kp.display.ClearDisplay()
	tinyfont.WriteLine(kp.display, kp.font, 2, 20, "Mode:", color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(kp.display, kp.font, 2, 40, modes[kp.currentMode].Name, color.RGBA{255, 255, 255, 255})
	kp.display.Display()

	time.Sleep(1000 * time.Millisecond)
}

func (kp *Kubepad) executeCommand(mapping KeyMapping) {
	// Type the command
	for _, char := range mapping.Command {
		macropad.Write([]byte{byte(char)})
	}

	// Only press Enter if NoEnter is false
	if !mapping.NoEnter {
		macropad.Write([]byte("\n"))
	}

	// Show on display
	kp.display.ClearDisplay()

	// Label
	labelText := fmt.Sprintf("[%s]", mapping.Label)
	if mapping.NoEnter {
		labelText += " ..." // Indicate waiting for input
	}
	tinyfont.WriteLine(kp.display, kp.font, 2, 15, labelText, color.RGBA{255, 255, 255, 255})

	// Mode indicator
	modeText := fmt.Sprintf("<%s>", modes[kp.currentMode].Name)
	tinyfont.WriteLine(kp.display, kp.font, 2, 55, modeText, color.RGBA{150, 150, 150, 255})

	kp.display.Display()
}

func (kp *Kubepad) showUnmapped(keyID string) {
	kp.display.ClearDisplay()
	tinyfont.WriteLine(kp.display, kp.font, 2, 20, "Not Mapped:", color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(kp.display, kp.font, 2, 45, keyID, color.RGBA{255, 255, 255, 255})
	kp.display.Display()
	time.Sleep(1000 * time.Millisecond)
}

func setupOLED(sda, scl machine.Pin) *ssd1306.Device {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * 1000,
		SDA:       sda,
		SCL:       scl,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C,
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
