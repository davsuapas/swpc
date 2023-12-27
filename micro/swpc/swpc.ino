/*
 *   Copyright (c) 2022 CARISA
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
#include <WiFi.h>
#include "tiny_websockets/message.hpp"
#include "tiny_websockets/client.hpp"
#include <ArduinoJson.h>
#include <OneWire.h>                
#include <DallasTemperature.h>

using namespace websockets;

// BRGIN configuration area
const char *rootCACertificate = \
  "-----BEGIN CERTIFICATE-----\n" \
  "MIIDZzCCAk8CFCtAzzSUwZs5ps9NJdV/xWP1xCBSMA0GCSqGSIb3DQEBCwUAMHA\n" \
  "CzAJBgNVBAYTAkVTMQ8wDQYDVQQIDAZNYWRyaWQxDzANBgNVBAcMBk1hZHJpZDE\n" \
  "MA8GA1UECgwIZWxpcGNlcm8xLDAqBgNVBAMMI3N3cGMuZXUtd2VzdC0xLmVsYXN\n" \ 
  "aWNiZWFuc3RhbGsuY29tMB4XDTIzMTAyMzE1MDIyOFoXDTI0MTAyMjE1MDIyOFo\n" \
  "cDELMAkGA1UEBhMCRVMxDzANBgNVBAgMBk1hZHJpZDEPMA0GA1UEBwwGTWFkcml\n" \
  "MREwDwYDVQQKDAhlbGlwY2VybzEsMCoGA1UEAwwjc3dwYy5ldS13ZXN0LTEuZWx\n" \
  "c3RpY2JlYW5zdGFsay5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoI\n" \
  "AQDJV8AvmkqnHOV9ZVG3ZxE/StTOKiPa6JtSCaz84BMIGs+cvezSz6NfAl/yJLU\n" \
  "92hwQoTtzJEbF6OmlaW99IIx6qUvpHfMZ4gOpyvwWzWExV365fkr9hz6GVWI1HM\n" \
  "RNeorLde/oklBe5bJN76QVdZ9OITyv0YJRi918GjOJxdfsG0ZZSWnLTG1A1z2oD\n" \
  "NDZi3SSN5rudKT10dnDe4rZU2KdqQi2JvlG2+2fsnfwDuK2Zj2XOvkwx823RqeY\n" \
  "U0BjPoOHGzrH5kawhE1BWJWzfrb/G7Ra6JNRI1rHls12PGcnBsn4FF65A094nGB\n" \
  "i2KNkThB5GcGvCDtllDqTHClAgMBAAEwDQYJKoZIhvcNAQELBQADggEBACcZViq\n" \
  "xIvezg2UQ3LuEm+xpVS0781KvtOlQ8i4A2hvbOmvD1p+FhS4fXonSmDtd1pS/j7\n" \
  "CtvoDYcJajOPlpfFOkNsQNuXmxGZ+yoGAgzhA2yOJfUjnc55PCphXUEaXfBR19c\n" \
  "niuArYvcA/WtoQO8YDY8H4a25wTSMNXml2xoMxIms0dqT73GMsANBscOmm9k6su\n" \
  "WEdCee0DOzbU9bZNtiioUldDfWZfUJiRzRUSSuk4eirtAQSc0qSp0Q1qVaeL5fS\n" \
  "3OTWa7LyJ7VVC/KF0GJ4CUQ2LfUsYnnOYrwwQR52sxiLOc2yIItR2Ic8XSe+9Fr\n" \
  "jZQs53bE1qYvC6k=\n" \
  "-----END CERTIFICATE-----\n";

#define URL "ws://192.168.1.135:5000"  

#define URLAPI URL "/micro/api"

// clientID define the ID to connect to the server
#define clientID ""
#define URIToken URL "/auth/token/" clientID

// WIFI definition
const char *ssid = "";
const char *password = "";

// retryTimeSeconds defines the the retry time to request server
#define retryTimeSeconds 10

// Size of buffer
#define sizeTempBuffer 50
#define sizePHBuffer 50
#define sizeORPBuffer 70

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
// END configuration area

// HTTP Method
#define HTTPGet "GET"
#define HTTPPost "POST"

// sleepServer puts the micro controller to sleep.
#define sleepServer 0
// transmitServer puts the micro controller to transmit metrics
#define transmitServer 1
// checkTransmissionServer puts the micro controller
// to check transmission status
#define checkTransmissionServer 2

// Actions of the states machine
// sleep puts the micro to sleep
#define sleep 0
// transmit puts the micro to send metrics
#define transmit 1
// checkNextAction queries the next action to execute
#define checkNextAction 2
// wakeupCheckNextAction performs the action query next action
#define wakeupCheckNextAction 3
// wakeupRetry performs the action retry
#define wakeupRetry 4
// wakeupCollectBuffer performs the action collecting buffer
#define wakeupCollectBuffer 5

// See internal.micro.Behavior
typedef struct {
  uint8_t checkTransTime;
  uint16_t collectMetricsTime;
  uint8_t buffer;
  uint8_t action;
} Config;

// It is stored outside the Config structure because
// then we can save it after hibernation.
// We don't save the whole config because the memory
// where it is saved is slow and the other data is
// obtained as soon as we connect to the server.
RTC_DATA_ATTR uint8_t wakeUpTime;

// Retry management
typedef struct {
  // action is the next action to be taken when time is up
  uint8_t action;
  // counter i el retry number
  uint8_t counter;
  // max is the max number of retry
  uint8_t max;
} Retry;

// Define Buffer struct. 
// This buffer is stored for a time set in the configuration
// coming from the server. This configuration is done by the user
typedef struct {
  char bufferTemp[50];
  char bufferPH[50];
  char bufferORP[70];
  // When you start collecting metrics in the buffer,
  // you record the time when you started at timeStartChrono
  int64_t timeStartChrono;
  // nextAction defines after collecting the metrics
  // what the next action will be.
  uint8_t nextAction;
} Buffer;

Config config;
Retry retry;
Buffer buffer;
// action sets the next action to execute
uint8_t action;
// securToken sets the security token
String securToken;
// timeStartChrono sets the start of the stopwatch
int64_t timeStartChrono;
// lastError the last error that occurred is set
String lastError;

// Communications
// To improve performance, they are kept global,
// as they are constantly used
WebsocketsClient *ws = NULL;

OneWire *ourWire;                    
DallasTemperature *temp;

float tempSensor() {
  temp->requestTemperatures();
  return temp->getTempCByIndex(0);                        
}

float phSensor() {
  // The sensor voltage is read and converted to millivolts.
  float phMillivolts = analogRead(pinPH) / 4095.0 * 3300;

  // Internet-based figures
  float pending = (7.0 - 4.0) / (( ph7 - 1500) / 3.0 - (ph4 - 1500) / 3.0);  

  // PH sensor is calibrated
  float offset = 7.0 - pending * (ph7 - 1500) / 3.0;                         

  return pending * (phMillivolts - 1500) / 3.0 + offset;
}

float orpSensor() {
  // Internet-based figures
  return (
    (30 * (double)voltageORP * 1000) -
    (75 * analogRead(pinORP) * voltageORP*1000/4095)
  ) / 75 - offsetORP;   
}

void setup() {
  // Temp sensor
  pinMode(GPIO_NUM_0, INPUT);         

  Serial.begin(9600);

  delay(5000); // Only for debugging on the serial monitor 

  // Init sensors
  ourWire = new OneWire(GPIO_NUM_0);
  temp = new DallasTemperature(ourWire);  

  if (esp_sleep_get_wakeup_cause() == ESP_SLEEP_WAKEUP_UNDEFINED) {
    Serial.println("The micro starts working");
    wakeUpTime = 30;
  } else {
    Serial.printf(
      "The micro wakes up after hibernation."
      "wakeUpTime: %u\n",
      wakeUpTime
    );
  }

  checkWIFIConnection();

  ws = new WebsocketsClient();
  
  // Default settings. When connecting to the server
  // the configuration will change to the values configured
  // by the user.
  Serial.println("Initialising configuration variables");
  config.checkTransTime = 5;
  config.collectMetricsTime = 800;
  config.buffer = 3;

  iniRetry();
  //Collect a sample before sending
  requestCollectsBuffer(checkNextAction);
}

void loop() {
  // It uses a state machine pattern.
  // Each iteration observes and determines
  // the next state to be executed.
  switch (action) {
    case sleep:
      doSleep();
      break;
    case checkNextAction:
      queryNextAction(checkNextAction);
      break;
    case transmit:
      if (doTrasnmisssion()) {
        requestCollectsBuffer(action);
      }
      break;
    case wakeupCollectBuffer:
      wakeupToCollectBuffer();
      break;
    case wakeupRetry:
      wakeupToRetry();
      break;
    case wakeupCheckNextAction:
      wakeupToCheckNextAction();
      break;
  }
}

// doSleep frees up resources and goes into hibernation mode
void doSleep() {
  Serial.println("Sleeping ...");
  // Everything that has to do with connections is released
  if (wifics) {
    wifics->stop();
    delete wifics;
    wifics = NULL;
  }

  delete http;
  http = NULL;

  ourWire->reset();
  delete ourWire;
  delete temp;

  Serial.flush();

  //config.wakeUpTime is in minutes. Convert to microseconds
  esp_sleep_enable_timer_wakeup(wakeUpTime * 60ULL * 1000000ULL);
  esp_deep_sleep_start();
}

// queryNextAction requests the next action to be executed
// and in case of error indicates by parameter which action
// to execute after the retry
void queryNextAction(uint8_t actionRetry) {
  Serial.println("Execute queryNextAction");

  doRequest("GET", "/action", "", actionRetry, 30);
}

// doTrasnmisssion sends the metrics to the server
int doTrasnmisssion() {
  Serial.println("Execute doTrasnmisssion");
  printMemoryInfo();

  // 2 is for the semicolon
  char buffert[
    strlen(buffer.bufferTemp) +
    strlen(buffer.bufferPH) +
    strlen(buffer.bufferORP) + 2
  ];

  strcpy(buffert, buffer.bufferTemp);
  strcat(buffert, ";");
  strcat(buffert, buffer.bufferPH);
  strcat(buffert, ";");
  strcat(buffert, buffer.bufferORP);

  Serial.printf("Buffer to send: '%s'\n", buffert);

  return doRequest("POST", "/download", buffert, transmit, 60);
}

// wakeupToCollectBuffer collects metrics each config.collectMetricsTime
// if it reaches the maximum buffer time,
// it performs the action indicated in buffer.nextAction
void wakeupToCollectBuffer() {
  int64_t timeElapsedMetricsSec =
    (millis() - buffer.timeStartChrono) / 1000;

  if (timeElapsedMetricsSec > config.buffer) {
    Serial.printf(
      "Buffer completed. wakeupToCollectBuffer.timeElapsedMetricsSec: %lld, "
      "config.buffer: %u, "
      "nextAction: %s\n",
      timeElapsedMetricsSec,
      config.buffer,
      printAction(buffer.nextAction));

    action = buffer.nextAction;

    return;
  }

  if (timeElapsedMiliSec() >= config.collectMetricsTime) {
    Serial.printf(
    "Time to collect metrics. timeElapsedMiliSec: %lld, "
    "config.collectMetricsTime: %u\n",
    timeElapsedMiliSec(),
    config.collectMetricsTime);

    startChrono();
    collectMetrics();
  }
}

// wakeupToRetry checks the retry time. If it is met,
// the action is reattempted,
// and if the maximum time is met, the action is slept.
void wakeupToRetry() {
  if (timeElapsedSec() >= retryTimeSeconds) {
    Serial.printf(
    "Time to retry. timeElapsedMiliSec: %lld \n",
    timeElapsedSec());

    if (retry.counter == retry.max) {
      iniRetry();

      Serial.printf(
        "The maximum retry time has been exceeded: retry-max: %d\n",
        retry.max);

      action = sleep;

      return;
    }

    retry.counter++;  

    Serial.printf(
      "wakeupToRetry.counter: %d, action: %s\n",
      retry.counter,
      printAction(retry.action));

    action = retry.action;
  }
}

// wakeupToCheckNextAction checks if he has to wake up to get next action
void wakeupToCheckNextAction() {
  if (timeElapsedSec() > config.checkTransTime) {
    Serial.printf(
      "wakeupToCheckNextAction.timeElapsedSec: %lld\n",
      timeElapsedSec());

    action = checkNextAction;
  }
}

// setNextAction calculates the next actions
void setNextAction(uint8_t actionServer) {
  switch (actionServer) {
    case sleepServer:
        action = sleep;
        break;
    case transmitServer:
        action = transmit;
        break;
    case checkTransmissionServer:
        action = wakeupCheckNextAction;
        startChrono();
        break;
  }

  Serial.printf("nextAction: '%s'\n", printAction(action));
}

// collectBuffer initialises the buffer and
// it requests to perform the action of collecting
// and it set next action by nextAction param
void requestCollectsBuffer(uint8_t nextAction) {
  Serial.printf(
    "requestCollectBuffer.nextAction: %s\n",
    printAction(nextAction));

  buffer.timeStartChrono = millis();
  buffer.nextAction = nextAction;
  strcpy(buffer.bufferTemp, "");
  strcpy(buffer.bufferPH, "");
  strcpy(buffer.bufferORP, "");

  action = wakeupCollectBuffer;

  startChrono();
}

// collectMetrics collects metrics of the sensor
void collectMetrics() {
  Serial.println("Execute collectMetrics");

  concatSensorMetrics(
    "temp", buffer.bufferTemp, sizeTempBuffer, tempSensor(), 4, 3, 1);
  concatSensorMetrics(
    "ph", buffer.bufferPH, sizePHBuffer, phSensor(), 4, 3, 1);
  concatSensorMetrics(
    "orp", buffer.bufferORP, sizeORPBuffer, orpSensor(), 6, 4, 1);
}

void concatSensorMetrics(
  const char *name,
  char *sensor,
  uint8_t bufferSize,
  float value,
  uint8_t bsize,
  signed int width,
  unsigned int prec) {

  char sensorf[bsize];
  dtostrf(value, width, prec, sensorf);

  // Elimino los espacios
  char *ptr = sensorf;
  char *ptr2 = sensorf;

  while (*ptr != '\0') {
      if (*ptr != ' ') {
          *ptr2 = *ptr;
          ptr2++;
      }
      ptr++;
  }

  *ptr2 = '\0';  

  // 1 is added by the comma
  if (strlen(sensor) + 1 + strlen(sensorf) > bufferSize) {
    Serial.printf(
      "%s buffer size has been exceeded. "
      "Sensor buffer size: %zu, "
      "Sensor value: %zu, "
      "Max size: %d\n",
      name,
      strlen(sensor),
      strlen(sensorf),
      bufferSize);

    return;
  }

  if (strlen(sensor) > 0) {
    strcat(sensor, ",");
  }

  strcat(sensor, sensorf);
} 

void iniRetry() {
    retry.counter = 0;
}

void retryAction(uint8_t raction, uint8_t max) {
    Serial.printf(
      "retryAction.action: %s, max: %d\n",
      printAction(raction),
      max);

    retry.action = raction;
    retry.max = max;

    action = wakeupRetry;

    startChrono();
}

// doRequest requests a http method to upload metrics
// or check the status of the server. In any case you
// always get the configuration that the micro needs to work.
// In case of error indicates by parameter which action
// to execute after the retry. Also by parameter
// indicates the max of retry
bool doRequest(
  const char *method,
  const char *uri,
  const char *body,
  uint8_t actionRetry,
  uint8_t maxRetry) {

  String url = String(URLAPI) + uri;

  String res;

  if (!REST(method, url.c_str(), body, res)) {
    retryAction(actionRetry, maxRetry);
    return false;
  }

  StaticJsonDocument<200> doc;
  DeserializationError err = deserializeJson(doc, res);

  if (err.code() != DeserializationError::Code::Ok) {
    setLastError(err.c_str());
    retryAction(actionRetry, maxRetry);

    return false;
  }

  config.action = doc["Action"];
  config.buffer = doc["Buffer"];
  config.checkTransTime = doc["CheckTransTime"];
  config.collectMetricsTime = doc["CollectMetricsTime"];
  wakeUpTime = doc["WakeUpTime"];

  Serial.printf("doRequest.config (Action: '%u', ", config.action);
  Serial.printf("Buffer: '%u', ", config.buffer);
  Serial.printf("CheckTransTime: '%u', ", config.checkTransTime);
  Serial.printf("CollectMetricsTime: '%d', ", config.collectMetricsTime);
  Serial.printf("WakeUpTime: '%u')\n", wakeUpTime);

  iniRetry();
  setNextAction(config.action);

  return true;
}

// REST performs a REST request and returns the result in result.
// If there are errors it returns false
bool REST(
  const char *method,
  const char *url,
  const char *body,
  String &result) {

  if (!restInternal(true, true, method, url, body, result)) {
    return false;
  }

  return true;
}

// restInternal makes an http request.
// In this request add the header with the security token.
// If the token does not exist or is expired,
// it is automatically renewed.
bool restInternal(
    bool auth,
    bool checkStatusUnauthorized,
    const char *method,
    const char *url,
    const char *body,
    String &result) {

    Serial.printf("REST.method: %s, URL: %s, Body: '%s'\n", method, url, body);

    result = "";

    checkWIFIConnection();
	
    // If there is no security token, gets token
    if (auth) {
        oAuthToken();
    }

    if (wifics == NULL) {
      if (!http->begin(url)) {
        setLastError("HTTP begin failed. URL: " + String(url));
        return false;
      }
    } else {
      if (!http->begin(*wifics, url)) {
        setLastError("HTTPS begin failed. URL: " + String(url));
        return false;
      }
    }

    http->addHeader("Content-Type", "text/plain");
    if (auth) {
		  http->addHeader("Authorization", "Bearer " + securToken);
	  }

    int httpCode = 0;

    if (method == HTTPPost) {
      httpCode = http->POST(body);
    } else {
      httpCode = http->GET();
    }

    result = http->getString();
    http->end();

    if (checkStatusUnauthorized && httpCode == HTTP_CODE_UNAUTHORIZED) {
        Serial.println("Security token expired. Getting new token");

        // When the securToken variable is initialised,
        // when rest is called again,
        // the renewed security token is obtained
        // and the requested operation is called again 
        // with the new token.
        securToken = "";

        return restInternal(auth, false, method, url, body, result);
    }

    if (httpCode != HTTP_CODE_OK) {
      setLastError(
        "Error doing request REST. " 
        "URL: " + String(url) +
        ", Code: " + String(httpCode) +
        ", result: " + result);
      return false;
    }

    return true;
}

// oAuthToken gets security token and stores int
void oAuthToken() {
    if (securToken.length() == 0) {
        Serial.println("oAuthToken.securToken empty. Getting");
        if (!restInternal(false, false, HTTPGet, URIToken, "", securToken)) {
            return;
        }

        Serial.println("oAuthToken.securToken received");
    }
}

// checkWebSocket
void checkWebSocket() {
  if (!ws->available()) {
    if (strstr(URL, "wss") == NULL) {
      Serial.println("Using WS");
    } else {
      Serial.println("Using WSS");
      ws->setCACert(rootCACertificate);
    }

    ws->connect()
  }
}

// checkWIFIConnection checks for WIFI connection and if not, try again.
void checkWIFIConnection() {
  if (WiFi.status() != WL_CONNECTED) {
    Serial.println(
      "Connection has not been established or has been lost. "
      "Re-attempted to connect...");

    WiFi.begin(ssid, password);
    while (WiFi.status() != WL_CONNECTED) {
      delay(1000);
      Serial.println("Connecting to WiFi...");
    }

    Serial.print("Connected to WiFi: ");
    Serial.println(WiFi.localIP());
  }
}

void startChrono() {
    timeStartChrono = millis();
}

int64_t timeElapsedSec() {
    return timeElapsedMiliSec() / 1000;
}

int64_t timeElapsedMiliSec() {
    return millis() - timeStartChrono;
}

// setLastError stores the last error and prints it
void setLastError(String le) {
    lastError = le;
    Serial.printf("lastError: %s\n", lastError.c_str());
}

const char *printAction(uint8_t a) {
    switch (a) {
        case sleep:
            return "sleep";
        case transmit:
            return "transmit";
        case checkNextAction:
            return "checkNextAction";
        case wakeupCheckNextAction:
            return "wakeupCheckNextAction";
        case wakeupRetry:
            return "wakeupRetry";
        case wakeupCollectBuffer:
            return "wakeupCollectBuffer";
    }

    return "None";
}

void printMemoryInfo() {
  Serial.print("Total memory: ");
  Serial.print(ESP.getHeapSize());
  Serial.print(" bytes. Free memory: ");
  Serial.print(ESP.getFreeHeap());
  Serial.println(" bytes.");
}