package qsort

import "C"
import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// #include <stdlib.h>
// typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
// extern int _cgo_qsort_compare(void* a, void* b);
import "C"

type CompareFunc C.qsort_cmp_func_t

var goQSortCompareInfo struct {
	base     unsafe.Pointer
	elemNum  int
	elemSize int
	less     func(a, b int) bool
	sync.Mutex
}

//export _cgo_qsort_compare
func _cgo_qsort_compare(a, b unsafe.Pointer) C.int {
	var (
		// array memory is locked
		base     = uintptr(goQSortCompareInfo.base)
		elemSize = uintptr(goQSortCompareInfo.elemSize)
	)

	i := int((uintptr(a) - base) / elemSize)
	j := int((uintptr(b) - base) / elemSize)

	switch {
	case goQSortCompareInfo.less(i, j): // v[i] < v[j]
		return -1
	case goQSortCompareInfo.less(j, i): // v[i] > v[j]
		return +1
	default:
		return 0
	}
}

func Slice(slice any, less func(a, b int) bool) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("qsort called with non-slice value of type %T", slice))
	}
	if sv.Len() == 0 {
		return
	}

	goQSortCompareInfo.Lock()
	defer goQSortCompareInfo.Unlock()

	defer func() {
		goQSortCompareInfo.base = nil
		goQSortCompareInfo.elemNum = 0
		goQSortCompareInfo.elemSize = 0
		goQSortCompareInfo.less = nil
	}()

	// baseMem = unsafe.Pointer(sv.Index(0).Addr().Pointer())
	// baseMem maybe moved, so must saved after call C.fn
	goQSortCompareInfo.base = unsafe.Pointer(sv.Index(0).Addr().Pointer())
	goQSortCompareInfo.elemNum = sv.Len()
	goQSortCompareInfo.elemSize = int(sv.Type().Elem().Size())
	goQSortCompareInfo.less = less

	C.qsort(
		goQSortCompareInfo.base,
		C.size_t(goQSortCompareInfo.elemNum),
		C.size_t(goQSortCompareInfo.elemSize),
		C.qsort_cmp_func_t(C._cgo_qsort_compare),
	)
}
