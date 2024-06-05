# DEPLOY ON AWS BEANSTALK

We are not going to explain how to deploy in beanstalk, we will explain how to generate the source.zip with everything needed to run the system inside beanstalk.

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
./scripts/deploy-awsbt.sh -build
```
