package mapiter

import (
	"context"
	"fmt"
)

// StrKeyVisitor is a custom visitor.
// Whereas Visitor supports any type of key, this
// visitor assumes the key is a string
type StrKeyVisitor interface {
	Visit(string, interface{}) error
}

type StrKeyVisitorFunc func(string, interface{}) error

func (fn StrKeyVisitorFunc) Visit(s string, v interface{}) error {
	return fn(s, v)
}

func WalkMap(ctx context.Context, src Source, visitor StrKeyVisitor) error {
	return Walk(ctx, src, VisitorFunc(func(k, v interface{}) error {
		//nolint:forcetypeassert
		return visitor.Visit(k.(string), v)
	}))
}

func AsStrIfaceMap(ctx context.Context, src Source) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := AsMap(ctx, src, &m); err != nil {
		return nil, fmt.Errorf(`mapiter.AsMap failed: %w`, err)
	}
	return m, nil
}
