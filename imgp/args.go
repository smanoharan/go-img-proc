// Args.go: for handling command line arguments. 
// See the usage* functions for more details.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
)

// builds the main Usage string
func usageMain() string {
	return "Usage: imgp [-i[n] files...] [-d[o] operations...] [-o[ut] (j|jpg|jpeg|p|png)]\n\n" +

		"\t-in (or -i for short) specifies the input file(s).\n" +
		"\t\tThe input files can be either jpg, gif, or png.\n" +
		"\t\tIf no input files are specified, or if the -in is omitted,\n" +
		"\t\tthe input files will be read from stdin, one file per line.\n\n" +

		"\t-out (or -o) specifies the output format of the file(s).\n" +
		"\t\tThe default output is png.\n" +
		"\t\tOnly one output format can be specified, and this chosen\n" +
		"\t\textension will be appended onto each of the input files.\n" +
		"\t\tE.g. \"imgproc -i ./bar/foo.jpg -o p\" will result in\n" +
		"\t\ta file named \"foo.jpg.png\" being placed in the folder \"./bar/\"\n\n" +

		"\t-do (or -d) specifies the operations(s) to apply to each image.\n" +
		"\t\tThe operations must be specified as list, separated by '+'.\n" +
		"\t\tEach operation must be in the form <keyword> par1=v1 par2=v2 ...\n" +
		"\t\tE.g. \"imgp -i file1 -d scale s=2 + flip o=vert -o p\")\n" +
		"\t\tIf only image format conversion is required, no operations need to be specified.\n\n" +

		"\tSupported Operations: (run \"imgp -h[elp] operation\" for more details on each operation).\n" +
		listSupportedOps() + "\n"
}

func listSupportedOps() string {
	res := bytes.NewBufferString("")
	for keyword, op := range supported_ops {
		res.WriteString(fmt.Sprintln("\t\t", keyword, op.Desc))
	}
	return res.String()
}

// a new flag-type: array of strings
type strArr []string

func (s *strArr) String() string {
	return strings.Join(*s, "|")
}

func (s *strArr) Set(value string) error {
	for _, elem := range strings.Split(value, "|") {
		*s = append(*s, elem)
	}
	return nil
}

// for suppressing output
type EmptyWriter struct {
}

func (e *EmptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil // do nothing
}

// preprocess args:
// group multi-value args into one field:
func preprocessArgs(args []string) []string {
	res := make([]string, 0)

	lastIndexStart := 0 // location of the start of the current multi-arg
	for index, arg := range args {
		if arg[0] == '-' {
			// current arg is the start of a flag
			if lastIndexStart < index {
				// copy over previous multi-arg
				res = append(res, strings.Join(args[lastIndexStart:index], "|"))
			}
			res = append(res, arg)
			lastIndexStart = index + 1
		}
	}

	// deal with remaining multi-arg (if any)
	lenArgs := len(args)
	if lastIndexStart < lenArgs {
		res = append(res, strings.Join(args[lastIndexStart:lenArgs], "|"))
	}
	return res
}

// parse command line args
func parseArgs() (input, operations, help strArr, output string, err error) {

	const (
		defaultOutType = "png"
		usage          = ""
	)

	flags := flag.NewFlagSet("main", flag.ContinueOnError)
	flags.SetOutput(&EmptyWriter{}) // suppress output. We have custom error printing.

	// out and o share the variable: output
	flags.StringVar(&output, "out", defaultOutType, usage)
	flags.StringVar(&output, "o", defaultOutType, usage)

	// in and i share the variable: input
	flags.Var(&input, "in", usage)
	flags.Var(&input, "i", usage)

	// help and h share the variable: help
	flags.Var(&help, "help", usage)
	flags.Var(&help, "h", usage)

	// do and d share the variable: operations
	flags.Var(&operations, "do", usage)
	flags.Var(&operations, "d", usage)

	err = flags.Parse(preprocessArgs(os.Args[1:]))
	return
}

// convert an array of the form:
//	 [ keyword1 par11=v11 par12=v12 ... + keyword2 par21=v21 ... ... ]
// into the (map) form
//   keyword1: [-par11=v11 -par12=v12 ... ]
//	 keyword2: [-par21=v21 ... ]
func collectArgs(ops []string) map[string][]string {
	res := make(map[string][]string)

	// prepend a dash to each elem in ops:
	for i := range ops { 
		ops[i] = "-" + ops[i]
	}
	
	last := 0 // location of the start of the current operation
	for cur, arg := range ops {
		if arg == "-+" {
			res[ops[last][1:]] = ops[last:cur]
			last = cur + 1
		}
	}

	// deal with remaining operation
	end := len(ops)
	if last < end  {
		res[ops[last][1:]] = ops[last:end]
	}
	return res
}

