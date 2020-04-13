# iter

Simple tools for container iteration

# DESCRIPTION

## Channel-based iterator

Iterators pass its values via channels. Your implementation must provide a "source"
that writes to this channel.

```go
func iterate(ch chan *mapiter.Pair) mapiter.Iterator {
  ch <- &mapiter.Pair{Key: "key1", Value: ...}
  ch <- &mapiter.Pair{Key: "key2", Value: ...}
  ...
	// DO NOT forget to close the channel
	close(ch)
  iter := mapiter.New(ch)
  return iter
}

ch := make(chan *mapiter.Pair)
go iterate(ch)

for iter := mapiter.New(ch); i.Next(ctx); {
  pair := i.Pair()
  ...
}
```

## Convenience functions

As long as an object implements the appropriate method, you can use the 
convenience functions

```go
fn := func(k string, v interface{}) error {
  ...
}

mapiter.Walk(ctx, source, mapiter.VisitorFunc(fn))
```

There are also functions to convert map-like objects to a map, and array-like objects
to an array/slice

```go
var l []string
err := arrayiter.AsArray(ctx, obj, &l)

var m map[string]int
err := mapiter.AsMap(ctx, obj, &m)
```
