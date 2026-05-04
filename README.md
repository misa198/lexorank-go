# LexoRank in Go

A Go implementation of a list ordering system based on [JIRA's LexoRank algorithm](https://www.youtube.com/watch?v=OjQv9xMoFbg).

Generates lexicographically sortable string keys for ordering items. String comparison produces correct ordering — no numeric parsing needed.

## Install

```sh
go get github.com/misa198/lexorank-go
```

## Usage

### Package-level functions

```go
import "github.com/misa198/lexorank-go"

// min rank: "0|000000:"
minRank := lexorank.Min()

// max rank: "0|zzzzzz:"
maxRank := lexorank.Max()

// middle rank: "0|hzzzzz:"
midRank := lexorank.Middle()

// parse from string
parsed := lexorank.ParseRank("0|0i0000:")
```

### Generate next/previous

```go
rank := lexorank.Min()

next := rank.GenNext() // "0|100000:"
prev := rank.GenPrev() // generates rank before current
```

### Insert between two ranks

```go
a := lexorank.Min()
b := a.GenNext().GenNext()

between := a.Between(b) // rank between a and b
fmt.Println(between)    // "0|100000:"
```

### Buckets

Three buckets (0, 1, 2) partition the rank space for rebalancing.

```go
rank := lexorank.Min()

// move to next bucket
nextBucket := rank.InNextBucket()

// move to previous bucket
prevBucket := rank.InPrevBucket()
```

### Comparing ranks

```go
a := lexorank.Min()
b := lexorank.Max()

a.CompareTo(b) // -1 (a < b)
a.Equals(b)    // false

// string comparison also works since ranks are lexicographically sorted
a.String() < b.String() // true
```

## Rank format

Ranks are strings like `"0|hzzzzz:"`:

- `0` — bucket (0, 1, or 2)
- `|` — separator
- `hzzzzz` — base-36 decimal value (digits `0-z`)
- `:` — radix point

Normal string comparison produces correct ordering.

## Related projects

- [LexoRank in TypeScript](https://github.com/kvandake/lexorank)
- [LexoRank in C#](https://github.com/kvandake/lexorank-dotnet)

## License

MIT
