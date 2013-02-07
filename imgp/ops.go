// Defines the supported operations, along with their usage statements and arg handling
package main

import (
	"github.com/smanoharan/go-img-proc/imgproc"
	)

// function signature for each operation: mutate the input image. 
type ImageOp func(*imgproc.FloatImage)

// do nothing
func IdentityOp(img *imgproc.FloatImage) {
}

// Compose two Image operations into a single operation.
// I.e. if h := Compose(f,g), then h(img) is equivalent to g(f(img))
func Compose(op1, op2 ImageOp) ImageOp {
	return func(img *imgproc.FloatImage) {
		// perform op1 then op2
		op1(img)
		op2(img)
	}
}

// for each op, we need:
//	keyword (i.e. name)
//	1-line description and a full usage message
//  argument interpreter : takes []string and returns ImageOp
type supportedOp struct {
	Desc, Usage string
	Factory func(args []string) ImageOp
}

func IdentityFactory(args []string) ImageOp {
	return IdentityOp
}

var supported_ops map[string]supportedOp = map[string]supportedOp {
	"ident": { 
		Desc: "<no arguments> -- Identity transform",
		Usage: "Identity transform: does not modify the image",
		Factory: IdentityFactory,
	}, 
}
