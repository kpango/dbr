package dbr

import (
	"reflect"
)

func buildCond(d Dialect, buf Buffer, pred string, cond ...Builder) error {
	for i, c := range cond {
		if i > 0 {
			buf.WriteString(" ")
			buf.WriteString(pred)
			buf.WriteString(" ")
		}
		buf.WriteString("(")
		err := c.Build(d, buf)
		if err != nil {
			return err
		}
		buf.WriteString(")")
	}
	return nil
}

// And creates AND from a list of conditions
func And(cond ...Builder) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCond(d, buf, "AND", cond...)
	})
}

// Or creates OR from a list of conditions
func Or(cond ...Builder) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCond(d, buf, "OR", cond...)
	})
}

func buildCmp(d Dialect, buf Buffer, pred string, column string, value interface{}) error {
	buf.WriteString(d.QuoteIdent(column))
	buf.WriteString(" ")
	buf.WriteString(pred)
	buf.WriteString(" ")
	buf.WriteString(placeholder)

	buf.WriteValue(value)
	return nil
}

// Eq is `=`.
// When value is nil, it will be translated to `IS NULL`.
// When value is a slice, it will be translated to `IN`.
// Otherwise it will be translated to `=`.
func Eq(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		if value == nil {
			buf.WriteString(d.QuoteIdent(column))
			buf.WriteString(" IS NULL")
			return nil
		}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			if v.Len() == 0 {
				buf.WriteString(d.EncodeBool(false))
				return nil
			}
			return buildCmp(d, buf, "IN", column, value)
		}
		return buildCmp(d, buf, "=", column, value)
	})
}

// Neq is `!=`.
// When value is nil, it will be translated to `IS NOT NULL`.
// When value is a slice, it will be translated to `NOT IN`.
// Otherwise it will be translated to `!=`.
func Neq(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		if value == nil {
			buf.WriteString(d.QuoteIdent(column))
			buf.WriteString(" IS NOT NULL")
			return nil
		}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			if v.Len() == 0 {
				buf.WriteString(d.EncodeBool(true))
				return nil
			}
			return buildCmp(d, buf, "NOT IN", column, value)
		}
		return buildCmp(d, buf, "!=", column, value)
	})
}

// Gt is `>`.
func Gt(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, ">", column, value)
	})
}

// Gte is '>='.
func Gte(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, ">=", column, value)
	})
}

// Lt is '<'.
func Lt(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, "<", column, value)
	})
}

// Lte is `<=`.
func Lte(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildCmp(d, buf, "<=", column, value)
	})
}

func buildLikeCmp(d Dialect, buf Buffer, pred string, column string, value interface{}) error {
	if value == nil {
		return ErrColumnNotSpecified
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		// pass as is
		return buildCmp(d, buf, pred, column, value)
	case reflect.Ptr, reflect.Interface: // pointer or interface
		// for pointers & interfaces check
		return buildLikeCmp(d, buf, pred, column, v.Elem().Interface())
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Uint8: // bytes
			// interpolator will handle this case
			return buildCmp(d, buf, pred, column, value)
		case reflect.Int32: // rune
			// need to convert into string
			return buildCmp(d, buf, pred, column, string(value.([]rune)))
		}
		fallthrough
	default:
		return ErrColumnNotSpecified
	}
}

// Like is `LIKE`.
// When value is nil, do nothing.
// When value is a slice, do nothing.
// Otherwise it will be translated to `LIKE`.
func Like(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildLikeCmp(d, buf, "LIKE", column, value)
	})
}

// NotLike is `NOT LIKE`.
// When value is nil, do nothing.
// When value is a slice, do nothing.
// Otherwise it will be translated to `NOT LIKE`.
func NotLike(column string, value interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		return buildLikeCmp(d, buf, "NOT LIKE", column, value)
	})
}
