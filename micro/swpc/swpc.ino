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

// Hardware utilizado en el proyecto y pines usados para ello
// **********************************************************

// PIN 2  (GPIO 2)    Sensor de temperatura DS18B20
// PIN 34 (ADC1_CH6)  Sensor de PH SKU-SEN0169
// PIN 35 (ADC1_CH7)  Sensor de ORP
// PIN 4  (GPIO 4)    LED rojo para señalizar que el micro no está en modo SLEEP

// Librerías necesarias para el sensor de temperatura DS18B20
// **********************************************************

#include <OneWire.h>                
#include <DallasTemperature.h>

// Constantes necesarias para el programa
// **************************************

#define FACTOR_DE_MICROSEG_A_SEG 1000000ULL       // Factor de conversión de microsegndos a segundos para el tiempo del modo SLEEP
#define TIEMPO_PARA_SLEEP  10                     // Tiempo para que el micro ESP32 vuelva a activarse tras el modo SLEEP (en segundos)
#define LECTURAS_PARA_SLEEP 5                     // Número de lecturas de los sensores antes de que el micro pase al modo SLEEP
#define TIEMPO_RETARDO_LECTURAS 3000              // Constante de tiempo para la función DELAY() en milisegundos 

#define PIN_SENSOR_PH  34                         // Número de pin para leer el sensor de PH => Pin 34 = Canal 6 del ADC1

#define PIN_SENSOR_ORP 35                         // Número de pin para leer el sensor de ORP => Pin 35 = Canal 7 del ADC1
#define VOLTAGE_ORP 5.00                          // Tensión de referencia del sensor de ORP 
#define OFFSET_ORP 13                             // Valor para la calibración del sensor de ORP según la página del fabricante. Habría que verlo bien. ¡¡PUEDE SER VARIABLE!!

// Variables necesarias para el programa
// *************************************

RTC_DATA_ATTR int numero_de_arranques = 0;        // Variable para contar el número de rearranques desde el modo SLEEP
int numero_de_lecturas = 0;                       // Variable para contar el número de lecturas realizadas antes de pasar al modo SLEEP

OneWire ourWire(0);                               // Variable necesaria para la librería del sensor de temperatura DS18B20 que ajusta el pin 2 como bus OneWire
DallasTemperature sensors(&ourWire);              // Estructura necesaria para la librería del sensor de temperatura DS18D20
float temperatura_valor_final=0;                  // Variable final para guardar la lectura de la temperatura con la librería del sensor DS18B20 en grados centígrados

float ph_valor_final = 0;                         // Variable final para guardar la lectura del PH en el rango de PH (0-14)     

double orp_valor_final;                           // Variable final para guardar la lectura del valor de ORP en mV

// Función para imprimir el motivo por el que se despiesta el micro desde el modo SLEEP
// ************************************************************************************ 

void print_wakeup_reason()
{  
  esp_sleep_wakeup_cause_t motivo_del_rearranque;            // Variable para guardar el motivo por el que se despiesta el micro desde el modo SLEEP 

  motivo_del_rearranque = esp_sleep_get_wakeup_cause();

  switch(motivo_del_rearranque)
  {
    case ESP_SLEEP_WAKEUP_TIMER : Serial.println("El rearranque tras el modo SLEEP ha sido realizado por tiempo despues de " + String(TIEMPO_PARA_SLEEP) + " segundos dormido"); break;
    case ESP_SLEEP_WAKEUP_EXT0 : Serial.println("El rearranque tras el modo SLEEP ha sido realizado por una señal externa usando RTC_IO"); break;
    case ESP_SLEEP_WAKEUP_EXT1 : Serial.println("El rearranque tras el modo SLEEP ha sido realizado por una señal externa usando RTC_CNTL"); break;
    case ESP_SLEEP_WAKEUP_TOUCHPAD : Serial.println("El rearranque tras el modo SLEEP ha sido realizado por un Touchpad"); break;
    case ESP_SLEEP_WAKEUP_ULP : Serial.println("El rearranque tras el modo SLEEP ha sido realizado por una señal externa usando 'ULP program'"); break;
    default : Serial.printf("El rearranque no ha sido realizado tras el modo SLEEP\n"); break;
  }
}

// Función para leer el sensor de temperatura
// ****************************************** 

 float funcion_leer_sensor_temperatura(void){
  
  float temp=0;                                             // Variable interna a la función para leer la temperatura

  sensors.requestTemperatures();                            // Primero se envía el comando para leer el sensor de temperatura DS18B20
  temp = sensors.getTempCByIndex(0);                        // Ahora se guarda el valor´de temperatura léido en ºC dentro de la variable "temp"
  
  return temp;                                              // Y se devuelve la temperatura al programa
}

// Función para leer el sensor de PH
// ********************************* 

 float funcion_leer_sensor_PH(void){
  
  float ph=0;                                       // Variable interna a la función para leer el PH
  float ph_valor_en_milivoltios = 0;                // Valor de PH final en milivoltios
  float valor_referencia_para_PH_4 = 1280;          // Valor de tensión de referencia para PH = 4.0 obtenido en la primera calibración>
  float valor_referencia_para_PH_7 = 1690;          // Valor de tensión de referencia para PH = 7.0 obtenido en la primera calibración>
  float pendiente = 0;                              // Variable necesaria para la fórmula del cálculo del PH   
  float offset = 0;                                 // Variable necesaria para la fórmula del cálculo del PH 

  ph_valor_en_milivoltios = analogRead(PIN_SENSOR_PH)/4095.0*3300;  // Leo la tensión del sensor y la paso a milivoltios
      
  pendiente = (7.0-4.0)/((valor_referencia_para_PH_7-1500)/3.0-(valor_referencia_para_PH_4-1500)/3.0);  // Fómulas sacadas de Internet
  offset = 7.0-pendiente*(valor_referencia_para_PH_7-1500)/3.0;                                         // para calibrar el sensor de PH

  ph = pendiente*(ph_valor_en_milivoltios-1500)/3.0+offset;         // Valor final de PH = (slope*tensión)+offset

  return ph;                                              // Y se devuelve el PH al programa
}

