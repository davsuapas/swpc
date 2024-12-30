# DEPLOY ON VPS

We are not going to explain how to deploy on virtual private Server. We will explain how to generate the scripts that will help us to create a service and how to update it. We will not explain how to expose that service to the outside.

### Build artifacts and scripts.

```shell
./scripts/deploy-vps.sh <path/to/swpc.env> -build
```

This script generates a folder in $GOPATH/swpc/vps containing the binaries and two main scripts:

- install.sh: Installs the service binaries in */opt/swpc*, creates the configuration file in */etc/swpc/swpc.env*, creates the */var/log/swpc* folder with the application log and creates the */var/lib/swpc* folder with the data files. The script creates a swpc user with the necessary permissions for the service to work. The script automatically creates and starts the service.
- update-files.sh: Updates the binaries and restarts the service.

Once the vps folder is generated, it is necessary to copy the content to the vps machine to be able to execute the previous scripts.

### swpc.env

Most of the functionality can be configured through the *SW_POOL_CONTROLLER_CONFIG* environment variable. This variable contains a [configuration json](../internal/config/config.go).

The swpc.env file must have the following format, which is an example for us:

```file
SW_POOL_CONTROLLER_CONFIG={"cloud":{"provider":"aws","aws":{"region":"eu-west-1"}},"server":{"Internal":{"port":5000,"host":"localhost","tls":false},"External":{"port":0,"host":"swpc.vps.cloud","tls":true}},"data":{"provider":"file","file":{"config":"/var/lib/swpc/config.dat","sample":"/var/lib/swpc/sample.csv"}},"log":{"development":false,"level":-1},"web":{"secretKey":"secret","auth":{"jwkUrl":"https:\/\/cognito-idp.eu-west-1.amazonaws.com\/eu-west-dddddddd\/.well-known\/jwks.json","clientId":"id","tokenUrl":"https:\/\/swpc.auth.eu-west-1.amazoncognito.com\/oauth2\/token","provider":"oauth2","loginUrl":"https:\/\/swpc.auth.eu-west-1.amazoncognito.com\/login?client_id=%client_id&response_type=code&scope=email+openid&state=%state&redirect_uri=%redirect_uri","logoutUrl":"https:\/\/swpc.auth.eu-west-1.amazoncognito.com\/logout?client_id=%client_id&logout_uri=%redirect_uri"}},"location":{"zone":"Europe\/Madrid"},"api":{"clientId":"id","tokenSecretKey":"token","heartbeatInterval":120},"iot":{"configUi":true,"sampleUi":true}}
```
