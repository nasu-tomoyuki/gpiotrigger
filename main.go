package main

import (
	"os/exec"
	"flag"
	"fmt"
	"syscall"
	"time"
	"os"
	"github.com/nasu-tomoyuki/gpiotrigger/gpio"
)

var command string
var gpioTarget int
var watchingSeconds int
var isWatching bool

func epollCallback(event *syscall.EpollEvent) {
	// prevent to re enter by chattering
	if isWatching {
		return
	}

	isWatching = true

	// begin to watch
	go func () {
		sv := 0
		sec := 0
		for {
			v, err := gpio.ReadValue(gpioTarget)
			if err == nil {
				if sec >= watchingSeconds {
					out, _ := exec.Command(os.Getenv("SHELL"), "-c", command).Output()
					fmt.Println(string(out))

					// close to exit
					gpio.Unwatch(gpioTarget)
					gpio.Close(gpioTarget)
					os.Exit(0)
				}
				if v != sv {
//					fmt.Println("v: ", v, " sec: ", sec)
					isWatching = false
					return
				}
			}
//			fmt.Println("> v: ", v, " sec: ", sec)
			time.Sleep(time.Second)
			sec += 1
		}
	}()
}

func main() {
	optCommand := flag.String("command", "echo hello, world", "execute command")
	optWatchingSeconds := flag.Int("time", 5, "watching time(sec)")
	optGpioTarget := flag.Int("pin", 27, "target pin")
	flag.Parse()

	command = *optCommand
	watchingSeconds = *optWatchingSeconds
	gpioTarget = *optGpioTarget

	isWatching	= false

	if err := gpio.Init(); err != nil {
		fmt.Println(err)
		return
	}
	defer gpio.Final()

	if err := gpio.Open(gpioTarget); err != nil {
		fmt.Println(err)
		return
	}

	if err := gpio.Watch(gpioTarget, epollCallback); err != nil {
		fmt.Println(err)
		return
	}

	for {
		// sleep for an hour
		time.Sleep(time.Hour)
	}
}

