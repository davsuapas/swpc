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

### Launch web server and iot hub

```shell
cd $GOPATH/swpc/release
./swpc-server
```

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

### Open app

http://localhost:5000/
