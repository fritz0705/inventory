// Implementation of the metric unit system, also knows as the SI unit system
// or International System of Units

package si

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Prefix represents the exponent to 10 of a SI prefix
type Prefix int

const (
	Yotta Prefix = 24
	Zetta        = 21
	Exa          = 18
	Peta         = 15
	Tera         = 12
	Giga         = 9
	Mega         = 6
	Kilo         = 3
	Hecto        = 2
	Deca         = 1
	None         = 0
	Deci         = -1
	Centi        = -2
	Milli        = -3
	Micro        = -6
	Nano         = -9
	Pico         = -12
	Femto        = -15
)

func (p Prefix) String() string {
	switch p {
	case Yotta:
		return "Y"
	case Zetta:
		return "Z"
	case Exa:
		return "E"
	case Peta:
		return "P"
	case Tera:
		return "T"
	case Giga:
		return "G"
	case Mega:
		return "M"
	case Kilo:
		return "k"
	case Hecto:
		return "h"
	case Deca:
		return "da"
	case Deci:
		return "d"
	case Centi:
		return "c"
	case Milli:
		return "m"
	case Micro:
		return "μ"
	case Nano:
		return "n"
	case Pico:
		return "p"
	case Femto:
		return "f"
	}
	return ""
}

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
	Hecto,
	Deca,
	Deci,
	Centi,
	Milli,
	Micro,
	Nano,
	Pico,
	Femto,
}

var PrefixMapping map[string]Prefix

// A Number is a float64 combined with a Exponent, it is similar to a decimal
// floating number with the restriction that it imposes additional inaccuracy
type Number struct {
	Significand float64
	Exponent    Prefix
}

// New creates a Number object from a float64
func New(val float64) Number {
	return Number{
		Significand: val,
		Exponent:    None,
	}
}

// Parse converts a strings to a Number object.
//
// The string s has to begin with a valid floating point number. Then, Parse
// looks for a prefix. You can use one space (' ') between the floating point
// number and the prefix string. Please note that the space is required when
// the input contains the "E" or "p" prefix.
func Parse(s string) (num Number, err error) {
	if strings.ContainsRune(s, ' ') {
		numberPrefix := strings.SplitN(s, " ", 2)
		if len(numberPrefix) == 2 {
			num.Exponent = PrefixMapping[numberPrefix[1]]
			s = numberPrefix[0]
		}
	}
	num.Significand, err = strconv.ParseFloat(s, 64)
	if err == nil {
		return
	}
	var p string
	_, err = fmt.Sscanf(s, "%f%s", &num.Significand, &p)
	if err != nil {
		return
	}
	num.Exponent = PrefixMapping[p]
	return
}

// Value returns the real value of a Number object as float64
func (n Number) Value() float64 {
	return n.Significand * math.Pow10(int(n.Exponent))
}

// String returns a string representation of a Number
func (n Number) String() string {
	var space string
	switch n.Exponent {
	case Pico:
		fallthrough
	case Exa:
		space = " "
	}
	return strconv.FormatFloat(n.Significand, 'f', -1, 64) + space + n.Exponent.String()
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
				Exponent:    prefix,
			}
		}
	}
	return n
}

func init() {
	PrefixMapping = make(map[string]Prefix)
	for _, prefix := range Prefixes {
		PrefixMapping[prefix.String()] = prefix
	}
	PrefixMapping["u"] = Micro
}
