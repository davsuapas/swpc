### Clone repository

```shell
git clone https://github.com/davsuapas/swpc.git
```

# DEVELOP ENVIRONMENT

### Install npm

https://docs.npmjs.com/downloading-and-installing-node-js-and-npm

### Install Go v1.22

Install go into $HOME

https://go.dev/dl/

> [!NOTE]
> Remember to set `export GOPATH="$HOME/go/bin"` into .profile or .bashrc

### Install go linter

```shell
# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

golangci-lint --version
```

### Install python3

```shell
sudo apt update
sudo apt install python3
sudo apt install python3.10-venv
```
### Install graphviz to show decision tree

```shell
sudo apt install graphviz
```

### Install UI npm package

```shell
cd swpc/ui
npm install
```

# ESP32 BOARD DEVELOP ENVIRONMENT

### Pre-requisites

The pyserial python module needs to be installed.

Install pip if it does not exist: 

~~~bash
sudo apt install python3-pip
~~~

Install pyserial

~~~bash
python3 -m pip install pyserial
~~~

### Arduino IDE installation

- Install app imagen from https://www.arduino.cc/en/software

- Copy imagen into $HOME/Arduino folder

### Visual Studio Code plugin Arduino installation

Install plugin: https://marketplace.visualstudio.com/items?itemName=vsciot-vscode.vscode-arduino

Configure the following settings in the file user preferences

~~~json
  "arduino.enableUSBDetection": true,
  "arduino.useArduinoCli": true,
~~~

### ESP32 Driver (AZ-Delivery ESP32 DevKitC V2)

### Arduino IDE

This setting can be found in the menu "File" -> "Preferences". In the
entry field "Additional URLs of the board administrator:" 
the following URL must be entered the following URL:

https://dl.espressif.com/dl/package_esp32_index.json

The corresponding board definitions are downloaded and installed
in the Arduino IDE board manager The corresponding dialog can be accessed
via the menu "Tools" -> "Board:" -> "Board Manager".

As soon as you enter "32" in the search field, the package "esp32
by Espressif Systems" appears. By clicking on the "Install" button,
the necessary components are downloaded and are
components are downloaded and are immediately available in the Arduino IDE.

### VSCode IDE

Configure the following settings in the file user preferences

~~~json
  "arduino.additionalUrls": [
    "https://dl.espressif.com/dl/package_esp32_index.json"
  ]
~~~

Install "esp32 by Expressif" using (shift+ctrl+P) *Arduino: Board manager*

### Board selection 

Select "ESP32 Dev Module"

### Add libraries using (shift+ctrl+P) *Arduino: Libary Manager* 

- WebSockets by Markus Sattler
- ArduinoJson by Benoit Blanchon
- DallasTemperature by Miles Burton <miles@mnetcs.com>, Tim Newsome
- Download and copy to $HOME/Arduino, if not exist (https://www.arduino.cc/reference/en/libraries/onewire)

### Post configuration

### VSCode IDE

Create the folder *micro/build*

After opening the *swpc.ino* file and clicking on the *Arduino:Verify button*, 
configure the following option within the file *.vscode/arduino.json*

~~~json
"output": "micro/build"
~~~

