package conf

type (
	KeyValue struct {
		Key   string
		Value interface{}
	}

	Annotation map[string]interface{}
)

// With adds one or more annotation key-value pairs.
func (an *Annotation) With(kvs ...KeyValue) Annotation {
	if *an == nil {
		*an = Annotation{}
	}
	for _, kv := range kvs {
		(*an)[kv.Key] = kv.Value
	}
	return *an
}

// Get returns the value of an annoation by key or nil.
func (an Annotation) Get(key string) interface{} {
	val, ok := an[key]
	if !ok {
		return nil
	}
	return val
}

// Has returns true if an annotation identified by key exists.
func (an Annotation) Has(key string) bool {
	_, ok := an[key]
	return ok
}

// SecretValue returns an annotation KeyValue that marks
// an option as secret.
func SecretValue() KeyValue {
	return KeyValue{
		Key:   "system-conf/secret",
		Value: true,
	}
}

// IsSecret returns true if spec is annotated as a secret.
func IsSecret(spec OptionSpec) bool {
	return spec.Annotations.Has("system-conf/secret")
}
