package lexorank

import (
	"fmt"
	"strings"
)

// Integer is an arbitrary-precision integer in any base.
// Magnitude is stored little-endian (index 0 = least significant digit).
type Integer struct {
	sys  NumeralSystem
	sign int   // -1, 0, or 1
	mag  []int // little-endian, each element in [0, base)
}

// NewInteger creates an Integer, trimming trailing zeros.
func NewInteger(sys NumeralSystem, sign int, mag []int) Integer {
	actual := len(mag)
	for actual > 0 && mag[actual-1] == 0 {
		actual--
	}
	if actual == 0 {
		return Zero(sys)
	}
	nmag := make([]int, actual)
	copy(nmag, mag[:actual])
	return Integer{sys: sys, sign: sign, mag: nmag}
}

// ParseInteger parses a string into an Integer.
func ParseInteger(s string, sys NumeralSystem) (Integer, error) {
	sign := 1
	str := s
	if len(str) > 0 {
		if str[0] == sys.PositiveChar() {
			str = str[1:]
		} else if str[0] == sys.NegativeChar() {
			str = str[1:]
			sign = -1
		}
	}
	mag := make([]int, len(str))
	strIdx := len(str) - 1
	for magIdx := 0; strIdx >= 0; magIdx++ {
		d, err := sys.ToDigit(str[strIdx])
		if err != nil {
			return Integer{}, fmt.Errorf("invalid digit at position %d: %w", strIdx, err)
		}
		mag[magIdx] = d
		strIdx--
	}
	return NewInteger(sys, sign, mag), nil
}

// Zero returns the zero Integer for the given system.
func Zero(sys NumeralSystem) Integer {
	return Integer{sys: sys, sign: 0, mag: []int{0}}
}

// One returns the one Integer for the given system.
func One(sys NumeralSystem) Integer {
	return NewInteger(sys, 1, []int{1})
}

func (i Integer) System() NumeralSystem { return i.sys }
func (i Integer) IsZero() bool          { return i.sign == 0 && len(i.mag) == 1 && i.mag[0] == 0 }
func (i Integer) IsOne() bool           { return i.sign == 1 && len(i.mag) == 1 && i.mag[0] == 1 }
func (i Integer) MagAt(index int) int   { return i.mag[index] }

func (i Integer) isOneish() bool {
	return len(i.mag) == 1 && i.mag[0] == 1
}

func (i Integer) checkSystem(other Integer) {
	if i.sys.Base() != other.sys.Base() {
		panic("expected numbers of same numeral system")
	}
}

func (i Integer) Add(other Integer) Integer {
	i.checkSystem(other)
	if i.IsZero() {
		return other
	}
	if other.IsZero() {
		return i
	}
	if i.sign != other.sign {
		if i.sign == -1 {
			pos := i.Negate()
			val := pos.Sub(other)
			return val.Negate()
		}
		pos := other.Negate()
		return i.Sub(pos)
	}
	result := addMag(i.sys, i.mag, other.mag)
	return NewInteger(i.sys, i.sign, result)
}

func (i Integer) Sub(other Integer) Integer {
	i.checkSystem(other)
	if i.IsZero() {
		return other.Negate()
	}
	if other.IsZero() {
		return i
	}
	if i.sign != other.sign {
		if i.sign == -1 {
			neg := i.Negate()
			sum := neg.Add(other)
			return sum.Negate()
		}
		neg := other.Negate()
		return i.Add(neg)
	}
	cmp := compareMag(i.mag, other.mag)
	if cmp == 0 {
		return Zero(i.sys)
	}
	if cmp < 0 {
		newSign := -1
		if i.sign == -1 {
			newSign = 1
		}
		return NewInteger(i.sys, newSign, subMag(i.sys, other.mag, i.mag))
	}
	newSign := 1
	if i.sign == -1 {
		newSign = -1
	}
	return NewInteger(i.sys, newSign, subMag(i.sys, i.mag, other.mag))
}

func (i Integer) Mul(other Integer) Integer {
	i.checkSystem(other)
	if i.IsZero() {
		return i
	}
	if other.IsZero() {
		return other
	}
	if i.isOneish() {
		if i.sign == other.sign {
			return NewInteger(i.sys, 1, other.mag)
		}
		return NewInteger(i.sys, -1, other.mag)
	}
	if other.isOneish() {
		if i.sign == other.sign {
			return NewInteger(i.sys, 1, i.mag)
		}
		return NewInteger(i.sys, -1, i.mag)
	}
	newMag := mulMag(i.sys, i.mag, other.mag)
	if i.sign == other.sign {
		return NewInteger(i.sys, 1, newMag)
	}
	return NewInteger(i.sys, -1, newMag)
}

func (i Integer) Negate() Integer {
	if i.IsZero() {
		return i
	}
	return NewInteger(i.sys, -i.sign, i.mag)
}

