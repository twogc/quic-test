//go:build arm64 && cgo

package fec

/*
#cgo CFLAGS: -march=armv8-a+simd
#cgo LDFLAGS: -lfec_neon
*/
import "C"
