// Contains the main entry point 
// for using the imgproc library as a executable.
// See Usage* functions for details.

package main

import (
	"bufio"
	"bytes"
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

type imageEncoder func(io.Writer, image.Image) error

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

	// convert to floatImage, perform operations, and save
	fImg := imgproc.ImageToFloatImage(image)
	op(fImg)
	return encode(output, fImg)
}

func printErrAndUsage(err error) {
	fmt.Fprintln(os.Stderr, err, "\n---\n"+usageMain())
}

func makeHelpMessage(helpRequests []string) string {
	
	// use a byte-buffer for efficiency, similar to using StringBuilder in Java.
	res := bytes.NewBufferString("Help:\n")

	// iterate over requests, lookup the help string in supported ops map
	for _, req := range helpRequests {
		op, found := supported_ops[req]
		if found {
			res.WriteString(fmt.Sprintln("\t", req, op.Desc, "\n\t\t", op.Usage))
		} else { // key unrecognized:
			res.WriteString(fmt.Sprintln("\t", req, "is not a supported operation"))
		}
	}
	
	// always append on the overall usage string
	res.WriteString("\n---\n" + usageMain())
	return res.String()
}

func buildOperations(operations []string) (ImageOp,error) {

	fullOp := IdentityOp 

	for keyword, args := range collectArgs(operations) {
		op, found := supported_ops[keyword]
		if found {
			fullOp = Compose(fullOp, op.Factory(args))
		} else {
			return nil, errors.New(keyword + " is not a supported operation")
		}
	}

	return fullOp, nil 
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
	op, err := buildOperations(operations)
	if err != nil {
		printErrAndUsage(err)
		return
	}

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
