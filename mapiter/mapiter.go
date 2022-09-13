package mapiter

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Iterate creates an iterator from arbitrary map types. This is not
// the most efficient tool, but it's the quickest way to create an
// iterator for maps.
// Also, note that you cannot make any assumptions on the order of
// pairs being returned.
func Iterate[K comparable, V any](ctx context.Context, m interface{}) (Iterator[K, V], error) {
	if rv := reflect.ValueOf(m); rv.Kind() != reflect.Map {
		return nil, fmt.Errorf(`argument must be a map (%s)`, rv.Type())
	}

	ch := make(chan *Pair[K, V])
	go func(ctx context.Context, ch chan *Pair[K, V], m map[K]V) {
		defer close(ch)
		for k, v := range m {
			pair := &Pair[K, V]{
				Key:   k,
				Value: v,
			}
			select {
			case <-ctx.Done():
				return
			case ch <- pair:
			}
		}
	}(ctx, ch, m.(map[K]V))

	return New(ch), nil
}

// Source represents a map that knows how to create an iterator
type Source[K comparable, V any] interface {
	Iterate(context.Context) Iterator[K, V]
}

// Pair represents a single pair of key and value from a map
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// Iterator iterates through keys and values of a map
type Iterator[K comparable, V any] interface {
	Next(context.Context) bool
	Pair() *Pair[K, V]
}

type iter[K comparable, V any] struct {
	ch   chan *Pair[K, V]
	mu   sync.RWMutex
	next *Pair[K, V]
}

// Visitor represents an object that handles each pair in a map
type Visitor[K comparable, V any] interface {
	Visit(K, V) error
}

// VisitorFunc is a type of Visitor based on a function
type VisitorFunc func(interface{}, interface{}) error

func (fn VisitorFunc) Visit(s interface{}, v interface{}) error {
	return fn(s, v)
}

func New[K comparable, V any](ch chan *Pair[K, V]) Iterator[K, V] {
	return &iter[K, V]{
		ch: ch,
	}
}

// Next returns true if there are more items to read from the iterator
func (i *iter[K, V]) Next(ctx context.Context) bool {
	i.mu.RLock()
	if i.ch == nil {
		i.mu.RUnlock()
		return false
	}
	i.mu.RUnlock()

	i.mu.Lock()
	defer i.mu.Unlock()
	select {
	case <-ctx.Done():
		i.ch = nil
		return false
	case v, ok := <-i.ch:
		if !ok {
			i.ch = nil
			return false
		}
		i.next = v
		return true
	}
}

// Pair returns the currently buffered Pair. Calling Next() will reset its value
func (i *iter[K, V]) Pair() *Pair[K, V] {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.next
}

// Walk walks through each element in the map
func Walk[K comparable, V any](ctx context.Context, s Source[K, V], v Visitor[K, V]) error {
	for i := s.Iterate(ctx); i.Next(ctx); {
		pair := i.Pair()
		if err := v.Visit(pair.Key, pair.Value); err != nil {
			return fmt.Errorf(`failed to visit key %v: %w`, pair.Key, err)
		}
	}
	return nil
}

// AsMap returns the values obtained from the source as a map
func AsMap[K comparable, V any](ctx context.Context, s interface{}, v interface{}) error {
	var iter Iterator[K, V]
	switch reflect.ValueOf(s).Kind() {
	case reflect.Map:
		x, err := Iterate[K, V](ctx, s.(map[K]V))
		if err != nil {
			return fmt.Errorf(`failed to iterate over map type: %w`, err)
		}
		iter = x
	default:
		ssrc, ok := s.(Source[K, V])
		if !ok {
			return fmt.Errorf(`cannot iterate over %T: not a mapiter.Source type`, s)
		}
		iter = ssrc.Iterate(ctx)
	}

	dst := reflect.ValueOf(v)

	// dst MUST be a pointer to a map type
	if kind := dst.Kind(); kind != reflect.Ptr {
		return fmt.Errorf(`dst must be a pointer to a map (%s)`, dst.Type())
	}

	dst = dst.Elem()
	if dst.Kind() != reflect.Map {
		return fmt.Errorf(`dst must be a pointer to a map (%s)`, dst.Type())
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	// dst must be assignable
	if !dst.CanSet() {
		return fmt.Errorf(`dst is not writeable`)
	}

	keytyp := dst.Type().Key()
	valtyp := dst.Type().Elem()

	for iter.Next(ctx) {
		pair := iter.Pair()

		rvkey := reflect.ValueOf(pair.Key)
		rvvalue := reflect.ValueOf(pair.Value)

		if !rvkey.Type().AssignableTo(keytyp) {
			return fmt.Errorf(`cannot assign key of type %s to map key of type %s`, rvkey.Type(), keytyp)
		}

		switch rvvalue.Kind() {
		// we can only check if we can assign to rvvalue to valtyp if it's non-nil
		case reflect.Invalid:
			rvvalue = reflect.New(valtyp).Elem()
		default:
			if !rvvalue.Type().AssignableTo(valtyp) {
				return fmt.Errorf(`cannot assign value of type %s to map value of type %s`, rvvalue.Type(), valtyp)
			}
		}

		dst.SetMapIndex(rvkey, rvvalue)
	}

	return nil
}
