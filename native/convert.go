package native

import (
	"reflect"
	"unsafe"
)

// StringToByteSlice unsafe
func StringToByteSlice(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	*(*string)(unsafe.Pointer(&b)) = s
	bh.Cap = len(s)
	return
}

// ByteSliceToString unsafe
func ByteSliceToString(b []byte) (s string) {
	s = *(*string)(unsafe.Pointer(&b))
	return
}
