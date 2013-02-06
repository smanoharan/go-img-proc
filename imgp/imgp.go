// Contains the main entry point 
// for using the imgproc library as a executable.
// See Usage* functions for details.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/smanoharan/go-img-proc/imgproc"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

type ImageOp func(*imgproc.FloatImage) *imgproc.FloatImage
type imageEncoder func(io.Writer, image.Image) error

// Compose two Image operations into a single operation.
// I.e. if h := Compose(f,g), then h(img) == g(f(img))
func Compose(op1, op2 ImageOp) ImageOp {
	return func(img *imgproc.FloatImage) *imgproc.FloatImage {
		return op2(op1(img)) // perform op1, then op2.
	}
}

// Simply pass along the image, without modifying it.
func IdentityOp(img *imgproc.FloatImage) *imgproc.FloatImage {
	return img
}

func toOutputEncoder(output string) (imageEncoder, error) {
	switch output {
	case "j", "jpg", "jpeg":
		return func(w io.Writer, m image.Image) error { return jpeg.Encode(w, m, nil) }, nil
	case "p", "png":
		return png.Encode, nil
	// case "g","gif": 
	//	return gif.Encode, nil // gif encoding is not supported in GO
	}
	return nil, errors.New("Unrecognized output format: " + output)
}

// read the inputFile, perform op and save as inputFile.outputFormat, using the supplied encoder.
func processFile(inputFile, outputFormat string, encode imageEncoder, op ImageOp) error {

	// open input file
	input, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	// check output file is writable
	output, err := os.Create(inputFile + "." + outputFormat)
	if err != nil {
		return err
	}
	defer output.Close()

	// decode into an image
	image, _, err := image.Decode(bufio.NewReader(input))
	if err != nil {
		return err
	}

	// perform operations and save
	return encode(output, op(imgproc.ImageToFloatImage(image)))
}

func printErrAndUsage(err error) {
	fmt.Fprintln(os.Stderr, err, "\n---\n"+usageMain())
}

func makeHelpMessage(helpRequests []string) string {
	// TODO
	return usageMain()
}

func buildOperations(operations []string) ImageOp {
	// TODO
	return IdentityOp
}

func main() {

	// parse arguments:
	input, operations, help, output, err := parseArgs()
	if err != nil {
		printErrAndUsage(err)
		return
	}

	// deal with help messages
	if help != nil && len(help) > 0 {
		fmt.Fprintln(os.Stderr, makeHelpMessage(help))
		return // if help requested, ignore other params.
	}

	// expand and verify output format:
	outputEncoder, err := toOutputEncoder(output)
	if err != nil {
		printErrAndUsage(err)
		return
	}

	// compose operations 
	op := buildOperations(operations)

	// if input is empty, read from stdin
	if input == nil || len(input) == 0 {
		// read from stdin
		line := ""
		for n, err := fmt.Scan(&line); err == nil && n > 0; n, err = fmt.Scan(&line) {
			input = append(input, line)
		}
	}

	// iterate over each file:
	for _, inputFile := range input {
		err = processFile(inputFile, output, outputEncoder, op)
		if err != nil {
			printErrAndUsage(err)
			return
		}
	}
}
