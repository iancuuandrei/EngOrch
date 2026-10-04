package canonical

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	typedJSONMarshaler = reflect.TypeFor[json.Marshaler]()
	typedTextMarshaler = reflect.TypeFor[encoding.TextMarshaler]()
	typedRawMessage    = reflect.TypeFor[json.RawMessage]()
)

// TypedGeneratedBytes canonicalizes a closed typed value without decoding the
// JSON produced by json.Marshal into maps. Unsupported values fall back to
// Bytes. It is intended only for internally constructed immutable values.
func TypedGeneratedBytes(value any) ([]byte, error) {
	if !typedValueSupported(reflect.ValueOf(value), 0) {
		return Bytes(value)
	}
	if err := validText(reflect.ValueOf(value), 0); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	// Keep the generic raw-byte admission boundary exactly unchanged.
	if len(raw) > MaxBytes || !utf8.Valid(raw) || stringsValid(raw) != nil {
		return Normalize(raw)
	}
	var out bytes.Buffer
	out.Grow(len(raw))
	if !typedWrite(&out, reflect.ValueOf(value), 0) {
		return Normalize(raw)
	}
	return out.Bytes(), nil
}

// TypedGeneratedHash is Hash for TypedGeneratedBytes. Untrusted data must use
// Hash, Normalize, or Decode directly.
func TypedGeneratedHash(domain string, value any) (string, error) {
	if domain == "" || strings.ContainsAny(domain, "\r\n") {
		return "", errors.New("invalid hash domain")
	}
	for _, c := range domain {
		if c < 33 || c > 126 {
			return "", errors.New("invalid hash domain")
		}
	}
	encoded, err := TypedGeneratedBytes(value)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(domain + "\n"))
	h.Write(encoded)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func typedValueSupported(value reflect.Value, depth int) bool {
	if depth > 64 || !value.IsValid() {
		return false
	}
	t := value.Type()
	if t == typedRawMessage || t.Implements(typedJSONMarshaler) || t.Implements(typedTextMarshaler) || reflect.PointerTo(t).Implements(typedJSONMarshaler) || reflect.PointerTo(t).Implements(typedTextMarshaler) {
		return false
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return false
	case reflect.Pointer:
		return value.IsNil() || typedValueSupported(value.Elem(), depth+1)
	case reflect.String:
		return utf8.ValidString(value.String())
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		for i := 0; i < value.Len(); i++ {
			if !typedValueSupported(value.Index(i), depth+1) {
				return false
			}
		}
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return false
		}
		iter := value.MapRange()
		for iter.Next() {
			if !utf8.ValidString(iter.Key().String()) || !typedValueSupported(iter.Value(), depth+1) {
				return false
			}
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" || field.Anonymous {
				return false
			}
			if !typedValueSupported(value.Field(i), depth+1) {
				return false
			}
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return false
	}
	return true
}

func typedWrite(out *bytes.Buffer, value reflect.Value, depth int) bool {
	if depth > 64 || !value.IsValid() {
		return false
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			out.WriteString("null")
			return true
		}
		return typedWrite(out, value.Elem(), depth+1)
	}
	switch value.Kind() {
	case reflect.String:
		quote(out, value.String())
	case reflect.Bool:
		out.WriteString(strconv.FormatBool(value.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := value.Int()
		if n > 9007199254740991 || n < -9007199254740991 {
			return false
		}
		out.WriteString(strconv.FormatInt(n, 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := value.Uint()
		if n > 9007199254740991 {
			return false
		}
		out.WriteString(strconv.FormatUint(n, 10))
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			out.WriteString("null")
			return true
		}
		out.WriteByte('[')
		for i := 0; i < value.Len(); i++ {
			if i > 0 {
				out.WriteByte(',')
			}
			if !typedWrite(out, value.Index(i), depth+1) {
				return false
			}
		}
		out.WriteByte(']')
	case reflect.Map:
		if value.IsNil() {
			out.WriteString("null")
			return true
		}
		keys := value.MapKeys()
		for _, key := range keys {
			if !typedASCIIKey(key.String()) {
				return false
			}
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		out.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			quote(out, key.String())
			out.WriteByte(':')
			if !typedWrite(out, value.MapIndex(key), depth+1) {
				return false
			}
		}
		out.WriteByte('}')
	case reflect.Struct:
		type fieldValue struct {
			name  string
			value reflect.Value
		}
		fields := make([]fieldValue, 0, value.NumField())
		for i := 0; i < value.NumField(); i++ {
			fieldType := value.Type().Field(i)
			if !fieldType.IsExported() {
				continue
			}
			name, omitEmpty := typedJSONField(fieldType)
			if name == "" {
				return false
			}
			field := value.Field(i)
			if omitEmpty && typedEmpty(field) {
				continue
			}
			fields = append(fields, fieldValue{name: name, value: field})
		}
		sort.Slice(fields, func(i, j int) bool { return fields[i].name < fields[j].name })
		out.WriteByte('{')
		for i, field := range fields {
			if i > 0 {
				out.WriteByte(',')
			}
			quote(out, field.name)
			out.WriteByte(':')
			if !typedWrite(out, field.value, depth+1) {
				return false
			}
		}
		out.WriteByte('}')
	default:
		return false
	}
	return true
}

func typedEmpty(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array:
		return false
	case reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.Interface, reflect.Pointer:
		return value.IsNil()
	}
	return false
}

func typedJSONField(field reflect.StructField) (string, bool) {
	if field.Anonymous {
		return "", false
	}
	parts := strings.Split(field.Tag.Get("json"), ",")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "-" {
		return "", false
	}
	for _, option := range parts[1:] {
		if option == "string" {
			return "", false
		}
	}
	return parts[0], slicesContains(parts[1:], "omitempty")
}

func typedASCIIKey(value string) bool {
	for _, c := range value {
		if c > 127 {
			return false
		}
	}
	return true
}

func slicesContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
