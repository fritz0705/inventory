// Implementation of the metric unit system, also knows as the SI unit system
// or International System of Units

package si

import (
	"fmt"
	"math"
	"strconv"
)

// Prefix represents the exponent to 10 of a SI prefix
type Prefix int

const (
	Yotta Prefix = 24
	Zetta = 21
	Exa   = 18
	Peta  = 15
	Tera  = 12
	Giga  = 9
	Mega  = 6
	Kilo  = 3
	None = 0
	Milli = -3
	Micro = -6
	Nano  = -9
	Pico  = -12
	Femto = -15
)

// Prefixes is a list of common SI prefixes
var Prefixes = []Prefix{
	Yotta,
	Zetta,
	Exa,
	Peta,
	Tera,
	Giga,
	Mega,
	Kilo,
	None,
	Milli,
	Micro,
	Nano,
	Pico,
	Femto,
}

// Symbols contains a mapping from SI prefixes to their symbol
var Symbols = map[Prefix]string{
	Yotta: "Y",
	Zetta: "Z",
	Exa: "E",
	Peta: "P",
	Tera: "T",
	Giga: "G",
	Mega: "M",
	Kilo: "k",
	None: "",
	Milli: "m",
	Micro: "μ",
	Nano: "n",
	Pico: "p",
	Femto: "f",
}

// A Number is a float64 combined with a Exponent, it is similar to a decimal
// floating number with the restriction that it imposes additional inaccuracy
type Number struct {
	Significand float64
	Exponent Prefix
}

// New creates a Number object from a float64
func New(val float64) Number {
	return Number{
		Significand: val,
		Exponent: None,
	}
}

// Value returns the real value of a Number object as float64
func (n Number) Value() float64 {
	return n.Significand * math.Pow10(int(n.Exponent))
}

// Canon tries to find the best matching SI prefix for a Value and returns a
// new Number object holding that
func (n Number) Canon() Number {
	val := n.Value()
	for _, prefix := range Prefixes {
		sig := val / math.Pow10(int(prefix))
		if sig >= 1.0 && sig <= 1000 {
			return Number{
				Significand: sig,
				Exponent: prefix,
			}
		}
	}
	return n
}

// String returns a string representation of a Number
func (n Number) String() string {
	return strconv.FormatFloat(n.Significand, 'f', -1, 64) + Symbols[n.Exponent]
}

func findPrefixForSymbol(sym string) Prefix {
	for p, s := range Symbols {
		if s == sym {
			return p
		}
	}

	return None
}

// Parse parses a string representation, which consists of a float string and
// an optional unit symbol.
func Parse(s string) (num Number, err error) {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		num = Number{
			Significand: f,
			Exponent: None,
		}
		return
	}

	var symbol string
	_, err = fmt.Sscanf(s, "%f%s", &num.Significand, &symbol)

	num.Exponent = findPrefixForSymbol(symbol)
	return
}
