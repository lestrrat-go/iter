package mapiter_test

import (
	"context"
	"testing"

	"github.com/lestrrat-go/iter/mapiter"
	"github.com/stretchr/testify/assert"
)

func TestAsStrIfaceMap(t *testing.T) {
	t.Run("Map-like object", func(t *testing.T) {
		src := &MapLike{
			Values: map[string]int{
				"one":   1,
				"two":   2,
				"three": 3,
				"four":  4,
				"five":  5,
			},
		}

		t.Run("returns identical map", func(t *testing.T) {
			got, err := mapiter.AsStrIfaceMap(context.Background(), src)
			if !assert.NoError(t, err, `AsStrIfaceMap returns identical MapLike object`) {
				return
			}

			want := map[string]interface{}{
				"one":   1,
				"two":   2,
				"three": 3,
				"four":  4,
				"five":  5,
			}

			if !assert.Equal(t, want, got, "maps should match") {
				return
			}
		})
	})
}
