package lexorank

import "fmt"

// NumeralSystem defines the interface for a base-N numeral system.
type NumeralSystem interface {
	Base() int
	PositiveChar() byte
	NegativeChar() byte
	RadixPointChar() byte
	ToDigit(ch byte) (int, error)
	ToChar(digit int) byte
}

const (
	digits10 = "0123456789"
	digits36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	digits64 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ^_abcdefghijklmnopqrstuvwxyz"
)

var (
	System10 NumeralSystem = numeralSystem10{}
	System36 NumeralSystem = numeralSystem36{}
	System64 NumeralSystem = numeralSystem64{}
)

// --- base 10 ---

type numeralSystem10 struct{}

func (numeralSystem10) Base() int            { return 10 }
func (numeralSystem10) PositiveChar() byte   { return '+' }
func (numeralSystem10) NegativeChar() byte   { return '-' }
func (numeralSystem10) RadixPointChar() byte { return '.' }

func (numeralSystem10) ToDigit(ch byte) (int, error) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), nil
	}
	return 0, fmt.Errorf("not valid digit: %c", ch)
}

func (numeralSystem10) ToChar(digit int) byte {
	return digits10[digit]
}

// --- base 36 ---

type numeralSystem36 struct{}

func (numeralSystem36) Base() int            { return 36 }
func (numeralSystem36) PositiveChar() byte   { return '+' }
func (numeralSystem36) NegativeChar() byte   { return '-' }
func (numeralSystem36) RadixPointChar() byte { return ':' }

func (numeralSystem36) ToDigit(ch byte) (int, error) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), nil
	}
	if ch >= 'a' && ch <= 'z' {
		return int(ch-'a') + 10, nil
	}
	return 0, fmt.Errorf("not valid digit: %c", ch)
}

func (numeralSystem36) ToChar(digit int) byte {
	return digits36[digit]
}

// --- base 64 ---

type numeralSystem64 struct{}

func (numeralSystem64) Base() int            { return 64 }
func (numeralSystem64) PositiveChar() byte   { return '+' }
func (numeralSystem64) NegativeChar() byte   { return '-' }
func (numeralSystem64) RadixPointChar() byte { return ':' }

func (numeralSystem64) ToDigit(ch byte) (int, error) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), nil
	}
	if ch >= 'A' && ch <= 'Z' {
		return int(ch-'A') + 10, nil
	}
	if ch == '^' {
		return 36, nil
	}
	if ch == '_' {
		return 37, nil
	}
	if ch >= 'a' && ch <= 'z' {
		return int(ch-'a') + 38, nil
	}
	return 0, fmt.Errorf("not valid digit: %c", ch)
}

func (numeralSystem64) ToChar(digit int) byte {
	return digits64[digit]
}
