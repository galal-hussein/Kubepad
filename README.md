# Kubepad ⎈

<img src="images/kubepad-logo.png" alt="drawing" width="200"/>

A programmable 12-key macropad designed specifically for Kubernetes operations, built with TinyGo for Raspberry Pi Pico. Features an OLED display for visual feedback and supports multiple modes for different command sets.

Kubepad aims to simplify working with Kubernetes clusters by providing quick access to common kubectl commands through dedicated physical keys.

## Features

- **12 Programmable Keys** - Cherry MX compatible switches (3 rows × 4 columns)
- **OLED Display** - 128x64 SSD1306 for command feedback
- **Multiple Modes** - Switch between kubectl, helm, and debug modes
- **USB HID** - Acts as a standard USB keyboard
- **Single Button Mode Switching** - Cycle through modes with bottom-right key

## Hardware

### Parts List

- **0.91" OLED Screen** (SSD1306, 128x64, I2C)
- **Raspberry Pi Pico** (or Pico W)
- **12x Cherry MX Switches**
- **12x 1N4148 Diodes**
- Optional: 3D printed case

### 3D Printed Case

Using the [Ocreb modular Macropad design](https://www.thingiverse.com/thing:6450013) with a small remix to remove the rotary encoder. STL files are in the [stls/](stls/) directory.

### Pinout

```
Key Matrix (3 rows × 4 columns):
- Rows: GP10, GP11, GP12, GP13 (scanned as columns in code)
- Cols: GP14, GP15, GP16 (scanned as rows in code)
- Key format: row-col (e.g., "2-3" = physical row 2, column 3)

OLED Display:
- SDA: GP4
- SCL: GP5
- I2C Address: 0x3C
```

### Key Layout

```
┌─────┬─────┬─────┬──────┐
│ 0-2 │ 1-2 │ 2-2 │ 3-2  │  ← Top row
├─────┼─────┼─────┼──────┤
│ 0-1 │ 1-1 │ 2-1 │ 3-1  │  ← Middle row
├─────┼─────┼─────┼──────┤
│ 0-0 │ 1-0 │ 2-0 │ 3-0  │  ← Bottom row (3-0 = MODE)
└─────┴─────┴─────┴──────┘
```

## Building the Hardware

Follow the excellent [GeekHack handwiring guide](https://geekhack.org/index.php?topic=87689.0) - it's a straightforward process if you know basic soldering:

### 1. Installing the diodes

<img src="images/1.jpg" alt="drawing" width="200"/>

Solder diodes to each switch with the cathode (black line) facing the row wire.

### 2. Soldering the rows and columns

<img src="images/2.jpg" alt="drawing" width="200"/>

Create the matrix by connecting rows and columns with wire.

### 3. Adding the Pico and connecting

<img src="images/3.jpg" alt="drawing" width="200"/>

Solder the matrix to the Pico according to the pinout above.

### 4. Putting it all together with the OLED

<img src="images/4.jpg" alt="drawing" width="200"/>

Connect the OLED to I2C pins and assemble the case.

### 5. Final look

<img src="images/5.jpg" alt="drawing" width="200"/>

## Software Setup

### Prerequisites

- [TinyGo](https://tinygo.org/getting-started/install/) installed

### Building and Flashing

```bash
# Build the firmware
tinygo build -target=pico -o kubepad.uf2 .

# Flash to Pico (put Pico in BOOTSEL mode first)
# Hold BOOTSEL button while plugging in USB
# Copy the UF2 file to the mounted drive
cp kubepad.uf2 /media/$USER/RPI-RP2/
```

## Usage

### Available Modes

The Kubepad has 3 modes that you can cycle through by pressing the **MODE** button (bottom-right key, position 3-0):

**Note:** Commands marked with `*` use `NoEnter` mode - they type the command but don't press Enter, allowing you to add arguments before executing.

#### Mode 0: kubectl (Basic Kubernetes)
```
┌─────────┬─────────┬─────────┬─────────┐
│  pods   │  nodes  │   svc   │  desc*  │
│ get -A  │   get   │  get -A │describe │
├─────────┼─────────┼─────────┼─────────┤
│  logs*  │   all   │   top   │ version │
│  -f     │  get -A │  nodes  │         │
├─────────┼─────────┼─────────┼─────────┤
│ apply*  │ delete* │ busybox │  MODE   │
│   -f    │   pod   │   run   │         │
└─────────┴─────────┴─────────┴─────────┘
```

#### Mode 1: helm
```
┌─────────┬─────────┬─────────┬─────────┐
│ status* │history* │rollback*│template*│
│         │         │         │         │
├─────────┼─────────┼─────────┼─────────┤
│  list   │  repo   │ search* │ values* │
│   -A    │  list   │  repo   │   get   │
├─────────┼─────────┼─────────┼─────────┤
│install* │upgrade* │ delete* │  MODE   │
│         │         │         │         │
└─────────┴─────────┴─────────┴─────────┘
```

#### Mode 2: debug
```
┌─────────┬─────────┬─────────┬─────────┐
│ events  │   pvc   │ ingress │ drain*  │
│ sorted  │  get -A │  get -A │         │
├─────────┼─────────┼─────────┼─────────┤
│top pod  │ api-res │ explain*│ cordon* │
│   -A    │         │         │         │
├─────────┼─────────┼─────────┼─────────┤
│  curl*  │netshoot │  exec*  │  MODE   │
│ pod run │pod run  │   -it   │         │
└─────────┴─────────┴─────────┴─────────┘
```

### Switching Modes

Press the **bottom-right key** (position 3-0, labeled MODE) to cycle through modes. The OLED will display the current mode name.

### NoEnter Feature

Commands marked with `*` have the `NoEnter` flag set, which means:
- The command is typed but Enter is **not** automatically pressed
- You can add arguments (pod names, chart names, etc.) before manually pressing Enter
- The OLED shows `...` after the label to indicate it's waiting for input

For example, pressing "desc" types `kubectl describe pod ` and waits for you to type the pod name.

### Key Mappings

Edit the `modes` array in [main.go](main.go:43) to customize your key mappings. Each mode has this structure:

```go
{
    Name: "kubectl",
    Keys: map[string]KeyMapping{
        "0-2": {Label: "pods", Command: "kubectl get pods -A"},
        "3-2": {Label: "desc", Command: "kubectl describe pod ", NoEnter: true},
        "3-0": {Label: "MODE", Command: ""}, // Mode switch
        // ... more keys
    },
},
```

## Customization

To add or modify commands, edit the `modes` variable in [main.go](main.go:43). The structure is simple:

- **Name**: Mode name shown on display
- **Keys**: Map of key positions (row-col) to commands
- **Label**: Short label for display (keep under 8 chars)
- **Command**: The actual command to type
- **NoEnter**: Set to `true` to wait for additional input before executing

Example - adding a new mode:

```go
{
    Name: "docker",
    Keys: map[string]KeyMapping{
        "0-2": {Label: "ps", Command: "docker ps"},
        "1-2": {Label: "images", Command: "docker images"},
        "0-1": {Label: "logs", Command: "docker logs -f ", NoEnter: true},
        "1-0": {Label: "exec", Command: "docker exec -it ", NoEnter: true},
        "3-0": {Label: "MODE", Command: ""},
        // ... etc
    },
},
```

## Development

### Project Structure

```
Kubepad/
├── main.go             # Main firmware (edit this!)
├── go.mod
├── go.sum
├── stls/              # 3D printable case files
├── images/            # Photos and diagrams
└── README.md
```

### Code Overview

The code in [main.go](main.go) is organized into:

- **Configuration** (lines 15-106) - Pin definitions and mode mappings
- **Kubepad struct** - Main state management
- **scanKeys()** - Key matrix scanning with debouncing
- **handleKeyPress()** - Key event handling and mode switching
- **executeCommand()** - Command execution via USB HID keyboard (respects NoEnter flag)
- **Display functions** - Startup screen, mode display, command feedback

## Troubleshooting

### Device Not Recognized

1. Make sure Pico is in BOOTSEL mode (hold button while connecting)
2. Check that `/media/$USER/RPI-RP2` is mounted
3. Try manually copying `kubepad.uf2` to the mounted drive

### Keys Not Working

1. Check key matrix wiring matches the pinout
2. Verify diode orientation (cathode toward rows)
3. Test individual keys to identify hardware issues

### OLED Not Working

1. Verify I2C connections (SDA=GP4, SCL=GP5)
2. Check I2C address (0x3C is standard, try 0x3D if needed)
3. Ensure proper power connection to OLED

## Future Ideas

- Add support for switching kubeconfig contexts
- Long press detection for alternate commands
- RGB LED support for mode indication
- Rotary encoder for namespace/context scrolling
- More modes (git, docker, terraform, etc.)

## Contributing

Contributions welcome! Feel free to:
- Add new default modes
- Improve the display layouts
- Add new features
- Share your custom key mappings
- Improve documentation

## License

MIT License - feel free to use, modify, and distribute!

## Acknowledgments

- Built with [TinyGo](https://tinygo.org/)
- Uses [tinygo-drivers](https://github.com/tinygo-org/drivers) for OLED
- [GeekHack handwiring guide](https://geekhack.org/index.php?topic=87689.0) for construction
- [Ocreb modular Macropad design](https://www.thingiverse.com/thing:6450013) for the case
- Inspired by the mechanical keyboard and Kubernetes communities

---

**Happy Kubepading! ⎈**
