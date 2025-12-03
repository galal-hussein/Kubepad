# KubePad

<img src="images/kubepad-logo.png" alt="drawing" width="200"/>

kubepad is a handwired macropad that aimed to simplify writing `kubectl` commands and working with Kubernetes cluster.

## Parts

- 0.91" oled screen
- Pi Pico W
- 12 Cherry MX switches
- 12 1N4148 Diodes

## 3D Parts

I am using [Ocreb modular Macropad design](https://www.thingiverse.com/thing:6450013) with a small remix to remove the rotary encoder.

## Wiring

I followed the https://geekhack.org/index.php?topic=87689.0 guide to handwiring a keyboard, its pretty straight process if you know basic soldering:

### Installing the diodes

<img src="images/1.jpg" alt="drawing" width="200"/>

### Soldering the rows and coloumns

<img src="images/2.jpg" alt="drawing" width="200"/>

### Adding the pico and soldering the cols/rows to it

<img src="images/3.jpg" alt="drawing" width="200"/>

### Putting it all together with the oled

<img src="images/4.jpg" alt="drawing" width="200"/>

## Software

I decided to use tinygo for this project, the software just maps the keys to specific kubectl commands for now.

## TODO

- Add config file to map commands to the keyboard.
- Configure a button to switch between kubeconfig contexts.

## Final look

<img src="images/5.jpg" alt="drawing" width="200"/>