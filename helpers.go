package glutils

import (
	"reflect"
	"github.com/go-gl/mathgl/mgl32"
	"fmt"
	"strconv"
)

// checks if o implements i
// i must be an interface and not an instance of a struct.
// example: (*MyInterface)(nil)
func Implements(o, i interface{}) bool {
	ot := reflect.ValueOf(o).Type()
	bs := reflect.TypeOf(i).Elem()
	return ot.Implements(bs)
}

// compares 2 objects and determines if they have the same type.
func IsType(o, i interface{}) bool {
	a := reflect.ValueOf(o).Type()
	b := reflect.ValueOf(i).Type()
	return a == b
}

func PrintMat4(m mgl32.Mat4) {
	fmt.Printf("%s\n%s\n%s\n%s\n-------\n",
		ftos([]float32{m[0], m[4], m[8], m[12]}),
		ftos([]float32{m[1], m[5], m[9], m[13]}),
		ftos([]float32{m[2], m[6], m[10], m[14]}),
		ftos([]float32{m[3], m[7], m[11], m[15]}),
	)
}

func ftos(f []float32) string {
    out := ""
	for i := range f {
        out += strconv.FormatFloat(float64(f[i]), 'f', 2, 32) + ", "
    }
	return out
}