package main

import (
	"flag"
	"github.com/stianeikeland/go-rpio"
	"log"
	"time"
	"github.com/yryz/ds18b20"
	"fmt"
	"net/http"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"os"
)

// todo: add build script to build .deb (daemontools)

// todo: install watchdog & power mgmt, collectd?, readonly FS, hostapd

var (
	relayGpioPtr = flag.Int("gpio-relay", 22, "Gpio of relay")
	statusLedGpioPtr = flag.Int("gpio-led-status", 15, "Gpio of status LED")
	relayLedGpioPtr = flag.Int("gpio-led-relay", 14, "Gpio of relay LED")
	enableOnStartupPtr = flag.Bool("enable-on-startup", false, "Enable heater on startup")
	relayLedBrightnessPtr = flag.Int("relay-led-brightness", 10, "Brightness of relay LED (percentage 0-100)")
)

type HeaterData struct {
	heater	rpio.Pin
	sensors	[]string
}

func returnHeaterStatus(pin *rpio.Pin) string {
	status := "error"
	switch rpio.ReadPin(*pin) {
	case rpio.Low:
		return "on"
	case rpio.High:
		return "off"
	}

	return status
}


func showtime(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Time: " + time.Now().String())
}

func statusGoRoutine(){
	for range time.Tick(time.Second * 1){
		setLed(*statusLedGpioPtr, 1)
		time.Sleep(time.Millisecond * 120)
		setLed(*statusLedGpioPtr, 2)
		time.Sleep(time.Millisecond * 120)
	}
}

func heaterOn(heater *rpio.Pin){
	heater.Low()
	setLed(*relayLedGpioPtr, *relayLedBrightnessPtr)
	fmt.Println("Turning heater on")
}

func heaterOff(heater *rpio.Pin){
	heater.High()
	setLed(*relayLedGpioPtr, 0)
	fmt.Println("Turning heater off")
}

func (heaterData *HeaterData) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Heater: " + returnHeaterStatus(&heaterData.heater) + "\n")
	for _, sensor := range heaterData.sensors {
		t, err := ds18b20.Temperature(sensor)
		if err == nil {
			io.WriteString(w, fmt.Sprintf("Temp: %s: %.2f°C\n", sensor, t))
		}
	}
	dat, _ := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	cpuTemp, _ := strconv.Atoi(strings.TrimSpace(string(dat)))
	io.WriteString(w, fmt.Sprintf("Temp: CPU: %.2f\n", float32(cpuTemp)/1000))
	io.WriteString(w, "Time: " + time.Now().String())
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func setLed(gpio int, brightness int){
	if _, err := os.Stat("/dev/pi-blaster"); os.IsNotExist(err) {
		fmt.Errorf("%s\n", "/dev/pi-blaster does not exist! Is it running?")
		panic(err)
	}
	brOut := float64(brightness) / 100
	brOut2 := fmt.Sprintf("%.2f", brOut)
	d1 := []byte(strconv.Itoa(gpio) + "=" + brOut2 + "\n")
	err := ioutil.WriteFile("/dev/pi-blaster", d1, 0644)
	check(err)
}

func main() {
	flag.Parse()

	err := rpio.Open()
	if err != nil {
		log.Fatal(err.Error())
	}

	heater := rpio.Pin(*relayGpioPtr)
	heater.Output()
	if *enableOnStartupPtr == true {
		heaterOn(&heater)
	}else{
		heaterOff(&heater)   // Disable heater
	}

	go statusGoRoutine()

	sensors, err := ds18b20.Sensors()

	for _, sensor := range sensors {
		t, err := ds18b20.Temperature(sensor)
		if err == nil {
			fmt.Printf("sensor: %s temperature: %.2f°C\n", sensor, t)
		}
	}

	heaterdata := &HeaterData{heater: heater, sensors: sensors}

	mux := http.NewServeMux()
	mux.HandleFunc("/time", showtime)
	mux.HandleFunc("/status", heaterdata.status)
	mux.HandleFunc("/on", func(w http.ResponseWriter, r *http.Request){
		heaterOn(&heater)
		io.WriteString(w, "OK:Enabled\nTime: " + time.Now().String())
	})
	mux.HandleFunc("/off", func(w http.ResponseWriter, r *http.Request){
		heaterOff(&heater)
		io.WriteString(w, "OK:Disabled\nTime: " + time.Now().String())
	})
	http.ListenAndServe(":80", mux)
}