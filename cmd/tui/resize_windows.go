//go:build windows
// +build windows

package main

import "os"

// notifyResize - пустая функция для Windows (SIGWINCH не поддерживается)
func notifyResize(c chan<- os.Signal) {
	// На Windows SIGWINCH не существует, поэтому ничего не делаем
}

