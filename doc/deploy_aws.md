# DEPLOY ON AWS BEANSTALK

We are not going to explain how to deploy on beanstalk, we will explain how to generate the source.zip with everything needed to run the system inside beanstalk.

The source.zip contains:

- Web server and IOT Hub binaries
- SSL configuration on port 443
- Certificates
- Configuration of the static web files
- nginx proxy configuration

```shell
cd swpc
```

### Create self-certificate

Only for test environments we can generate a self-certificate. To do so, we can execute the following command:

```shell
./scripts/self-certificate.sh
```

The certificate is generated into folder: *swpc/deploy/self-certificate* 

### Build source.zip

```shell
./scripts/deploy-awsbt.sh -build
```

The source.zip file is generated into the path: *$GOPATH/swpc/awsbt*

### Web server and IOT Hub configuration

Most of the functionality can be configured through the *SW_POOL_CONTROLLER_CONFIG* environment variable. This variable contains a [configuration json](../internal/config/config.go). In AWS Beanstalk there is a section to include environment variables that the system automatically injects into our server.

An example of such a configuration could be the following:

```json
{
  "cloud": {
    "provider": "aws",
    "aws": {"region": "eu-west-1"}
  },
  "server": {
    "Internal": {"port": 5001, "host": "localhost", "tls": false},
    "External": {
      "port": 0,
      "host": "swpc.eu-west-1.elasticbeanstalk.com",
      "tls": true
    }
  },
  "data": {
    "provider": "cloud",
    "aws": {"configTableName": "", "samplesTableName": "swpc_sample"}
  },
  "log": {"development": false, "level": 0},
  "web": {
    "secretKey": "12345678901234567890123456789012",
    "auth": {
      "jwkUrl": "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_ffffff/.well-known/jwks.json",
      "clientId": "1223334343434343434343",
      "tokenUrl": "https://swpc.auth.eu-west-1.amazoncognito.com/oauth2/token",
      "provider": "oauth2",
      "loginUrl": "https://swpc.auth.eu-west-1.amazoncognito.com/login?client_id=%client_id&response_type=code&scope=email+openid&state=%state&redirect_uri=%redirect_uri",
      "logoutUrl": "https://swpc.auth.eu-west-1.amazoncognito.com/logout?client_id=%client_id&logout_uri=%redirect_uri"
    }
  },
  "location": {"zone": "Europe/Madrid"},
  "api": {
    "clientId": "323232323232322323",
    "tokenSecretKey": "4343434343434343434343",
    "heartbeatInterval": 120
  },
  "iot": {"configUi": false, "sampleUi": true}
}
```
