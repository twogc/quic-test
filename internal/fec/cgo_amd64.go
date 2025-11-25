//go:build amd64 && cgo

package fec

/*
#cgo CFLAGS: -mavx2
#cgo linux LDFLAGS: -lfec_avx2 -lnuma
#cgo darwin LDFLAGS: -lfec_avx2
*/
import "C"
