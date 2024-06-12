package pascal

import (
	"testing"

	. "github.com/stevegt/goadapt"
)

// test the PascalsTriangle function
func TestPascalsTriangle(t *testing.T) {

	// PascalsTriangle, given an integer n, returns the Nth row of
	// Pascal's Triangle as a slice of integers.
	// The first row is 1, the second row is 1,1, the third row is 1,2,1, etc.
	// The Nth row is the coefficients of the expansion of (a+b)^N.
	// The row numbers are zero-based, so the first row is row 0.
	// The coefficients are the binomial coefficients, and are the Nth row of Pascal's Triangle.
	// The coefficients are calculated using the formula:
	// C(n,k) = n! / (k! * (n-k)!)
	// where n is the row number, and k is the column number.

	// call the PascalsTriangle function with n=0 (the first row)
	result := PascalsTriangle(0)
	// the result should be [1]
	Tassert(t, len(result) == 1, "Expected 1 element in the result")
	Tassert(t, result[0] == 1, "Expected 1 in the result")
}
