package utils

import "unsafe"

const INT8 int8 = 1
const INT16 int16 = 1
const INT32 int32 = 1
const INT64 int64 = 1

func SizeofIn8() uintptr {
	return unsafe.Sizeof(INT8)
}

func SizeofIn16() uintptr {
	return unsafe.Sizeof(INT16)
}

func SizeofIn32() uintptr {
	return unsafe.Sizeof(INT32)
}

func SizeofIn64() uintptr {
	return unsafe.Sizeof(INT64)
}
