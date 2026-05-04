package lexorank

import (
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

func init() {
	sys36 = System36
	zeroDecimal = ParseDecimal("0", sys36)
	oneDecimal = ParseDecimal("1", sys36)
	eightDecimal = ParseDecimal("8", sys36)
	minDecimal = zeroDecimal
	maxDecimal = ParseDecimal("1000000", sys36).Sub(oneDecimal)
	initialMinDec = ParseDecimal("100000", sys36)
	initialMaxDec = ParseDecimal(string(rune(sys36.ToChar(sys36.Base()-2)))+"00000", sys36)
}

// --- Bucket ---

// Bucket represents one of the 3 rank buckets (0, 1, 2).
type Bucket struct {
	id int
}

var (
	Bucket0 = Bucket{id: 0}
	Bucket1 = Bucket{id: 1}
	Bucket2 = Bucket{id: 2}
)

// MaxBucket returns Bucket2.
func MaxBucket() Bucket { return Bucket2 }

// BucketFrom parses a bucket from a string ("0", "1", or "2").
func BucketFrom(s string) Bucket {
	switch s {
	case "0":
		return Bucket0
	case "1":
		return Bucket1
	case "2":
		return Bucket2
	default:
		panic("unknown bucket: " + s)
	}
}

// ResolveBucket resolves a bucket by numeric ID.
func ResolveBucket(bucketID int) Bucket {
	switch bucketID {
	case 0:
		return Bucket0
	case 1:
		return Bucket1
	case 2:
		return Bucket2
	default:
		panic("no bucket found with id")
	}
}

func (b Bucket) Format() string { return string(rune(sys36.ToChar(b.id))) }
func (b Bucket) String() string { return b.Format() }

func (b Bucket) Next() Bucket {
	switch b.id {
	case 0:
		return Bucket1
	case 1:
		return Bucket2
	default:
		return Bucket0
	}
}

func (b Bucket) Prev() Bucket {
	switch b.id {
	case 0:
		return Bucket2
	case 1:
		return Bucket0
	default:
		return Bucket1
	}
}

func (b Bucket) Equals(other Bucket) bool {
	return b.id == other.id
}

// --- Rank ---

// Rank is a lexicographically sortable rank string.
type Rank struct {
	value   string
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
	return minRank.Between(Max())
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
	if bucket.Equals(Bucket0) {
		return RankFrom(bucket, initialMinDec)
	}
	return RankFrom(bucket, initialMaxDec)
}

// ParseRank parses a rank string like "0|000000:".
func ParseRank(s string) Rank {
	parts := strings.SplitN(s, "|", 2)
	bucket := BucketFrom(parts[0])
	decimal := ParseDecimal(parts[1], sys36)
	return RankFrom(bucket, decimal)
}

// RankFrom creates a Rank from a bucket and decimal.
func RankFrom(bucket Bucket, decimal Decimal) Rank {
	return Rank{
		value:   bucket.Format() + "|" + formatDecimal(decimal),
		bucket:  bucket,
		decimal: decimal,
	}
}

func (r Rank) String() string      { return r.value }
func (r Rank) Format() string      { return r.value }
func (r Rank) GetBucket() Bucket   { return r.bucket }
func (r Rank) GetDecimal() Decimal { return r.decimal }

func (r Rank) IsMin() bool { return r.decimal.Equals(minDecimal) }
func (r Rank) IsMax() bool { return r.decimal.Equals(maxDecimal) }

func (r Rank) Equals(other Rank) bool { return r.value == other.value }

func (r Rank) CompareTo(other Rank) int {
	if r.value == other.value {
		return 0
	}
	if r.value < other.value {
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

func (r Rank) Between(other Rank) Rank {
	if !r.bucket.Equals(other.bucket) {
		panic("between works only within the same bucket")
	}
	cmp := r.decimal.CompareTo(other.decimal)
	if cmp > 0 {
		return RankFrom(r.bucket, betweenDecimals(other.decimal, r.decimal))
	}
	if cmp == 0 {
		panic("try to rank between issues with same rank")
	}
	return RankFrom(r.bucket, betweenDecimals(r.decimal, other.decimal))
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

func formatDecimal(decimal Decimal) string {
	formatVal := decimal.Format()
	radixChar := decimal.System().RadixPointChar()
	zeroChar := decimal.System().ToChar(0)

	partialIdx := strings.IndexByte(formatVal, radixChar)
	if partialIdx < 0 {
		partialIdx = len(formatVal)
		formatVal = formatVal + string(radixChar)
	}

	for partialIdx < 6 {
		formatVal = string(zeroChar) + formatVal
		partialIdx++
	}

	for len(formatVal) > 0 && formatVal[len(formatVal)-1] == zeroChar {
		formatVal = formatVal[:len(formatVal)-1]
	}

	return formatVal
}
