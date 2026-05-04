package lexorank

import "testing"

func TestMin(t *testing.T) {
	rank := Min()
	if got := rank.String(); got != "0|000000:" {
		t.Errorf("Min() = %q, want %q", got, "0|000000:")
	}
}

func TestMax(t *testing.T) {
	rank := Max()
	if got := rank.String(); got != "0|zzzzzz:" {
		t.Errorf("Max() = %q, want %q", got, "0|zzzzzz:")
	}
}

func TestBetweenMinMax(t *testing.T) {
	minRank := Min()
	maxRank := Max()
	between := minRank.Between(maxRank)
	if got := between.String(); got != "0|hzzzzz:" {
		t.Errorf("min.between(max) = %q, want %q", got, "0|hzzzzz:")
	}
	if minRank.CompareTo(between) >= 0 {
		t.Error("min should be less than between")
	}
	if maxRank.CompareTo(between) <= 0 {
		t.Error("max should be greater than between")
	}
}

func TestBetweenMinGenNext(t *testing.T) {
	minRank := Min()
	nextRank := minRank.GenNext()
	between := minRank.Between(nextRank)
	if got := between.String(); got != "0|0i0000:" {
		t.Errorf("min.between(min.genNext()) = %q, want %q", got, "0|0i0000:")
	}
	if minRank.CompareTo(between) >= 0 {
		t.Error("min should be less than between")
	}
	if nextRank.CompareTo(between) <= 0 {
		t.Error("nextRank should be greater than between")
	}
}

func TestBetweenMaxGenPrev(t *testing.T) {
	maxRank := Max()
	prevRank := maxRank.GenPrev()
	between := maxRank.Between(prevRank)
	if got := between.String(); got != "0|yzzzzz:" {
		t.Errorf("max.between(max.genPrev()) = %q, want %q", got, "0|yzzzzz:")
	}
	if maxRank.CompareTo(between) <= 0 {
		t.Error("max should be greater than between")
	}
	if prevRank.CompareTo(between) >= 0 {
		t.Error("prevRank should be less than between")
	}
}

func TestMoveTo(t *testing.T) {
	tests := []struct {
		prevStep int
		nextStep int
		expected string
	}{
		{0, 1, "0|0i0000:"},
		{1, 0, "0|0i0000:"},
		{3, 5, "0|10000o:"},
		{5, 3, "0|10000o:"},
		{15, 30, "0|10004s:"},
		{31, 32, "0|10006s:"},
		{100, 200, "0|1000x4:"},
		{200, 100, "0|1000x4:"},
	}
	for _, tt := range tests {
		prevRank := Min()
		for i := 0; i < tt.prevStep; i++ {
			prevRank = prevRank.GenNext()
		}
		nextRank := Min()
		for i := 0; i < tt.nextStep; i++ {
			nextRank = nextRank.GenNext()
		}
		between := prevRank.Between(nextRank)
		if got := between.String(); got != tt.expected {
			t.Errorf("between(genNext^%d(min), genNext^%d(min)) = %q, want %q", tt.prevStep, tt.nextStep, got, tt.expected)
		}
	}
}
