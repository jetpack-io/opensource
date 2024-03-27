package typeid

// lazy is a type that can be used to defer the parsing of a typeid string until
// it is actually needed. This can be useful to avoid repetitive parsing and
// instead allow for deferred parsing and error handling in a central location.
type lazy[T Subtype, PT SubtypePtr[T]] string

// Lazy interface can be used with subtype, e.g. typeid.Lazy[MyID]
type Lazy[T Subtype] interface {
	Parse() (T, error)
}

func (l lazy[T, PT]) Parse() (T, error) {
	return Parse[T, PT](string(l))
}

func LazyParse[T Subtype, PT SubtypePtr[T]](s string) lazy[T, PT] {
	return lazy[T, PT](s)
}
