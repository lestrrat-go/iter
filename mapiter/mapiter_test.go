package mapiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/iter/mapiter"
	"github.com/stretchr/testify/assert"
)

func TestIterator(t *testing.T) {
	chSize := 2

	ch := make(chan *mapiter.Pair, chSize)
	ch <- &mapiter.Pair{Key: "one", Value: 1}
	ch <- &mapiter.Pair{Key: "two", Value: 2}

	i := mapiter.New(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var loopCount int
	for i.Next(ctx) {
		loopCount++
		p := i.Pair()
		if !assert.Equal(t, p.Value, loopCount, "expected values to match") {
			return
		}
	}

	if !assert.Equal(t, chSize, loopCount, "expected to loop for %d times", chSize) {
		return
	}
}

type MapLike struct {
	Values map[string]int
}

func (m *MapLike) Iterate(ctx context.Context) mapiter.Iterator {
	ch := make(chan *mapiter.Pair)
	go m.iterate(ctx, ch)
	return mapiter.New(ch)
}

func (m *MapLike) iterate(ctx context.Context, ch chan *mapiter.Pair) {
	defer close(ch)
	for k, v := range m.Values {
		ch <-&mapiter.Pair{Key: k, Value: v}
	}
}

func TestAsMap(t *testing.T) {
	src := &MapLike{
		Values: map[string]int{
			"one": 1,
			"two": 2,
			"three": 3,
			"four": 4,
			"five": 5,
		},
	}

	t.Run("dst is nil", func(t *testing.T) {
		var m map[string]int
		if !assert.NoError(t, mapiter.AsMap(context.Background(), src, &m), `AsMap against nil map should succeed`) {
			return
		}

		if !assert.Equal(t, src.Values, m, "maps should match") {
			return
		}
	})
	t.Run("dst is nil (elem type does not match)", func(t *testing.T) {
		var m map[string]string
		if assert.Error(t, mapiter.AsMap(context.Background(), src, &m), `AsMap against nil map should fail`) {
			return
		}
	})
	t.Run("dst is not nil", func(t *testing.T) {
		m := make(map[string]int)
		if !assert.NoError(t, mapiter.AsMap(context.Background(), src, &m), `AsMap against nil map should succeed`) {
			return
		}

		if !assert.Equal(t, src.Values, m, "maps should match") {
			return
		}
	})
}
