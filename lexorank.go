package lexorank

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	sys36 NumeralSystem

	zeroDecimal   Decimal
	oneDecimal    Decimal
	eightDecimal  Decimal
	minDecimal    Decimal
	maxDecimal    Decimal
	initialMinDec Decimal
	initialMaxDec Decimal
)

func mustParseDecimal(s string, sys NumeralSystem) Decimal {
	d, err := ParseDecimal(s, sys)
	if err != nil {
		panic(err)
	}
	return d
}

func init() {
	sys36 = System36
	zeroDecimal = mustParseDecimal("0", sys36)
	oneDecimal = mustParseDecimal("1", sys36)
	eightDecimal = mustParseDecimal("8", sys36)
	minDecimal = zeroDecimal
	maxDecimal = mustParseDecimal("1000000", sys36).Sub(oneDecimal)
	initialMinDec = mustParseDecimal("100000", sys36)
	initialMaxDec = mustParseDecimal(string(rune(sys36.ToChar(sys36.Base()-2)))+"00000", sys36)
}

// --- Bucket ---

// Bucket represents one of the 3 rank buckets (0, 1, 2).
type Bucket int

const (
	Bucket0 Bucket = 0
	Bucket1 Bucket = 1
	Bucket2 Bucket = 2
)

// MaxBucket returns Bucket2.
func MaxBucket() Bucket { return Bucket2 }

// BucketFrom parses a bucket from a string ("0", "1", or "2").
func BucketFrom(s string) (Bucket, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid bucket %q: %w", s, err)
	}
	b := Bucket(n)
	if b < Bucket0 || b > Bucket2 {
		return 0, fmt.Errorf("bucket %d out of range [0,2]", n)
	}
	return b, nil
}

// ResolveBucket resolves a bucket by numeric ID.
func ResolveBucket(bucketID int) (Bucket, error) {
	b := Bucket(bucketID)
	if b < Bucket0 || b > Bucket2 {
		return 0, fmt.Errorf("bucket %d out of range [0,2]", bucketID)
	}
	return b, nil
}

func (b Bucket) Format() string { return string(rune(sys36.ToChar(int(b)))) }
func (b Bucket) String() string { return b.Format() }

func (b Bucket) Next() Bucket {
	switch b {
	case Bucket0:
		return Bucket1
	case Bucket1:
		return Bucket2
	default:
		return Bucket0
	}
}

func (b Bucket) Prev() Bucket {
	switch b {
	case Bucket0:
		return Bucket2
	case Bucket1:
		return Bucket0
	default:
		return Bucket1
	}
}

// --- Rank ---

// Rank is a lexicographically sortable rank string.
type Rank struct {
	bucket  Bucket
	decimal Decimal
}

// Min returns the minimum rank.
func Min() Rank {
	return RankFrom(Bucket0, minDecimal)
}

// Middle returns the midpoint between min and max.
func Middle() Rank {
	minRank := Min()
	mid, _ := minRank.Between(Max())
	return mid
}

// Max returns the maximum rank for the given bucket (default Bucket0).
func Max(bucket ...Bucket) Rank {
	b := Bucket0
	if len(bucket) > 0 {
		b = bucket[0]
	}
	return RankFrom(b, maxDecimal)
}

// Initial returns the initial rank for a bucket.
func Initial(bucket Bucket) Rank {
	if bucket == Bucket0 {
		return RankFrom(bucket, initialMinDec)
	}
	return RankFrom(bucket, initialMaxDec)
}

// ParseRank parses a rank string like "0|000000:".
func ParseRank(s string) (Rank, error) {
	parts := strings.SplitN(s, "|", 2)
	if len(parts) != 2 {
		return Rank{}, fmt.Errorf("invalid rank format %q: expected 'bucket|decimal'", s)
	}
	bucket, err := BucketFrom(parts[0])
	if err != nil {
		return Rank{}, err
	}
	decimal, err := ParseDecimal(parts[1], sys36)
	if err != nil {
		return Rank{}, fmt.Errorf("invalid decimal in rank %q: %w", s, err)
	}
	return RankFrom(bucket, decimal), nil
}

// RankFrom creates a Rank from a bucket and decimal.
func RankFrom(bucket Bucket, decimal Decimal) Rank {
	return Rank{
		bucket:  bucket,
		decimal: decimal,
	}
}

func (r Rank) String() string      { return r.Format() }
func (r Rank) Format() string      { return r.bucket.Format() + "|" + formatDecimal(r.decimal) }
func (r Rank) GetBucket() Bucket   { return r.bucket }
func (r Rank) GetDecimal() Decimal { return r.decimal }

func (r Rank) IsMin() bool { return r.decimal.Equals(minDecimal) }
func (r Rank) IsMax() bool { return r.decimal.Equals(maxDecimal) }

func (r Rank) Equals(other Rank) bool { return r.Format() == other.Format() }

