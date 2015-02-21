package gpio

import (
	"fmt"
	"os"
	"errors"
	"path/filepath"
	"github.com/nasu-tomoyuki/gpiotrigger/epoll"
)

const (
	gpioBasePath		= "/sys/class/gpio"
	gpioExportPath		= "/sys/class/gpio/export"
	gpioUnexportPath	= "/sys/class/gpio/unexport"
)

type Pin struct {
	initial		bool
	number		int
	fd		int
	value		*os.File
	basePath	string
	valuePath	string
	directionPath	string
	edgePath	string
}

func newPin(number int) *Pin {
	basePath := filepath.Join(gpioBasePath, fmt.Sprintf("gpio%d", number))
	pin := &Pin{
		number: number,
		fd: -1,
		basePath: basePath,
		valuePath: filepath.Join(basePath, "value"),
		directionPath: filepath.Join(basePath, "direction"),
		edgePath: filepath.Join(basePath, "edge"),
	}
	return pin
}

func (self *Pin)open() error {
	file, err := os.OpenFile(self.valuePath, os.O_RDWR|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	self.fd = int(file.Fd())
	self.value = file
	return nil
}

func (self *Pin)close() error {
	if self.value == nil {
		return nil
	}
	self.value.Close()
	self.value = nil
	self.fd = 0
	return nil
}




var pinMap	map[int]*Pin


func Init() error {
	if nil != pinMap {
		return nil
	}
	if err := epoll.Init(); err != nil {
		return err
	}
	pinMap	= make(map[int]*Pin)
	return nil
}

func Final() error {
	if nil == pinMap {
		return nil
	}
	keys := make([]int, len(pinMap))
	for k, _ := range pinMap {
		keys = append(keys, k)
	}
	for _, v := range keys {
		Unwatch(pinMap[v].number)
	}
	if err := epoll.Final(); err != nil {
		return err
	}
	pinMap = nil
	return nil
}

func Open(number int) error {
	if err := export(number); err != nil {
		return err
	}

	pin := newPin(number)

	// set to in
	if err := write(pin.directionPath, "in"); err != nil {
		unexport(number)
		return err
	}
	if err := write(pin.directionPath, "high"); err != nil {
		unexport(number)
		return err
	}
	// set to edge falling
	if err := write(pin.edgePath, "falling"); err != nil {
		unexport(number)
		return err
	}

	if err := pin.open(); err != nil {
		unexport(number)
		return err
	}

	pinMap[pin.fd]		= pin

	return nil
}

func GetFile(fd int) *os.File {
	pin, ok := pinMap[fd]
	if false == ok {
		return nil
	}
	return pin.value
}

func findPin(number int) *Pin {
	for k, v := range pinMap {
		if v.number == number {
			return pinMap[k]
		}
	}
	return nil
}

func Close(number int) error {
	pin := findPin(number)
	if pin == nil {
		return nil
	}

	if err := pin.close(); err != nil {
		return err
	}
	if err := unexport(number); err != nil {
		return err
	}
	delete(pinMap, pin.fd)

	return nil
}

func Watch(number int, callback epoll.EventCallback) error {
	pin := findPin(number)
	if pin == nil {
		return nil
	}


	if err := epoll.Watch(pin.fd, callback); err != nil {
		unexport(number)
		return err
	}

	return nil
}

func Unwatch(number int) error {
	pin := findPin(number)
	if pin == nil {
		return nil
	}

	if err := epoll.Unwatch(pin.fd); err != nil {
		return err
	}

	return nil
}

/*
func WriteValue(number int, ok bool) error {
	pin := findPin(number)
	if pin == nil {
		return nil
	}

	file := pin.value

	var f string
	if ok {
		f	= "1"
	} else {
		f	= "0"
	}

	file.Seek(0, 0)
	file.Write([]byte(f))
	return nil
}
*/

func ReadValue(number int) (int, error) {
	pin := findPin(number)
	if pin == nil {
		return 0, nil
	}

	file := pin.value

	var num int
	num = -1
	file.Seek(0, 0)
	b	:= make([]byte, 1000)
	_, err := file.Read(b)
	if err != nil {
		return 0, err
	}
	num = int(b[0]) - int('0')
	return int(num), nil
}



func write(path string, format string, args ...interface{}) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	s := fmt.Sprintf(format, args...)
	fmt.Fprint(file, s)
	return nil
}

func export(number int) error {
	filename := filepath.Join(gpioBasePath, fmt.Sprintf("gpio%d/value", number))
	// it's error if the file exists
	if _, err := os.Stat(filename); err == nil {
		return errors.New("already exported")
	}

	if err := write(gpioExportPath, "%d", number); err != nil {
		return err
	}
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	return nil
}

func unexport(number int) error {
	err := write(gpioUnexportPath, "%d", number)
	if err != nil {
		return err
	}
	return nil
}
