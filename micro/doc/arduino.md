# Arduino IDE configuration in LINUX

## Prerequisites

The pyserial python module needs to be installed.

Install pip if it does not exist: 

~~~bash
python3 -m pip install pyserial
~~~

Install pyserial

~~~bash
sudo apt install python3-pip
~~~

## Arduino IDE installation

https://www.arduino.cc/en/software

Install into $HOME/Arduino folder

## Visual Studio Code plugin Arduino installation

Install plugin: https://marketplace.visualstudio.com/items?itemName=vsciot-vscode.vscode-arduino

Configure the following settings in the file user preferences

~~~json
  "arduino.enableUSBDetection": true,
  "arduino.useArduinoCli": true,
~~~


## ESP32 Driver (AZ-Delivery ESP32 DevKitC V2)

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

## Board selection 

Select "ESP32 Dev Module" and port

## Permissions

Configure permission on port

~~~bash
sudo chmod 666 /dev/ttyUSB0
~~~

## Post configuration

### VSCode IDE

Create the folder *micro/build*

After opening the swpc.ino file and clicking on the Arduino:Verify button,
configure the following option within the file *.vscode/arduino.json*

~~~json
"output": "micro/build"
~~~