// Función para leer el sensor ORP
// ******************************* 

 float funcion_leer_sensor_ORP(void){
  
  float orp=0;                                            // Variable interna a la función para leer el ORP
 
  orp=((30*(double)VOLTAGE_ORP*1000)-(75*analogRead(PIN_SENSOR_ORP)*VOLTAGE_ORP*1000/4095))/75-OFFSET_ORP;   // Fórmula del ORP sacada de Internet

  return orp;                                             // Y se devuelve el ORP al programa
}

// Función SETUP del programa para los ajustes iniciales necesarios
// **************************************************************** 

void setup() 
{
  pinMode(GPIO_NUM_2, INPUT);         // Pin 2 = Entrada digital para el sensor de temperatura DS18B20
  pinMode(GPIO_NUM_4, OUTPUT);        // Pin 4 = Salida digital para el LED rojo
  
  Serial.begin(9600);                 // Se inicializa la velocidad del puerto serie para el monitor serie
    
  ++numero_de_arranques;              // Se incrementa la variable de contaje de arranques

  Serial.println("     ");            // Se muestra en el monitor serie que el microsontrolador ha rearrancado de nuevo
  Serial.println("El microcontrolador ESP32 ha rearrancado de nuevo.");

  print_wakeup_reason();              // Se muestra en el monitor serie el motivo por el que se despierta el micro desde el modo SLEEP
  
  Serial.println("El numero de arranques desde la conexion inicial han sido " + String(numero_de_arranques)); // Se muestra en el monitor serie el número  
  Serial.println("     ");                                                                                    // de rearranques desde la conexión
  Serial.println("-------------------------------------------------------"); 
  Serial.println("     ");
  Serial.println("Se haran " + String(LECTURAS_PARA_SLEEP) + " medidas antes de volver al modo SLEEP"); // Y se indican la medidas que se harán 
  Serial.println("     ");
  Serial.println("-------------------------------------------------------"); 

  esp_sleep_enable_timer_wakeup(TIEMPO_PARA_SLEEP * FACTOR_DE_MICROSEG_A_SEG);  // Se ejusta que la fuente de despertar del microcontrolador desde el modo SLEEP
                                                                                // será un timer el tiempo indicado en la constante TIEMPO_PARA_SLEEP (en segundos)

  digitalWrite(GPIO_NUM_4, HIGH);   // Se enciende el LED rojo para ver visualmente que ya no se está en modo SLEEP

  sensors.begin();                  // Se arranca el sensor de temperatura DS18B20

  delay(TIEMPO_RETARDO_LECTURAS);   // Se retarda el tiempo dado en la constante TIEMPO_RETARDO_LECTURAS para ver los mensajes del rearraque en el monitor serie                       
}

// Función LOOP del programa: Programa cíclico del micro
// ***************************************************** 

void loop() 
{  
    if (numero_de_lecturas == LECTURAS_PARA_SLEEP) // Si ya se han hecho las lecturas indicadas en la contante LECTURAS_PARA_SLEEP, se pasa al modo SLEEP
                                                   // el tiempo indicado en la constante TIEMPO_PARA_SLEEP (en segundos) y se muestra en el monitor serie
    {                   
      Serial.println("     ");
      Serial.println("El microcontrolador ESP32 pasa a modo SLEEP durante " + String(TIEMPO_PARA_SLEEP) + " segundos");
      Serial.println("     "); 
      Serial.println("-------------------------------------------------------"); 
      Serial.println("     "); 

      esp_deep_sleep_start();
    }
    else  // Si no se han hecho las lecturas indicadas en la contante LECTURAS_PARA_SLEEP, se leen los sensores 
    {
     
      Serial.println("   ");                         // Al principio de cada lectura se muestra el número de lectura
      Serial.println("Lectura numero " + String(numero_de_lecturas+1) );   

      // Lectura del sensor de temperatura temperatura DS18B20

      temperatura_valor_final = funcion_leer_sensor_temperatura();   // Llamo a la función para leer el sensor de temperatura

      Serial.println("   ");                                                                                // Y finalmente se muestra en el monitor serie
      Serial.println(" -> La temperatura del agua son " + String(temperatura_valor_final) + " grados centigrados"); 

      // Lectura del sensor de PH SKU-SEN0169 
      
      ph_valor_final = funcion_leer_sensor_PH();       // Llamo a la función para leer el sensor de PH 
      
      Serial.println(" -> El valor de PH del agua es " + String(ph_valor_final));                                    // Y finalmente se muestra en el monitor serie
  
      // Lectura del sensor ORP
    
      orp_valor_final = funcion_leer_sensor_ORP();     // Llamo a la función para leer el sensor ORP 
      
      Serial.println(" -> El valor de ORP del agua es " + String(orp_valor_final) + " mV");                 // Y finalmente se muestra en el monitor serie
      Serial.println("   "); 
      Serial.println("-------------------------------------------------------"); 

      numero_de_lecturas++;                    // Se incrementa la variable que cuenta el número de lecturas para poder realizar la siguiente
      
      delay(TIEMPO_RETARDO_LECTURAS);          // Se retarda un tiempo indicado en la constante TIEMPO_RETARDO_LECTURAS para empezar con la siguiente lectura 
    }
 }