/*
 *   Copyright (c) 2022 ELIPCERO
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

#include <Arduino.h>
#include <ArduinoJson.h>
#include <DallasTemperature.h>
#include <HTTPClient.h>
#include <OneWire.h>
#include <WebSocketsClient.h>
#include <WiFi.h>

// BEGIN user configuration area
// rootCACertificate = nullptr y ssl = true
// it's used for self-certification
const char *rootCACertificate = nullptr;

#define ssl false
#define host "192.168.1.135"
#define port 5000

#define URIAPI "/api/device/ws"
#define URIToken "/auth/token/"

// clientID define the ID to connect to the server
#define clientID "sw3kf$fekdy56dfh"

// WIFI definition
const char *ssid = "";
const char *password = "";

// Device ID
#define DeviceID ""

// BEGIN Sensors configuration
// Pin to Temp is defined directly in setup function
// Pin to PH sensor
#define pinPH 34
// Reference voltage value for PH = 4.0 obtained in the first calibration
const float ph4 = 1280;
// Reference voltage value for PH = 7.0 obtained in the first calibration
const float ph7 = 1690;

// Pin to ORP sensor
#define pinORP 35
// ORP sensor reference voltage
#define voltageORP 5.00
// ORP sensor calibration according to the manufacturer's website.
// It would be necessary to have a good look at it, it CAN BE VARIABLE!!!
#define offsetORP 13
// END Sensors configuration
// END user configuration area

#define seconds 1000UL
#define minutes 60 * seconds

// actionCollectMetrics actives the CollectMetricsJob
#define actionCollectMetrics 1
// actionTransmit actives transmitJob
#define actionTransmit 2
// actionStandby actives standbyJob
#define actionStandby 4
// actionSleepTime actives the sleepJob
#define actionSleep 5

// Type of messages sent by the hub
#define mtypeDeviceConfig 0
#define mtypeAction 1

// Type of actions sent by the hub
// mtypeActionSleep puts the micro controller to sleep
#define mtypeActionSleep 0
// mtypeActionTransmit puts the micro controller to
// Transmit metrics
#define mtypeActionTransmit 1
// mtypeActionStandby puts the micro controller to
// standby until there are customers
#define mtypeActionStandby 2

// Type of messages send to the hub
#define mtypeMetrics 1
#define mtypeTraces 2

// CollectMetricsParam are the parameters
// for collecting metrics
typedef struct {
  // time is the starting time when you start
  // collecting metrics.
  uint64_t time;
} CollectMetricsParam;

CollectMetricsParam collectMetricsParam;

// See pkg.iot.hub
typedef struct {
  uint8_t wakeUpTime;
  uint16_t collectMetricsTime;
  uint8_t buffer;
  uint8_t heartbeatInterval;
  uint8_t heartbeatTimeoutCount;
} Config;

Config configData;

// It is stored outside the Config structure because
// then we can save it after hibernation.
// We don't save the whole config because the memory
// where it is saved is slow and the other data is
// obtained as soon as we connect to the server.
RTC_DATA_ATTR uint8_t wakeUpTime;

// bufferm is the storage of sensor metrics.
#define sizeConfig 384

StaticJsonDocument<sizeConfig> bufferm;
JsonArray tempb;
JsonArray phpb;
JsonArray orpb;

// Communications
// pongTimeout is the time it takes to receive
// the pong after sending the ping.
#define pongTimeout 5 * seconds

#define timeoutDisconnected 5 * minutes

WebSocketsClient ws;
unsigned long lastConnectionTime;

// Sensors
OneWire *ourWire;
DallasTemperature *temp;

// action sets the next action to execute
uint8_t action;

float tempSensor() {
  temp->requestTemperatures();

  return temp->getTempCByIndex(0);
}

float phSensor() {
  // The sensor voltage is read and converted to millivolts.
  float phMillivolts = analogRead(pinPH) / 4095.0 * 3300;

  // Internet-based figures
  float pending = (7.0 - 4.0) / ((ph7 - 1500) / 3.0 - (ph4 - 1500) / 3.0);

  // PH sensor is calibrated
  float offset = 7.0 - pending * (ph7 - 1500) / 3.0;

  return pending * (phMillivolts - 1500) / 3.0 + offset;
}

float orpSensor() {
  // Internet-based figures
  return ((30 * (double)voltageORP * 1000) -
          (75 * analogRead(pinORP) * voltageORP * 1000 / 4095)) /
             75 -
         offsetORP;
}

void setup() {
  // Temp sensor
  pinMode(GPIO_NUM_0, INPUT);
  // Led
  pinMode(GPIO_NUM_4, OUTPUT);

  Serial.begin(115200);

  delay(2 * seconds);

  // Init sensors
  ourWire = new OneWire(GPIO_NUM_0);
  temp = new DallasTemperature(ourWire);

  if (esp_sleep_get_wakeup_cause() == ESP_SLEEP_WAKEUP_UNDEFINED) {
    Serial.println("(setup).The device starts working");
    wakeUpTime = 30;
  } else {
    Serial.printf(
        "(setup).The device wakes up after hibernation."
        "wakeUpTime: %u\n",
        wakeUpTime);
  }

  // Default settings. When connecting to the server
  // the configuration will change to the values configured
  // by the user.
  Serial.println("(setup).Initialising configuration variables");

  configData.collectMetricsTime = 1000;
  configData.buffer = 3;
  configData.heartbeatInterval = 30;
  configData.heartbeatTimeoutCount = 2;

  // Communications
  wsBegin();
  wsSetup();
}

void loop() {
  wsLoop();

  // It uses a state machine pattern.
  // Each iteration observes and determines
  // the next action to be executed.
  switch (action) {
    case actionCollectMetrics:
      collectMetricsJob();
      break;
    case actionTransmit:
      transmissionJob();
      break;
    case actionStandby:
      standbyJob();
      break;
    case actionSleep:
      sleepJob();
      break;
  }
}

// wsBegin initialise web socket to connect to hub
void wsBegin() {
  // Configure receiver for hub messages
  ws.onEvent(hubEvent);

  if (ssl) {
    if (rootCACertificate == nullptr) {
      ws.beginSSL(host, port, URIAPI, SSL_FINGERPRINT_NULL);
    } else {
      ws.beginSslWithCA(host, port, URIAPI, rootCACertificate);
    }
  } else {
    ws.begin(host, port, URIAPI);
  }

  ws.enableHeartbeat(configData.heartbeatInterval * seconds, pongTimeout,
                     configData.heartbeatTimeoutCount);
}

// wsSetup configures web socket to connect
// Uses websocket for communication
void wsSetup() {
  if (action == actionSleep) {
    return;
  }

  Serial.println("(wsSetup).Configures web socket to connect to the hub");

  if (!activeWIFIConnection()) {
    sleep();
    return;
  }

  String token;
  if (!getToken(token)) {
    sleep();
    return;
  }

  char header[200];
  snprintf(header, sizeof(header), "Authorization: Bearer %s\r\nid: %s",
           token.c_str(), DeviceID);

  ws.setExtraHeaders(header);

  // Wait for the hub's commands
  standby();

  Serial.printf("(wsSetup).Configuring web socket (Header: %s)\n", header);
}

// wsLoop polls websocket and manages the retry
// of the connections
void wsLoop() {
  if (action == actionSleep) {
    return;
  }

  ws.loop();

  if (!ws.isConnected()) {
    // The logical way to set up lastConnectionTime would be to
    // initialise millis() on the disconnect event,
    // but the disconnect event is triggered even if there is a
    // 401 (unauthorised) error when the header is sent.
    // This invalidates being able to configure on this event.
    // On the other hand, the connected event works correctly.
    // Therefore, in this event we configure lastConnectionTime=0
    // and so the first time we detect disconnection we configure
    // lastConnectionTime to millis() to perform a possible timeout.
    if (lastConnectionTime == 0) {
      lastConnectionTime = millis();
    }

    // The time cannot be longer than the maximum value
    // of the token expiration set in internal/api/auth.go
    if (millis() - lastConnectionTime >= timeoutDisconnected) {
      // Expired
      Serial.println("(wsLoop).The web socket connection time has expired");
      sleep();
    }
  }
}

// hubEvent process the message received via event hub
void hubEvent(WStype_t type, uint8_t *payload, size_t length) {
  switch (type) {
    case WStype_ERROR:
      Serial.printf("(hubEvent-ERROR).Web socket error. Error: %s", payload);
      ws.disconnect();

      break;
    case WStype_DISCONNECTED:
      Serial.println(
          "(hubEvent).The socket to communicate with the hub has "
          "been disconnected.");

      digitalWrite(GPIO_NUM_4, LOW);
      wsSetup();

      break;
    case WStype_CONNECTED:
      Serial.println(
          "(hubEvent).The socket to communicate with the hub has "
          "been connected.");
      lastConnectionTime = 0;

      break;
    case WStype_TEXT:
      uint8_t typeMessage = ((char *)payload)[0];

      char data[sizeConfig];
      strcpy(data, (char *)(payload + 1));

      Serial.print("(hubEvent). Message received via event hub");
      Serial.printf("(typeMessage: %u, data: %s)\n", typeMessage, data);

      if (typeMessage == mtypeDeviceConfig) {
        config(data);
      }

      if (typeMessage == mtypeAction) {
        if (strlen(data) > 1) {
          Serial.print(
              "(hubEvent-ERROR)."
              "Message received of type action but the data is bad");

          return;
        }
        uint8_t actionh = data[0] - '0';

        switch (actionh) {
          case mtypeActionSleep:
            digitalWrite(GPIO_NUM_4, LOW);
            sleep();
            break;
          case mtypeActionTransmit:
            digitalWrite(GPIO_NUM_4, HIGH);
            transmitMetricsAlready();
            break;
          case mtypeActionStandby:
            digitalWrite(GPIO_NUM_4, LOW);
            standby();
            break;
          default:
            Serial.println(
                "(hubEvent-ERROR). Message received not acknowledged");
            break;
        }
      }

      break;
  }
}

// transmitMetrics requests transmission action
void transmitMetrics() {
  Serial.println("(transmitMetrics).Requesting transmit metrics");

  action = actionTransmit;
}

// transmitMetricsAlready collects an initial set of metrics and performs
// the transmission request transmission action
void transmitMetricsAlready() {
  Serial.println(
      "(transmitMetricsAlready)."
      "Requesting transmit initial metrics");

  initSensorBuffer();

  for (uint8_t i = 0; i < configData.buffer; i++) {
    storeMetrics();
  }

  action = actionTransmit;
}

// transmissionJob is the job that transmits through the socket
// connected to the hub the metrics buffer.
// After transmitting, it continues to collect metrics
// to continue transmitting.
void transmissionJob() {
  Serial.println("(transmissionJob).Transmitting metrics to the hub");
  printMemoryInfo();

  size_t len = measureJson(bufferm);
  char *buffers = new char[len + 2];

  // Defines the message type
  buffers[0] = mtypeMetrics;

  serializeJson(bufferm, buffers + 1, len + 1);

  Serial.printf("(transmissionJob).Metrics to send: '%s'\n", buffers);

  if (!ws.sendTXT(buffers)) {
    Serial.printf("(transmissionJob-ERROR).Error transmitting metrics");
    ws.disconnect();

    return;
  }

  delete[] buffers;

  Serial.println("(transmissionJob).Transmission made");

  collectMetrics();
}

// collectMetrics actives collect metrics action
void collectMetrics() {
  Serial.println("(collectMetrics).Requesting collect metrics");

  collectMetricsParam.time = millis();
  initSensorBuffer();

  action = actionCollectMetrics;
}

// collectMetricsJob collects the metrics every "configData.collectMetricsTime"
// set in the configuration and when the preset time is reached,
// the metrics are transmitted to the hub.
void collectMetricsJob() {
  int64_t timeElapsedMetricsSec =
      (millis() - collectMetricsParam.time) / seconds;

  if (timeElapsedMetricsSec >= configData.buffer) {
    Serial.printf(
        "(collectMetricsJob).Buffer filled. ("
        "timeElapsedMetricsSec: %lld, "
        "config.buffer: %u)\n",
        timeElapsedMetricsSec, configData.buffer);

    transmitMetrics();

    return;
  }

  delay(configData.collectMetricsTime);
  storeMetrics();

  action = actionCollectMetrics;
}

// standby actives standby job
void standby() {
  Serial.println("(standby).Requesting standby action");

  action = actionStandby;
}

// standbyJob puts the device in standby mode,
// waiting for events from the hub to come in.
void standbyJob() {
  delay(200);  // to save energy
  action = actionStandby;
}

// sleep actives sleep job
void sleep() {
  Serial.println("(sleep).Requesting sleep action");

  action = actionSleep;
}

// sleepJo frees up resources and goes into hibernation mode
void sleepJob() {
  Serial.println("(sleepJob).Sleeping device...");

  ws.disconnect();

  ourWire->reset();
  delete ourWire;
  delete temp;

  bufferm.clear();

  Serial.flush();

  // config.wakeUpTime is in minutes. Convert to microseconds
  esp_sleep_enable_timer_wakeup(wakeUpTime * 60ULL * 1000000ULL);
  esp_deep_sleep_start();
}

// config updates the device settings
void config(const char *data) {
  Serial.println("(config).Updating configuration");

  StaticJsonDocument<128> doc;
  DeserializationError err = deserializeJson(doc, data);

  if (err.code() != DeserializationError::Code::Ok) {
    Serial.printf(
        "(config-ERROR).Error deserialization configuration. Error: %s\n",
        err.c_str());

    return;
  }

  configData.buffer = doc["buffer"];
  configData.collectMetricsTime = doc["cmt"];
  configData.heartbeatInterval = doc["hbi"];
  configData.heartbeatTimeoutCount = doc["hbtc"];
  wakeUpTime = doc["wut"];

  ws.enableHeartbeat(configData.heartbeatInterval * seconds, pongTimeout,
                     configData.heartbeatTimeoutCount);

  Serial.printf("(config).The configuration is updated via hub (");
  Serial.printf("heartbeatInterval: %u, ", configData.heartbeatInterval);
  Serial.printf("heartbeatTimeCount: %u, ", configData.heartbeatTimeoutCount);
  Serial.printf("Buffer: %u, ", configData.buffer);
  Serial.printf("CollectMetricsTime: %d, ", configData.collectMetricsTime);
  Serial.printf("WakeUpTime: %u)\n", wakeUpTime);
}

void initSensorBuffer() {
  bufferm.clear();

  tempb = bufferm.createNestedArray("temp");
  phpb = bufferm.createNestedArray("ph");
  orpb = bufferm.createNestedArray("orp");
}

// storeMetrics store metrics of the sensor in the buffer
void storeMetrics() {
  Serial.println("(storeMetrics).collect metrics");

  char val[10];

  sprintf(val, "%.1f", tempSensor());
  rtrim(val);
  tempb.add(val);

  sprintf(val, "%.1f", phSensor());
  rtrim(val);
  phpb.add(val);

  sprintf(val, "%.1f", orpSensor());
  rtrim(val);
  orpb.add(val);
}

void rtrim(char *str) {
  if (str == NULL || *str == '\0') return;

  char *ptr = str + strlen(str) - 1;
  while (ptr >= str && *ptr == ' ') *ptr-- = '\0';
}

// getToken gets security token
bool getToken(String &result) {
  uint16_t maxRetry = 5 * 60;  // Minutes
  uint16_t retry;

  Serial.println("(getToken).Getting security token");

  char uri[50];
  strcpy(uri, URIToken);
  strcat(uri, clientID);

  HTTPClient http;
  if (!httpBegin(&http, uri)) {
    Serial.println("(getToken-ERROR). Error setting http to get token");

    return false;
  }

  http.addHeader("Content-Type", "text/plain");

  int httpCode = 0;

  while (++retry < maxRetry) {
    httpCode = http.GET();

    if (httpCode > 0) {
      break;
    }

    if (retry % 10 == 0) {
      Serial.printf(
          "\n(getToken-ERROR).Unable to open the connection to %s (Code: %d)\n",
          host, httpCode);
    }

    Serial.print(".");
    delay(1 * seconds);
  }

  Serial.println();

  result = http.getString();

  http.end();

  if (httpCode < 0) {
    Serial.printf(
        "(getToken-ERROR).Unable to open the connection to %s (Code: %d)\n",
        host, httpCode);

    return false;
  }

  if (httpCode != HTTP_CODE_OK) {
    Serial.printf("(getToken).Cannot get the token. ");
    Serial.printf("(URI: %s, Code: %s, Result: %s)\n", URIToken, httpCode,
                  result);

    return false;
  }

  return true;
}

// activeWIFIConnection checks for WIFI connection
// and if not, try again.
bool activeWIFIConnection() {
  if (WiFi.status() != WL_CONNECTED) {
    Serial.println(
        "(activeWIFIConnection).Connection has not been established "
        "or has been lost. Re-attempted to connect...");

    WiFi.begin(ssid, password);

    Serial.printf("(activeWIFIConnection).Connecting to WIFI (ID: %s)", ssid,
                  password);

    uint8_t retry;
    while (WiFi.status() != WL_CONNECTED) {
      if (++retry == 2 * 60) {  // Minutes
        Serial.printf(
            "\n(activeWIFIConnection-ERROR). Error connecting to WIFI: %u\n",
            WiFi.status());

        return false;
      }

      Serial.print(".");
      delay(1 * seconds);
    }

    Serial.println();
    Serial.print("(activeWIFIConnection).Connected to WIFI: ");
    Serial.println(WiFi.localIP());

    return true;
  }
}

// httpBegin configures a http service via uri param
bool httpBegin(HTTPClient *http, const char *uri) {
  if (ssl) {
    char url[sizeof(host) + 50];
    snprintf(url, sizeof(url), "https://%s:%u%s", host, port, uri);

    Serial.printf("(httpConnect).Connecting to %s\n", url);

    return http->begin(url, rootCACertificate);
  }

  Serial.printf("(httpConnect).Connecting to http://%s:%u%s\n", host, port,
                uri);

  return http->begin(host, port, uri);
}

void printMemoryInfo() {
  Serial.print("Total memory: ");
  Serial.print(ESP.getHeapSize());
  Serial.print(" bytes. Free memory: ");
  Serial.print(ESP.getFreeHeap());
  Serial.println(" bytes.");
}