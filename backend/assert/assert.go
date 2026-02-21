package assert

import "fmt"

// True asserts that condition must hold. description should explain the
// expected state. Optional detail arguments are formatted with fmt.Sprint and
// appended to the panic message.
func True(condition bool, description string, detail ...any) {
	if !condition {
		panic(message("expected true", description, detail...))
	}
}

// False asserts that condition must be false. description documents the
// forbidden state.
func False(condition bool, description string, detail ...any) {
	if condition {
		panic(message("expected false", description, detail...))
	}
}

// Nil asserts that value is nil.
func Nil(value any, name string, detail ...any) {
	if !isNil(value) {
		panic(message("expected nil", name, detail...))
	}
}

// NotNil asserts that value is non-nil.
func NotNil(value any, name string, detail ...any) {
	if isNil(value) {
		panic(message("expected non-nil", name, detail...))
	}
}

// EmptyString asserts that s is empty.
func EmptyString(s string, name string, detail ...any) {
	if s != "" {
		panic(message("expected empty string", name, detail...))
	}
}

// NotEmptyString asserts that s is non-empty.
func NotEmptyString(s string, name string, detail ...any) {
	if s == "" {
		panic(message("expected non-empty string", name, detail...))
	}
}

// Equal asserts equality for comparable types.
func Equal[T comparable](got, want T, name string, detail ...any) {
	if got != want {
		extra := append(detail, fmt.Sprintf("want %v, got %v", want, got))
		panic(message("expected equal", name, extra...))
	}
}

// NotEqual asserts inequality for comparable types.
func NotEqual[T comparable](left, right T, name string, detail ...any) {
	if left == right {
		panic(message("expected different values", name, detail...))
	}
}

// Implies enforces that when antecedent is true, consequent must also be true.
func Implies(antecedent, consequent bool, rule string, detail ...any) {
	if antecedent && !consequent {
		panic(message("expected implication", rule, detail...))
	}
}

// InRange asserts that value is between min and max inclusive.
func InRange[T Number](value, min, max T, name string, detail ...any) {
	if min > max {
		panic(message("invalid range", name, append(detail, fmt.Sprintf("min %v > max %v", min, max))...))
	}
	if value < min {
		panic(message("value below minimum", fmt.Sprintf("%s >= %v", name, min), detail...))
	}
	if value > max {
		panic(message("value above maximum", fmt.Sprintf("%s <= %v", name, max), detail...))
	}
}

// Unreachable marks code that must never execute.
func Unreachable(detail ...any) {
	panic(message("unreachable", "", detail...))
}

// Fail triggers an assertion failure explicitly.
func Fail(description string, detail ...any) {
	panic(message("assertion failed", description, detail...))
}

// Number captures numeric types accepted by InRange.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func message(prefix, subject string, detail ...any) string {
	msg := prefix
	if subject != "" {
		msg = fmt.Sprintf("%s: %s", msg, subject)
	}
	if len(detail) > 0 {
		msg = fmt.Sprintf("%s – %s", msg, fmt.Sprint(detail...))
	}
	return msg
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	return fmt.Sprint(value) == "<nil>"
}