func (r Rank) CompareTo(other Rank) int {
	a := r.Format()
	b := other.Format()
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

func (r Rank) GenPrev() Rank {
	if r.IsMax() {
		return RankFrom(r.bucket, initialMaxDec)
	}
	floorInt := r.decimal.Floor()
	floorDec := DecimalFrom(floorInt)
	nextDec := floorDec.Sub(eightDecimal)
	if nextDec.CompareTo(minDecimal) <= 0 {
		nextDec = betweenDecimals(minDecimal, r.decimal)
	}
	return RankFrom(r.bucket, nextDec)
}

func (r Rank) GenNext() Rank {
	if r.IsMin() {
		return RankFrom(r.bucket, initialMinDec)
	}
	ceilInt := r.decimal.Ceil()
	ceilDec := DecimalFrom(ceilInt)
	nextDec := ceilDec.Add(eightDecimal)
	if nextDec.CompareTo(maxDecimal) >= 0 {
		nextDec = betweenDecimals(r.decimal, maxDecimal)
	}
	return RankFrom(r.bucket, nextDec)
}

// Between returns the midpoint rank between r and other.
// Returns an error if the ranks are in different buckets or have the same decimal.
func (r Rank) Between(other Rank) (Rank, error) {
	if r.bucket != other.bucket {
		return Rank{}, fmt.Errorf("between works only within the same bucket")
	}
	cmp := r.decimal.CompareTo(other.decimal)
	if cmp > 0 {
		return RankFrom(r.bucket, betweenDecimals(other.decimal, r.decimal)), nil
	}
	if cmp == 0 {
		return Rank{}, fmt.Errorf("cannot rank between issues with same rank")
	}
	return RankFrom(r.bucket, betweenDecimals(r.decimal, other.decimal)), nil
}

func (r Rank) InNextBucket() Rank {
	return RankFrom(r.bucket.Next(), r.decimal)
}

func (r Rank) InPrevBucket() Rank {
	return RankFrom(r.bucket.Prev(), r.decimal)
}

// --- between algorithm ---

func betweenDecimals(oLeft, oRight Decimal) Decimal {
	left := oLeft
	right := oRight

	if oLeft.Scale() < oRight.Scale() {
		nLeft := oRight.SetScale(oLeft.Scale(), false)
		if oLeft.CompareTo(nLeft) >= 0 {
			return midDecimal(oLeft, oRight)
		}
		right = nLeft
	}

	if oLeft.Scale() > right.Scale() {
		nLeft := oLeft.SetScale(right.Scale(), true)
		if nLeft.CompareTo(right) >= 0 {
			return midDecimal(oLeft, oRight)
		}
		left = nLeft
	}

	for scale := left.Scale(); scale > 0; {
		nScale := scale - 1
		nLeft := left.SetScale(nScale, true)
		nRight := right.SetScale(nScale, false)
		cmp := nLeft.CompareTo(nRight)
		if cmp == 0 {
			return checkMid(oLeft, oRight, nLeft)
		}
		if cmp > 0 {
			break
		}
		scale = nScale
		left = nLeft
		right = nRight
	}

	mid := middleInternal(oLeft, oRight, left, right)

	for mScale := mid.Scale(); mScale > 0; {
		nScale := mScale - 1
		nMid := mid.SetScale(nScale, false)
		if oLeft.CompareTo(nMid) >= 0 || nMid.CompareTo(oRight) >= 0 {
			break
		}
		mid = nMid
		mScale = nScale
	}

	return mid
}

func midDecimal(left, right Decimal) Decimal {
	sum := left.Add(right)
	mid := sum.Mul(Half(left.System()))
	scale := left.Scale()
	if right.Scale() > scale {
		scale = right.Scale()
	}
	if mid.Scale() > scale {
		roundDown := mid.SetScale(scale, false)
		if roundDown.CompareTo(left) > 0 {
			return roundDown
		}
		roundUp := mid.SetScale(scale, true)
		if roundUp.CompareTo(right) < 0 {
			return roundUp
		}
	}
	return mid
}

func checkMid(lbound, rbound, mid Decimal) Decimal {
	if lbound.CompareTo(mid) >= 0 {
		return midDecimal(lbound, rbound)
	}
	if mid.CompareTo(rbound) >= 0 {
		return midDecimal(lbound, rbound)
	}
	return mid
}

func middleInternal(lbound, rbound, left, right Decimal) Decimal {
	mid := midDecimal(left, right)
	return checkMid(lbound, rbound, mid)
}

const rankDecimalDigits = 6

func formatDecimal(decimal Decimal) string {
	formatVal := decimal.Format()
	radixChar := decimal.System().RadixPointChar()
	zeroChar := decimal.System().ToChar(0)

	partialIdx := strings.IndexByte(formatVal, radixChar)
	if partialIdx < 0 {
		partialIdx = len(formatVal)
		formatVal = formatVal + string(radixChar)
	}

	if partialIdx < rankDecimalDigits {
		formatVal = strings.Repeat(string(rune(zeroChar)), rankDecimalDigits-partialIdx) + formatVal
	}

	formatVal = strings.TrimRight(formatVal, string(rune(zeroChar)))

	return formatVal
}
