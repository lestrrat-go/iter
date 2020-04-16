package mapiter

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

// Source represents a map that knows how to create an iterator
type Source interface {
	Iterate(context.Context) Iterator
}

// Pair represents a single pair of key and value from a map
type Pair struct {
	Key   interface{}
	Value interface{}
}

// Iterator iterates through keys and values of a map
type Iterator interface {
	Next(context.Context) bool
	Pair() *Pair
}

type iter struct {
	ch   chan *Pair
	mu   sync.RWMutex
	next *Pair
}

// Visitor represents an object that handles each pair in a map
type Visitor interface {
	Visit(interface{}, interface{}) error
}

// VisitorFunc is a type of Visitor based on a function
type VisitorFunc func(interface{}, interface{}) error

func (fn VisitorFunc) Visit(s interface{}, v interface{}) error {
	return fn(s, v)
}

func New(ch chan *Pair) Iterator {
	return &iter{
		ch: ch,
	}
}

// Next returns true if there are more items to read from the iterator
func (i *iter) Next(ctx context.Context) bool {
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

	return false // never reached
}

// Pair returns the currently buffered Pair. Calling Next() will reset its value
func (i *iter) Pair() *Pair {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.next
}

// Walk walks through each element in the map
func Walk(ctx context.Context, s Source, v Visitor) error {
	for i := s.Iterate(ctx); i.Next(ctx); {
		pair := i.Pair()
		if err := v.Visit(pair.Key, pair.Value); err != nil {
			return errors.Wrapf(err, `failed to visit key %s`, pair.Key)
		}
	}
	return nil
}

// AsMap returns the values obtained from the source as a map[interface{}]interface{}
func AsMap(ctx context.Context, src Source, v interface{}) error {
	dst := reflect.ValueOf(v)

	// dst MUST be a pointer to a map type
	if kind := dst.Kind(); kind != reflect.Ptr {
		return errors.Errorf(`dst must be a pointer to a map (%s)`, dst.Type())
	}

	dst = dst.Elem()
	if dst.Kind() != reflect.Map {
		return errors.Errorf(`dst must be a pointer to a map (%s)`, dst.Type())
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	// dst must be assignable
	if !dst.CanSet() {
		return errors.New(`dst is not writeable`)
	}

	keytyp := dst.Type().Key()
	valtyp := dst.Type().Elem()

	for iter := src.Iterate(ctx); iter.Next(ctx); {
		pair := iter.Pair()

		rvkey := reflect.ValueOf(pair.Key)
		rvvalue := reflect.ValueOf(pair.Value)

		if !rvkey.Type().AssignableTo(keytyp) {
			return errors.Errorf(`cannot assign key of type %s to map key of type %s`, rvkey.Type(), keytyp)
		}
		if !rvvalue.Type().AssignableTo(valtyp) {
			return errors.Errorf(`cannot assign value of type %s to map value of type %s`, rvvalue.Type(), valtyp)
		}

		dst.SetMapIndex(rvkey, rvvalue)
	}

	return nil
}
