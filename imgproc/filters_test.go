// Test file for floatImage.go

package imgproc

import (
	"bytes"
	"fmt"
	"math"
	"testing"
)

func assert(t *testing.T, cond bool, msg string) bool {
	if !cond {
		t.Errorf(msg)
	}
	return cond
}

func assertIntEquals(t *testing.T, exp, act int, title string) bool {
	return assert(t, exp == act, title+fmt.Sprintf(": exp=%d, act=%d", exp, act))
}

func assertFloat32Equals(t *testing.T, exp, act float32, title string) bool {
	return assert(t, math.Abs(float64(exp-act)) < TOLERANCE, title+fmt.Sprintf(": exp=%f, act=%f", exp, act))
}

func float32SliceToString(li []float32) string {
	// use a byte-buffer for efficiency, similar to using StringBuilder in Java.
	res := bytes.NewBufferString("[")
	lenli := len(li)

	// special case: first item has no preceding comma
	if lenli > 0 {
		res.WriteString(fmt.Sprintf(" %f", li[0]))
	}

	for i := 1; i < lenli; i++ {
		res.WriteString(fmt.Sprintf(", %f", li[i]))
	}

	res.WriteString(" ]")
	return res.String()
}

func assertFloat32SliceEquals(t *testing.T, exp, act []float32, title string) bool {
	// check length
	lenExp, lenAct := len(exp), len(act)
	wasEqual := assertIntEquals(t, lenExp, lenAct, title+"-length")
	if !wasEqual {
		return false
	}

	// check each element
	allPassed := true
	for i := 0; i < lenExp; i++ {
		if !assertFloat32Equals(t, exp[i], act[i], title+fmt.Sprintf("[index=%d]", i)) {
			allPassed = false
		}
	}

	return allPassed
}

func assertConvKernelEquals(t *testing.T, expKernel []float32, expRadius int, act *ConvKernel, title string) bool {
	// check radius and kernel
	return assertIntEquals(t, expRadius, act.Radius, title+".Radius") &&
		assertFloat32SliceEquals(t, expKernel, act.Kernel, title+".Kernel")
}

// helper
func testEmptyKernelHasCorrectArea(t *testing.T, radius int) {

	msg := "[radius=" + string(radius) + "]."
	area, diameter, kernel := emptyKernel(radius)

	// diameter should be radius*2 + 1
	expDiameter := 2*radius + 1
	assertIntEquals(t, diameter, expDiameter, msg+"diameter")

	// area should be diameter squared
	expArea := expDiameter * expDiameter
	assertIntEquals(t, area, expArea, msg+"area")

	// kernel should be all zeros
	expKernel := make([]float32, expArea)
	assertFloat32SliceEquals(t, kernel, expKernel, msg+"kernel")
}

func TestEmptyKernelHasCorrectArea(t *testing.T) {
	// iterate over radii : 1 -> 5
	for radius := 1; radius <= 5; radius++ {
		testEmptyKernelHasCorrectArea(t, radius)
	}

	// try larger radii: 10, 20 and 40.
	for radius := 10; radius <= 40; radius *= 2 {
		testEmptyKernelHasCorrectArea(t, radius)
	}
}

func TestMeanFilterKernelOfRadiusZero(t *testing.T) {
	expRadius := 0
	expKernel := []float32{float32(1)}
	actKernel := MeanFilterKernel(expRadius)
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "MeanFilterKernel[radius=0]")
}

func TestMeanFilterKernelOfRadiusOne(t *testing.T) {
	expRadius := 1
	inv := float32(1.0 / 9.0)
	expKernel := []float32{inv, inv, inv, inv, inv, inv, inv, inv, inv}
	actKernel := MeanFilterKernel(expRadius)
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "MeanFilterKernel[radius=1]")
}

func TestMeanFilterKernelOfRadiusTwo(t *testing.T) {
	expRadius := 2
	inv := float32(1.0 / 25.0)
	expKernel := []float32{
		inv, inv, inv, inv, inv,
		inv, inv, inv, inv, inv,
		inv, inv, inv, inv, inv,
		inv, inv, inv, inv, inv,
		inv, inv, inv, inv, inv}
	actKernel := MeanFilterKernel(expRadius)
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "MeanFilterKernel[radius=2]")
}

func TestLaplaceWithoutDiagonal(t *testing.T) {
	expRadius := 1
	expKernel := []float32{
		0, 1, 0,
		1, -4, 1,
		0, 1, 0}
	actKernel := LaplaceWithoutDiagonal()
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "LaplaceWithoutDiagonal")
}

func TestLaplaceWithDiagonal(t *testing.T) {
	expRadius := 1
	expKernel := []float32{
		0.5, 1, 0.5,
		1, -6, 1,
		0.5, 1, 0.5}
	actKernel := LaplaceWithDiagonal()
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "LaplaceWithDiagonal")
}

func TestLaplaceSpherical(t *testing.T) {
	expRadius := 1
	expKernel := []float32{
		1, 1, 1,
		1, -8, 1,
		1, 1, 1}
	actKernel := LaplaceSpherical()
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "LaplaceSpherical")
}

// unit Gaussian filter kernel
func TestGaussianFilterKernelOfRadiusZero(t *testing.T) {

	expRadius := 0
	expKernel := []float32{1}
	actKernel := GaussianFilterKernel(expRadius, 1.0)
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "GaussianKernel[radius=0,sigma=1.0]")
}

// for Gaussian kernel with radius=3, sigma=0.84089642
// source of expKernel: wikipedia [http://en.wikipedia.org/wiki/Gaussian_blur#Sample_Gaussian_matrix]
func TestGaussianFilterKernelOfRadiusThree(t *testing.T) {

	sigma := 0.84089642
	variance := sigma * sigma
	expRadius := 3
	expKernel := []float32{
		0.00000067, 0.00002292, 0.00019117, 0.00038771, 0.00019117, 0.00002292, 0.00000067,
		0.00002292, 0.00078633, 0.00655965, 0.01330373, 0.00655965, 0.00078633, 0.00002292,
		0.00019117, 0.00655965, 0.05472157, 0.11098164, 0.05472157, 0.00655965, 0.00019117,
		0.00038771, 0.01330373, 0.11098164, 0.22508352, 0.11098164, 0.01330373, 0.00038771,
		0.00019117, 0.00655965, 0.05472157, 0.11098164, 0.05472157, 0.00655965, 0.00019117,
		0.00002292, 0.00078633, 0.00655965, 0.01330373, 0.00655965, 0.00078633, 0.00002292,
		0.00000067, 0.00002292, 0.00019117, 0.00038771, 0.00019117, 0.00002292, 0.00000067}
	actKernel := GaussianFilterKernel(expRadius, variance)
	assertConvKernelEquals(t, expKernel, expRadius, actKernel, "GaussianKernel[radius=3,sigma=0.84]")
}
