# iter

Simple tools for container iteration

# DESCRIPTION

`iter` and its sub-packages provide a set of utilities to make it easy for
providers of objects that are iteratable.

For example, if your object is map-like and you want a way for users to
iterate through all or specific keys in your object, all you need to do
is to provide a function that iterates through the pairs that you want,
and send them to a channel.

Then you create an iterator from the channel, and pass the iterator to the
user. The user can then safely iterate through all elements

## Channel-based iterator

Iterators pass its values via channels. Your implementation must provide a "source"
that writes to this channel.

```go
func iterate(ch chan *mapiter.Pair) {
  ch <- &mapiter.Pair{Key: "key1", Value: ...}
  ch <- &mapiter.Pair{Key: "key2", Value: ...}
  ...
  // DO NOT forget to close the channel
  close(ch)
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

## Iterate over native containers (map/array)

```go
m := make(map[...]...) // key and value may be any type

for iter := mapiter.Iterate(ctx, m); iter.Next(ctx); {
	...
}
```

```go
s := make([]...) // element may be any type

for iter := arrayiter.Iterate(ctx, s); iter.Next(ctx); {
  ...
}
```
