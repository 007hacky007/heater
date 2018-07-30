# heater

I created "heater" in my beginnings with IoT for Raspberry Zero W. 
It is obviously a huge overkill to use Raspberry for switching (single) relay, but hey, it was cold 
and I wanted to control my heater remotely. It worked well for one winter, however I've rebuilt 
this in C++ for ESP8266 which suits the purpose much better.

Few things to note
  - heater.go is the original version which blinks the status LED, so you know it is running
  - heater_pi-blaster.go as the name suggest is the 2nd version (final) which utilizes the pi-blaster to dim the LEDs
  - oh and it uses ds18b20 sensor, so you know which temperature is around your heater. Sounds useless? It actually is.



