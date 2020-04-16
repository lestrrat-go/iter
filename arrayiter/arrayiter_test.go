package arrayiter_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/lestrrat-go/iter/arrayiter"
	"github.com/stretchr/testify/assert"
)

func TestIterator(t *testing.T) {
	chSize := 2

	ch := make(chan *arrayiter.Pair, chSize)
	ch <- &arrayiter.Pair{Index: 1, Value: 2}
	ch <- &arrayiter.Pair{Index: 2, Value: 4}
	close(ch)

	i := arrayiter.New(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var loopCount int
	for i.Next(ctx) {
		loopCount++
		p := i.Pair()
		if !assert.Equal(t, p.Value, 2*loopCount, "expected values to match") {
			return
		}
	}

	if !assert.Equal(t, chSize, loopCount, "expected to loop for %d times", chSize) {
		return
	}
}

type ArrayLike struct {
	Values []string
}

func (m *ArrayLike) Iterate(ctx context.Context) arrayiter.Iterator {
	ch := make(chan *arrayiter.Pair)
	go m.iterate(ctx, ch)
	return arrayiter.New(ch)
}

func (m *ArrayLike) iterate(ctx context.Context, ch chan *arrayiter.Pair) {
	defer close(ch)
	for k, v := range m.Values {
		ch <- &arrayiter.Pair{Index: k, Value: v}
	}
}

func TestAsArray(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		inputs := []interface{}{
			[]string{
				"foo",
				"bar",
				"baz",
			},
		}
		for _, x := range inputs {
			input := x
			t.Run(fmt.Sprintf("%T", input), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				dst := reflect.New(reflect.TypeOf(input))
				if !assert.NoError(t, arrayiter.AsArray(ctx, input, dst.Interface()), `arrayiter.AsArray should succeed`) {
					return
				}
				if !assert.Equal(t, input, dst.Elem().Interface(), `slices should be the same`) {
					return
				}
			})
		}
	})
	t.Run("Array-like object", func(t *testing.T) {
		src := &ArrayLike{
			Values: []string{
				"one",
				"two",
				"three",
				"four",
				"five",
			},
		}

		t.Run("dst is nil (slice)", func(t *testing.T) {
			var m []string
			if !assert.NoError(t, arrayiter.AsArray(context.Background(), src, &m), `AsArray against uninitialized array should succeed`) {
				return
			}

			if !assert.Equal(t, src.Values, m, "slices should match") {
				return
			}
		})
		t.Run("dst is nil (array)", func(t *testing.T) {
			var m [5]string
			if !assert.NoError(t, arrayiter.AsArray(context.Background(), src, &m), `AsArray against uninitialized array should succeed`) {
				return
			}

			var expected [5]string
			for i, v := range src.Values {
				expected[i] = v
			}

			if !assert.Equal(t, expected, m, "arrays should match") {
				return
			}
		})
		t.Run("dst is not nil", func(t *testing.T) {
			m := make([]string, len(src.Values))
			if !assert.NoError(t, arrayiter.AsArray(context.Background(), src, &m), `AsArray against nil map should succeed`) {
				return
			}

			if !assert.Equal(t, src.Values, m, "maps should match") {
				return
			}
		})
	})
}
