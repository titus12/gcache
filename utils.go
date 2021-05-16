package gcache

import (
	"github.com/davecgh/go-spew/spew"
	"runtime"
)

func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func isPowerOfTwo(number int) bool {
	return (number & (number - 1)) == 0
}

func PrintPanicStack(extras ...interface{}) {
	if x := recover(); x != nil {
		l.Errorf("%v", x)
		i := 0
		funcName, file, line, ok := runtime.Caller(i)
		for ok {
			l.Errorf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
			i++
			funcName, file, line, ok = runtime.Caller(i)
		}

		for k := range extras {
			l.Errorf("EXRAS#%v DATA:%v\n", k, spew.Sdump(extras[k]))
		}
	}
}
