package utils

import (
	"fmt"

	"github.com/shopspring/decimal"
	"testing"
)

func TestFormatPrice(t *testing.T) {
	tests := []string{
		"0.043549549",    // → 0.04354
		"0.0043549549",   // → 0.004354
		"0.00043549549",  // → 0.0004354
		"0.000043549549", // → 0.0{4}4354
		"2.00003456",     // → 2.0{4}3456
		"12.00003456",    // → 12.0{4}3456
		"10000.00003456", // → 10000.0{4}3456
		"0.000000123456", // → 0.0{6}1234
		"0.000012300045", // → 0.0{4}1230
		"123.000456789",  // → 123.0004
		"21.00000000000000000000000",
		"0.00000000000000001",
		"0.000000000000034200",
		"0",
	}

	tests2 := []struct {
		Amount   string
		Decimals int32
	}{
		{"0.0000", 9},
	}

	for _, test := range tests {
		fmt.Printf("Input: %s -> Output: %s\n", test, FormatPrice(test))
	}

	for _, test := range tests2 {
		fmt.Printf("Input: %s -> Output: %s\n", test.Amount, FormatAmountWithDecimals(test.Amount, test.Decimals))
	}

	ss := "123.0045"
	a := decimal.RequireFromString(ss)
	fmt.Println(a.Truncate(3).String())
}