func (i Integer) ShiftLeft(times int) Integer {
	if times == 0 {
		return i
	}
	if times < 0 {
		return i.ShiftRight(-times)
	}
	nmag := make([]int, len(i.mag)+times)
	copy(nmag[times:], i.mag)
	return NewInteger(i.sys, i.sign, nmag)
}

func (i Integer) ShiftRight(times int) Integer {
	if len(i.mag)-times <= 0 {
		return Zero(i.sys)
	}
	nmag := make([]int, len(i.mag)-times)
	copy(nmag, i.mag[times:])
	return NewInteger(i.sys, i.sign, nmag)
}

func (i Integer) Complement() Integer {
	return i.ComplementDigits(len(i.mag))
}

func (i Integer) ComplementDigits(digits int) Integer {
	return NewInteger(i.sys, i.sign, complementMag(i.sys, i.mag, digits))
}

func (i Integer) CompareTo(other Integer) int {
	if i.sign == -1 {
		if other.sign == -1 {
			cmp := compareMag(i.mag, other.mag)
			if cmp == -1 {
				return 1
			}
			if cmp == 1 {
				return -1
			}
			return 0
		}
		return -1
	}
	if i.sign == 1 {
		if other.sign == 1 {
			return compareMag(i.mag, other.mag)
		}
		return 1
	}
	// i is zero
	if other.sign == -1 {
		return 1
	}
	if other.sign == 1 {
		return -1
	}
	return 0
}

func (i Integer) Equals(other Integer) bool {
	if i.sys.Base() != other.sys.Base() {
		return false
	}
	return i.CompareTo(other) == 0
}

func (i Integer) Format() string {
	if i.IsZero() {
		return string(rune(i.sys.ToChar(0)))
	}
	var sb strings.Builder
	for idx := len(i.mag) - 1; idx >= 0; idx-- {
		sb.WriteByte(i.sys.ToChar(i.mag[idx]))
	}
	if i.sign == -1 {
		result := sb.String()
		return string(rune(i.sys.NegativeChar())) + result
	}
	return sb.String()
}

func (i Integer) String() string {
	return i.Format()
}

// --- magnitude helpers ---

func addMag(sys NumeralSystem, l, r []int) []int {
	size := len(l)
	if len(r) > size {
		size = len(r)
	}
	result := make([]int, size)
	carry := 0
	base := sys.Base()
	for i := 0; i < size; i++ {
		lnum := 0
		if i < len(l) {
			lnum = l[i]
		}
		rnum := 0
		if i < len(r) {
			rnum = r[i]
		}
		sum := lnum + rnum + carry
		carry = 0
		for sum >= base {
			sum -= base
			carry++
		}
		result[i] = sum
	}
	return extendWithCarry(result, carry)
}

func extendWithCarry(mag []int, carry int) []int {
	if carry > 0 {
		ext := make([]int, len(mag)+1)
		copy(ext, mag)
		ext[len(ext)-1] = carry
		return ext
	}
	return mag
}

func subMag(sys NumeralSystem, l, r []int) []int {
	rComp := complementMag(sys, r, len(l))
	rSum := addMag(sys, l, rComp)
	rSum[len(rSum)-1] = 0
	return addMag(sys, rSum, []int{1})
}

func mulMag(sys NumeralSystem, l, r []int) []int {
	base := sys.Base()
	result := make([]int, len(l)+len(r))
	for li := 0; li < len(l); li++ {
		for ri := 0; ri < len(r); ri++ {
			idx := li + ri
			result[idx] += l[li] * r[ri]
			for result[idx] >= base {
				result[idx] -= base
				result[idx+1]++
			}
		}
	}
	// Final carry propagation sweep
	for i := 0; i < len(result)-1; i++ {
		for result[i] >= base {
			result[i] -= base
			result[i+1]++
		}
	}
	return result
}

func complementMag(sys NumeralSystem, mag []int, digits int) []int {
	if digits <= 0 {
		panic("expected at least 1 digit")
	}
	base := sys.Base()
	nmag := make([]int, digits)
	for i := range nmag {
		nmag[i] = base - 1
	}
	for i := 0; i < len(mag); i++ {
		nmag[i] = base - 1 - mag[i]
	}
	return nmag
}

func compareMag(l, r []int) int {
	if len(l) < len(r) {
		return -1
	}
	if len(l) > len(r) {
		return 1
	}
	for i := len(l) - 1; i >= 0; i-- {
		if l[i] < r[i] {
			return -1
		}
		if l[i] > r[i] {
			return 1
		}
	}
	return 0
}

// --- Decimal ---

// Decimal is an arbitrary-precision decimal built on Integer.
type Decimal struct {
	mag Integer
	sig int // number of digits after the radix point
}

// Half returns 0.5 in the given numeral system.
func Half(sys NumeralSystem) Decimal {
	mid := sys.Base() / 2
	return MakeDecimal(NewInteger(sys, 1, []int{mid}), 1)
}

