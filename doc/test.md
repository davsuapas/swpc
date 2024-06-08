# TEST

The project can be tested locally, except for some functionalities. In local it is NOT possible to add samples for the generation of the predictive model, although it is possible to generate a model with random samples.

## TEST WITH VSCODE

### Generate model

[Generate model with random samples](./ai.md)

### Build

```shell
cd swpc
```

```shell
./scripts/build.sh
```

### Launch web server and iot hub into terminal

```shell
cd $GOPATH/swpc/release
SW_POOL_CONTROLLER_CONFIG='{"server":{"internal":{"host":"192.168.1.135"},"external":{"host":"192.168.1.135"}},"iot":{"configUi":true,"sampleUi":true}}' ./swpc-server
```

The *host* property replace it by you IP machine

### Launch web server and iot hub in VSCODE

Create a launch.json into .vscode and push play in *Run and Debug* menu:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/swpc-server/main.go",
      "env": {
        "SW_POOL_CONTROLLER_CONFIG": "{\"server\":{\"internal\":{\"host\":\"192.168.1.135\"},\"external\":{\"host\":\"192.168.1.135\"}},\"iot\":{\"configUi\":true,\"sampleUi\":true}}"
      }
    }
  ]
}
```

The *host* property replace it by you IP machine

### Deploy the micro-controller code

- Open swpc in vscode
- Connect board via the USB port.
- Select port into vscode
- Configure permission on port

~~~bash
sudo chmod 666 /dev/ttyUSB0
~~~

- Open micro/swpc/swpc.ino
- Configure

~~~c++
#define host ""

// WIFI definition
const char *ssid = "";
const char *password = "";

// Device ID
#define DeviceID ""
~~~

- Deploy code via *Arduino: Upload*
- To view the logs use the *serial monitor* at 115200 baud.

### Open web app

http://localhost:5000/
