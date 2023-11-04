package metaerr

import "fmt"

func StringMeta(name string) func(string) ErrorMetadata {
	return func(val string) ErrorMetadata {
		return func(err Error) []MetaValue {
			return []MetaValue{
				{
					Name:   name,
					Values: []string{val},
				},
			}
		}
	}
}

func StringMetaFromContext(name string, ctxKey string) func() ErrorMetadata {
	return func() ErrorMetadata {
		return func(err Error) []MetaValue {
			if err.Context == nil {
				return nil
			}
			val := err.Context.Value(ctxKey)
			if val == nil {
				return nil
			}
			strVal := fmt.Sprintf("%v", val)
			return []MetaValue{
				{
					Name:   name,
					Values: []string{strVal},
				},
			}
		}
	}
}

func StringsMeta(name string) func(...string) ErrorMetadata {
	return func(values ...string) ErrorMetadata {
		return func(err Error) []MetaValue {
			return []MetaValue{
				{
					Name:   name,
					Values: values,
				},
			}
		}
	}
}

func StringerMeta[T fmt.Stringer](name string) func(T) ErrorMetadata {
	return func(val T) ErrorMetadata {
		return func(err Error) []MetaValue {
			strVal := val.String()
			if strVal == "" {
				return nil
			}
			return []MetaValue{
				{
					Name:   name,
					Values: []string{strVal},
				},
			}
		}
	}
}

type MetaValue struct {
	Name   string
	Values []string
}

type ErrorMetadata = func(err Error) []MetaValue