// ParseDecimal parses a string with an optional radix point into a Decimal.
func ParseDecimal(s string, sys NumeralSystem) (Decimal, error) {
	radix := sys.RadixPointChar()
	partialIdx := strings.IndexByte(s, radix)
	if strings.LastIndexByte(s, radix) != partialIdx {
		return Decimal{}, fmt.Errorf("more than one radix point %q in %q", radix, s)
	}
	if partialIdx < 0 {
		integer, err := ParseInteger(s, sys)
		if err != nil {
			return Decimal{}, err
		}
		return MakeDecimal(integer, 0), nil
	}
	intStr := s[:partialIdx] + s[partialIdx+1:]
	sig := len(s) - 1 - partialIdx
	integer, err := ParseInteger(intStr, sys)
	if err != nil {
		return Decimal{}, err
	}
	return MakeDecimal(integer, sig), nil
}

// DecimalFrom wraps an Integer with scale 0.
func DecimalFrom(integer Integer) Decimal {
	return MakeDecimal(integer, 0)
}

// MakeDecimal creates a Decimal, trimming trailing fractional zeros.
func MakeDecimal(integer Integer, sig int) Decimal {
	if integer.IsZero() {
		return Decimal{mag: integer, sig: 0}
	}
	zeroCount := 0
	for i := 0; i < sig && integer.MagAt(i) == 0; i++ {
		zeroCount++
	}
	newInt := integer.ShiftRight(zeroCount)
	newSig := sig - zeroCount
	return Decimal{mag: newInt, sig: newSig}
}

func (d Decimal) System() NumeralSystem { return d.mag.System() }
func (d Decimal) Scale() int            { return d.sig }

func (d Decimal) Add(other Decimal) Decimal {
	tmag := d.mag
	tsig := d.sig
	omag := other.mag
	osig := other.sig
	for tsig < osig {
		tmag = tmag.ShiftLeft(1)
		tsig++
	}
	for tsig > osig {
		omag = omag.ShiftLeft(1)
		osig++
	}
	return MakeDecimal(tmag.Add(omag), tsig)
}

func (d Decimal) Sub(other Decimal) Decimal {
	tmag := d.mag
	tsig := d.sig
	omag := other.mag
	osig := other.sig
	for tsig < osig {
		tmag = tmag.ShiftLeft(1)
		tsig++
	}
	for tsig > osig {
		omag = omag.ShiftLeft(1)
		osig++
	}
	return MakeDecimal(tmag.Sub(omag), tsig)
}

func (d Decimal) Mul(other Decimal) Decimal {
	return MakeDecimal(d.mag.Mul(other.mag), d.sig+other.sig)
}

func (d Decimal) Floor() Integer {
	return d.mag.ShiftRight(d.sig)
}

func (d Decimal) Ceil() Integer {
	if d.IsExact() {
		return d.Floor()
	}
	f := d.Floor()
	return f.Add(One(f.System()))
}

func (d Decimal) IsExact() bool {
	if d.sig == 0 {
		return true
	}
	for i := 0; i < d.sig; i++ {
		if d.mag.MagAt(i) != 0 {
			return false
		}
	}
	return true
}

func (d Decimal) SetScale(nsig int, ceiling bool) Decimal {
	if nsig >= d.sig {
		return d
	}
	if nsig < 0 {
		nsig = 0
	}
	diff := d.sig - nsig
	nmag := d.mag.ShiftRight(diff)
	if ceiling {
		nmag = nmag.Add(One(nmag.System()))
	}
	return MakeDecimal(nmag, nsig)
}

func (d Decimal) CompareTo(other Decimal) int {
	tMag := d.mag
	oMag := other.mag
	if d.sig > other.sig {
		oMag = oMag.ShiftLeft(d.sig - other.sig)
	} else if d.sig < other.sig {
		tMag = tMag.ShiftLeft(other.sig - d.sig)
	}
	return tMag.CompareTo(oMag)
}

func (d Decimal) Equals(other Decimal) bool {
	return d.mag.Equals(other.mag) && d.sig == other.sig
}

func (d Decimal) Format() string {
	intStr := d.mag.Format()
	if d.sig == 0 {
		return intStr
	}
	head := byte(0)
	specialHead := false
	if len(intStr) > 0 {
		h := intStr[0]
		if h == d.mag.System().PositiveChar() || h == d.mag.System().NegativeChar() {
			head = h
			specialHead = true
			intStr = intStr[1:]
		}
	}
	for len(intStr) < d.sig+1 {
		intStr = string(rune(d.mag.System().ToChar(0))) + intStr
	}
	insertPos := len(intStr) - d.sig
	intStr = intStr[:insertPos] + string(rune(d.mag.System().RadixPointChar())) + intStr[insertPos:]
	if specialHead {
		intStr = string(rune(head)) + intStr
	}
	return intStr
}

func (d Decimal) String() string {
	return d.Format()
}
