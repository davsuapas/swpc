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

// Time
#define seconds 1000UL
#define minutes 60 * seconds

// BEGIN user configuration area
// rootCACertificate = nullptr y ssl = true
// it's used for self-certification
const char *rootCACertificate = nullptr;

#define ssl true
#define host ""
#define port 443

#define URIAPI "/api/device/ws"
#define URIToken "/auth/token/"

// clientID define the ID to connect to the server
#define clientID ""

// WIFI definition
const char *ssid = "";
const char *password = "";

// Device ID
#define DeviceID ""

// BEGIN Sensors configuration
// Pin to temp sensor
#define pinTemp 0

// Pin to led indicator
#define pinLed 4

// Pin to PH sensor
#define pinPH 34
// Reference voltage value for PH = 4.0 obtained in the first calibration
const float ph4 = 1280;
// Reference voltage value for PH = 7.0 obtained in the first calibration
const float ph7 = 1690;

// Pin to ORP sensor
#define pinORP 33
// Analog bits mask
#define adcRes 4098
// Votage 5v
#define voltRef 5000

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
typedef struct
{
  // time is the starting time when you start
  // collecting metrics.
  uint64_t time;
} CollectMetricsParam;

CollectMetricsParam collectMetricsParam;

// See pkg.iot.hub
typedef struct
{
  uint8_t wakeUpTime;
  uint16_t collectMetricsTime;
  uint8_t buffer;
  uint8_t heartbeatInterval;
  uint8_t heartbeatTimeoutCount;
  bool calibratingORP; 
  float targetORP;
  float calibrationORP;
  bool calibratingPH; 
  float targetPH;
  float calibrationPH;
  unsigned long stabilizationTime;
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

// Calibration variables
typedef struct
{
  bool stabilizationDone;
  float incrementalAverage;
  int numReadings;
} Calibration;

Calibration calibrationORP;
Calibration calibrationPH;

unsigned long calibrationStartTime;

// ORP sensor
typedef struct
{
  float incrementalAverage;
  int numReadings;
} ORP;

ORP orp;

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

float tempSensor()
{
  temp->requestTemperatures();

  return temp->getTempCByIndex(0);
}

// Calibrates the sensor by adjusting the calibration
// value to match the target level.
// Ensures stabilization by averaging readings over a
// defined period.
// Returns the calibration value,
// which can be displayed on the screen.
float calibrate(float measure, float target, Calibration &calibStore) {
  // Calculate current calibration to match target
  float calibration = target - measure;
   
  calibStore.numReadings++;

   // Calculate incremental avg
   calibStore.incrementalAverage +=
    (calibration - calibStore.incrementalAverage) /
    calibStore.numReadings;

  Serial.printf(
    "Measure calibrates: %.2f = (measure: %.2f + calibration: %.2f); "
    "target: %.2f, calibrationAvg: %.2f\n",
    measure + calibration,
    measure,
    calibration,
    target,
    calibStore.incrementalAverage);
    
   // Calculate time elapsed in stabilization period
  unsigned long elapsedTime = millis() - calibrationStartTime;
    
   // Check if we've reached the stabilization time
  if (elapsedTime >= configData.stabilizationTime) {
    calibration = calibStore.incrementalAverage;

    // Mark stabilization as complete
    calibStore.stabilizationDone = true;
    
    // Final debug output
    Serial.printf("Calibration stabilization COMPLETE "
      "after %lu ms with %d readings\n", 
      elapsedTime,
      calibStore.numReadings);
    Serial.printf("Final calibration: %.2f\n", calibration);
  }
  
  return calibration;
}

void initStabilizePH() {
  calibrationPH.stabilizationDone = false;
  calibrationPH.incrementalAverage = 0;
  calibrationPH.numReadings = 0;
  calibrationStartTime = millis();
}


float phSensor()
{
  // The sensor voltage is read and converted to millivolts.
  float phMillivolts = analogRead(pinPH) / 4095.0 * 3300;

  // Internet-based figures
  float pending = (7.0 - 4.0) / ((ph7 - 1500) / 3.0 - (ph4 - 1500) / 3.0);

  // PH sensor is calibrated
  float offset = 7.0 - pending * (ph7 - 1500) / 3.0;

  float valor = pending * (phMillivolts - 1500) / 3.0 + offset;

  if (configData.calibratingPH) {
    // If calibration is already stabilized,
    // just return the current calibration value
    if (calibrationPH.stabilizationDone) {
      return configData.calibrationPH;
    }

    configData.calibrationPH = 
      calibrate(valor, configData.targetPH, calibrationPH);

    return configData.calibrationPH;
  } else {
    return valor + configData.calibrationPH;
  }
}

void initStabilizeORP() {
  calibrationORP.stabilizationDone = false;
  calibrationORP.incrementalAverage = 0;
  calibrationORP.numReadings = 0;
  calibrationStartTime = millis();
}

void initORP() {
  orp.incrementalAverage = 0;
  orp.numReadings = 0;
}

float orpSensor()
{
  float adcVoltage =
   ((unsigned long)analogRead(pinORP) * voltRef + adcRes / 2) / adcRes;

  if (configData.calibratingORP) {
    // If calibration is already stabilized,
    // just return the current calibration value
    if (calibrationORP.stabilizationDone) {
      return configData.calibrationORP;
    }

    configData.calibrationORP = 
      calibrate(adcVoltage, configData.targetORP, calibrationORP);

    return configData.calibrationORP;
  } else {
    float orpValue = adcVoltage + configData.calibrationORP;

    orp.numReadings++;
    orp.incrementalAverage +=
      (orpValue - orp.incrementalAverage) / orp.numReadings;    

    return orp.incrementalAverage;
  }
}

void setup()
{
  // Temp sensor
  pinMode(pinTemp, INPUT);
  // Led
  pinMode(pinLed, OUTPUT);

  Serial.begin(115200);

  delay(2 * seconds);

  // Init sensors
  ourWire = new OneWire(pinTemp);
  temp = new DallasTemperature(ourWire);

  if (esp_sleep_get_wakeup_cause() == ESP_SLEEP_WAKEUP_UNDEFINED)
  {
    Serial.println("(setup).The device starts working");
    wakeUpTime = 30;
  }
  else
  {
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
  configData.calibratingORP = false;
  configData.calibrationORP = 1440;
  configData.stabilizationTime = 30 * seconds;

  initStabilizeORP();
  initStabilizePH();
  initORP();

  // Communications
  wsBegin();
  wsSetup();
}

void loop()
{
  wsLoop();

  // It uses a state machine pattern.
  // Each iteration observes and determines
  // the next action to be executed.
  switch (action)
  {
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
void wsBegin()
{
  // Configure receiver for hub messages
  ws.onEvent(hubEvent);

  if (ssl)
  {
    if (rootCACertificate == nullptr)
    {
      ws.beginSSL(host, port, URIAPI, SSL_FINGERPRINT_NULL);
    }
    else
    {
      ws.beginSslWithCA(host, port, URIAPI, rootCACertificate);
    }
  }     

  else
  {
    ws.begin(host, port, URIAPI);
  }

  ws.enableHeartbeat(
    configData.heartbeatInterval * seconds,
    pongTimeout,
    configData.heartbeatTimeoutCount);
}

// wsSetup configures web socket to connect
// Uses websocket for communication
void wsSetup()
{
  if (action == actionSleep)
  {
    return;
  }

  Serial.println("(wsSetup).Configures web socket to connect to the hub");

  if (!activeWIFIConnection())
  {
    sleep();
    return;
  }

  String token;
  if (!getToken(token))
  {
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
void wsLoop()
{
  if (action == actionSleep)
  {
    return;
  }

  ws.loop();

  if (!ws.isConnected())
  {
    // The logical way to set up lastConnectionTime would be to
    // initialise millis() on the disconnect event,
    // but the disconnect event is triggered even if there is a
    // 401 (unauthorised) error when the header is sent.
    // This invalidates being able to configure on this event.
    // On the other hand, the connected event works correctly.
    // Therefore, in this event we configure lastConnectionTime=0
    // and so the first time we detect disconnection we configure
    // lastConnectionTime to millis() to perform a possible timeout.
    if (lastConnectionTime == 0)
    {
      lastConnectionTime = millis();
    }

    // The time cannot be longer than the maximum value
    // of the token expiration set in internal/api/auth.go
    if (millis() - lastConnectionTime >= timeoutDisconnected)
    {
      // Expired
      Serial.println("(wsLoop).The web socket connection time has expired");
      sleep();
    }
  }
}

// hubEvent process the message received via event hub
void hubEvent(WStype_t type, uint8_t *payload, size_t length)
{
  switch (type)
  {
  case WStype_ERROR:
    Serial.printf("(hubEvent-ERROR).Web socket error. Error: %s", payload);
    ws.disconnect();

    break;
  case WStype_DISCONNECTED:
    Serial.println(
        "(hubEvent).The socket to communicate with the hub has "
        "been disconnected.");

    digitalWrite(pinLed, LOW);
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

    if (typeMessage == mtypeDeviceConfig)
    {
      config(data);
    }

    if (typeMessage == mtypeAction)
    {
      if (strlen(data) > 1)
      {
        Serial.print(
            "(hubEvent-ERROR)."
            "Message received of type action but the data is bad");

        return;
      }
      uint8_t actionh = data[0] - '0';

      switch (actionh)
      {
      case mtypeActionSleep:
        digitalWrite(pinLed, LOW);
        sleep();
        break;
      case mtypeActionTransmit:
        digitalWrite(pinLed, HIGH);
        transmitMetricsAlready();
        break;
      case mtypeActionStandby:
        digitalWrite(pinLed, LOW);
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
void transmitMetrics()
{
  Serial.println("(transmitMetrics).Requesting transmit metrics");

  action = actionTransmit;
}

// transmitMetricsAlready collects an initial set of metrics and performs
// the transmission request transmission action
void transmitMetricsAlready()
{
  Serial.println(
      "(transmitMetricsAlready)."
      "Requesting transmit initial metrics");

  initSensorBuffer();
  initORP();

  for (uint8_t i = 0; i < configData.buffer; i++)
  {
    storeMetrics();
  }

  action = actionTransmit;
}

// transmissionJob is the job that transmits through the socket
// connected to the hub the metrics buffer.
// After transmitting, it continues to collect metrics
// to continue transmitting.
void transmissionJob()
{
  Serial.println("(transmissionJob).Transmitting metrics to the hub");
  printMemoryInfo();

  size_t len = measureJson(bufferm);
  char *buffers = new char[len + 2];

  // Defines the message type
  buffers[0] = mtypeMetrics;

  serializeJson(bufferm, buffers + 1, len + 1);

  Serial.printf("(transmissionJob).Metrics to send: '%s'\n", buffers);

  if (!ws.sendTXT(buffers))
  {
    Serial.printf("(transmissionJob-ERROR).Error transmitting metrics");
    ws.disconnect();

    return;
  }

  delete[] buffers;

  Serial.println("(transmissionJob).Transmission made");

  collectMetrics();
}

// collectMetrics actives collect metrics action
void collectMetrics()
{
  Serial.println("(collectMetrics).Requesting collect metrics");

  collectMetricsParam.time = millis();
  initSensorBuffer();

  action = actionCollectMetrics;
}

// collectMetricsJob collects the metrics every "configData.collectMetricsTime"
// set in the configuration and when the preset time is reached,
// the metrics are transmitted to the hub.
void collectMetricsJob()
{
  int64_t timeElapsedMetricsSec =
      (millis() - collectMetricsParam.time) / seconds;

  if (timeElapsedMetricsSec >= configData.buffer)
  {
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
void standby()
{
  Serial.println("(standby).Requesting standby action");

  action = actionStandby;
}

// standbyJob puts the device in standby mode,
// waiting for events from the hub to come in.
void standbyJob()
{
  delay(200); // to save energy
  action = actionStandby;
}

// sleep actives sleep job
void sleep()
{
  Serial.println("(sleep).Requesting sleep action");

  action = actionSleep;
}

// sleepJo frees up resources and goes into hibernation mode
void sleepJob()
{
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
void config(const char *data)
{
  Serial.println("(config).Updating configuration");

  StaticJsonDocument<128> doc;
  DeserializationError err = deserializeJson(doc, data);

  if (err.code() != DeserializationError::Code::Ok)
  {
    Serial.printf(
        "(config-ERROR).Error deserialization configuration. Error: %s\n",
        err.c_str());

    return;
  }

  configData.buffer = doc["buffer"];
  configData.collectMetricsTime = doc["cmt"];
  configData.heartbeatInterval = doc["hbi"];
  configData.heartbeatTimeoutCount = doc["hbtc"];
  configData.calibratingORP = doc["cgorp"];
  configData.targetORP = doc["torp"];
  configData.calibrationORP = doc["corp"];
  configData.calibratingPH = doc["cgph"];
  configData.targetPH = doc["tph"];
  configData.calibrationPH = doc["cph"];
  configData.stabilizationTime = (int)doc["st"] * seconds;
  wakeUpTime = doc["wut"];

  initStabilizeORP();
  initStabilizePH();
  initORP();

  ws.enableHeartbeat(
    configData.heartbeatInterval * seconds,
    pongTimeout,
    configData.heartbeatTimeoutCount);

  Serial.printf("(config).The configuration is updated via hub (");
  Serial.printf("heartbeatInterval: %u, ", configData.heartbeatInterval);
  Serial.printf("heartbeatTimeCount: %u, ", configData.heartbeatTimeoutCount);
  Serial.printf("Buffer: %u, ", configData.buffer);
  Serial.printf("CollectMetricsTime: %d, ", configData.collectMetricsTime);
  Serial.printf("WakeUpTime: %u, ", wakeUpTime);
  Serial.printf("calibrationORP: %f, ", configData.calibrationORP);
  Serial.printf("calibrationPH: %f, ", configData.calibrationPH);
  Serial.printf(
    "calibratingORP: %s, ",
    configData.calibratingORP ? "true" : "false");
  Serial.printf("targetORP: %f, ", configData.targetORP);
  Serial.printf(
    "calibratingPH: %s, ",
    configData.calibratingPH ? "true" : "false");
  Serial.printf("targetPH: %f, ", configData.targetPH);
  Serial.printf("stabilizationTime: %lu)", configData.stabilizationTime);
}

void initSensorBuffer()
{
  bufferm.clear();

  tempb = bufferm.createNestedArray("temp");
  phpb = bufferm.createNestedArray("ph");
  orpb = bufferm.createNestedArray("orp");
}

// storeMetrics store metrics of the sensor in the buffer
void storeMetrics()
{
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

void rtrim(char *str)
{
  if (str == NULL || *str == '\0')
    return;

  char *ptr = str + strlen(str) - 1;
  while (ptr >= str && *ptr == ' ')
    *ptr-- = '\0';
}

// getToken gets security token
bool getToken(String &result)
{
  uint16_t maxRetry = 10 * 60; // Minutes
  uint16_t retry;

  Serial.println("(getToken).Getting security token");

  char uri[50];
  strcpy(uri, URIToken);
  strcat(uri, clientID);

  HTTPClient http;
  if (!httpBegin(&http, uri))
  {
    Serial.println("(getToken-ERROR). Error setting http to get token");

    return false;
  }

  http.addHeader("Content-Type", "text/plain");

  int httpCode = 0;

  while (++retry < maxRetry)
  {
    httpCode = http.GET();

    if (httpCode == HTTP_CODE_OK)
    {
      break;
    }

    if (retry % 30 == 0)
    {
      Serial.printf(
          "\n(getToken-ERROR).Unable to open the connection to %s (Code: %d)\n",
          host, httpCode);
    }

    Serial.print(".");
    delay(1 * seconds);
  }

  bool ok = false;

  if (httpCode == HTTP_CODE_OK)
  {
    result = http.getString();
    ok = true;
  }
  else
  {
    Serial.printf("\n(getToken).Cannot get the token. ");
    Serial.printf("(Code: %d)\n", httpCode);
  }

  http.end();

  return ok;
}

// activeWIFIConnection checks for WIFI connection
// and if not, try again.
bool activeWIFIConnection()
{
  if (WiFi.status() != WL_CONNECTED)
  {
    Serial.println(
        "(activeWIFIConnection).Connection has not been established "
        "or has been lost. Re-attempted to connect...");

    WiFi.begin(ssid, password);

    Serial.printf("(activeWIFIConnection).Connecting to WIFI (ID: %s)", ssid);

    uint8_t retry;
    while (WiFi.status() != WL_CONNECTED)
    {
      if (++retry == 10 * 60)
      { // Minutes
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
bool httpBegin(HTTPClient *http, const char *uri)
{
  if (ssl)
  {
    char url[sizeof(host) + 50];
    snprintf(url, sizeof(url), "https://%s:%u%s", host, port, uri);

    Serial.printf("(httpConnect).Connecting to %s\n", url);

    return http->begin(url, rootCACertificate);
  }

  Serial.printf(
    "(httpConnect).Connecting to http://%s:%u%s\n",
    host,
    port,
    uri);

  return http->begin(host, port, uri);
}

void printMemoryInfo()
{
  Serial.print("Total memory: ");
  Serial.print(ESP.getHeapSize());
  Serial.print(" bytes. Free memory: ");
  Serial.print(ESP.getFreeHeap());
  Serial.println(" bytes.");
}
