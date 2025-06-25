package float

import (
	"fmt"
	"math/big"
	"strings"
)

type Decimal struct {
	integer   *big.Int
	decimal   *big.Int
	decPlaces int
}

func NewDecimal(s string) (*Decimal, error) {
	s = strings.ReplaceAll(s, ",", "")
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("invalid decimal: %s", s)
	}

	isNeg := r.Sign() < 0
	if isNeg {
		r.Abs(r)
	}

	// 这里固定保留40位小数，后续去掉尾零
	decimalStr := r.FloatString(40)
	parts := strings.Split(decimalStr, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) == 2 {
		decPart = strings.TrimRight(parts[1], "0")
	}

	decPlaces := len(decPart)
	fullStr := intPart + decPart
	val := new(big.Int)
	if _, ok := val.SetString(fullStr, 10); !ok {
		return nil, fmt.Errorf("invalid combined number: %s", fullStr)
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decPlaces)), nil)
	integer := new(big.Int)
	decimal := new(big.Int)
	integer.DivMod(val, divisor, decimal)

	if isNeg {
		integer.Neg(integer)
		if decimal.Sign() != 0 {
			decimal.Neg(decimal)
		}
	}

	return &Decimal{
		integer:   integer,
		decimal:   decimal,
		decPlaces: decPlaces,
	}, nil
}

func (d *Decimal) toScaledInt(scale int) *big.Int {
	scaleFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.decPlaces)), nil)
	base := new(big.Int).Mul(d.integer, scaleFactor)
	base.Add(base, d.decimal)

	diff := scale - d.decPlaces
	if diff > 0 {
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil)
		base.Mul(base, multiplier)
	}
	return base
}

func fromScaledInt(val *big.Int, decPlaces int) *Decimal {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decPlaces)), nil)
	intPart := new(big.Int)
	decPart := new(big.Int)
	intPart.DivMod(val, divisor, decPart)

	// 如果val为负数且decPart非零，需要调整
	if val.Sign() < 0 && decPart.Sign() != 0 {
		// intPart向上调整 +1
		intPart.Add(intPart, big.NewInt(1))
		// decPart改为负值： decPart = decPart - divisor
		decPart.Sub(decPart, divisor)
	}

	return &Decimal{
		integer:   intPart,
		decimal:   decPart,
		decPlaces: decPlaces,
	}
}

func (d *Decimal) Add(o *Decimal) *Decimal {
	maxDec := max(d.decPlaces, o.decPlaces)
	a := d.toScaledInt(maxDec)
	b := o.toScaledInt(maxDec)
	sum := new(big.Int).Add(a, b)
	return fromScaledInt(sum, maxDec)
}

func (d *Decimal) Mul(o *Decimal) *Decimal {
	a := d.toScaledInt(d.decPlaces)
	b := o.toScaledInt(o.decPlaces)
	product := new(big.Int).Mul(a, b)

	return fromScaledInt(product, d.decPlaces+o.decPlaces)
}

// String 返回自动去尾0的字符串
func (d *Decimal) String() string {
	intStr := d.integer.String()
	decAbs := new(big.Int).Abs(d.decimal) // 取小数部分绝对值字符串
	decStr := decAbs.String()

	for len(decStr) < d.decPlaces {
		decStr = "0" + decStr
	}
	// decStr = strings.TrimRight(decStr, "0")

	if decStr == "" {
		return intStr
	}
	return intStr + "." + decStr
}

// StringFixed 固定小数位数，补0不去尾，用于显示和测试
func (d *Decimal) StringFixed(n int) string {
	val := d.toScaledInt(n)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	intPart := new(big.Int)
	decPart := new(big.Int)
	intPart.DivMod(val, divisor, decPart)
	intStr := intPart.String()
	decStr := decPart.String()
	for len(decStr) < n {
		decStr = "0" + decStr
	}
	return intStr + "." + decStr
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
