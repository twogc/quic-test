//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

// notifyResize регистрирует обработчик SIGWINCH для Unix-систем
func notifyResize(c chan<- os.Signal) {
	signal.Notify(c, syscall.SIGWINCH)
}

