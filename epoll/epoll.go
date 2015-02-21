package epoll

import (
	"fmt"
	"syscall"
)

const (
	MaxEpollEvents		int = 32
)

type EventCallback	func(event *syscall.EpollEvent)

type RegisteredEvent struct {
	initial		bool
	fd			int
	callback	EventCallback
}

func newRegisteredEvent(fd int, callback EventCallback) *RegisteredEvent {
	re := RegisteredEvent{
		initial: true,
		fd: fd,
		callback: callback}
	return &re
}


var registeredEvents		map[int]*RegisteredEvent


func Init() error {
	if nil != registeredEvents {
		return nil
	}
	registeredEvents		= make(map[int]*RegisteredEvent)
	if err := setupEpoll(); err != nil {
		return err
	}
	return nil
}

func Final() error {
	syscall.Close(epollFd)

	registeredEvents = nil
	epollFd	= 0

	return nil
}

func Watch(fd int, callback EventCallback) error {
	registeredEvent			:= newRegisteredEvent(fd, callback)
	registeredEvents[fd]	= registeredEvent

	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN | (syscall.EPOLLET & 0xffffffff) | syscall.EPOLLPRI

	if err := syscall.SetNonblock(fd, true); err != nil {
		return err
	}

	event.Fd = int32(fd)

	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		return err
	}
	return nil
}

func Unwatch(fd int) error {
	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
		return err
	}
	if err := syscall.SetNonblock(fd, false); err != nil {
		return err
	}

	delete(registeredEvents, fd)

	return nil
}


var epollFd int

func setupEpoll() error {
	var err error
	epollFd, err = syscall.EpollCreate1(0)
	if err != nil {
		return err
	}

	go func() {
		var epollEvents [MaxEpollEvents]syscall.EpollEvent

		for {
			numEvents, err := syscall.EpollWait(epollFd, epollEvents[:], -1)
			if err != nil {
				if err == syscall.EAGAIN {
					continue
				}
				panic(fmt.Sprintf("EpollWait error: %v", err))
			}
			for i := 0; i < numEvents; i++ {
				if re, ok := registeredEvents[int(epollEvents[i].Fd)]; ok {
					if re.initial {
						re.initial = false
					} else {
						re.callback(&epollEvents[i])
					}
				}
			}
		}
	}()

	return nil
}



