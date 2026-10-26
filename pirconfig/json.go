package pirconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Check each object separately so identical keys in different entries are legal.
// Reject null rather than interpreting it as an omitted or zero-valued field.
func checkJSON(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := checkValue(dec, 0); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON document")
	}
	return nil
}

func checkValue(dec *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("JSON nesting exceeds 64")
	}
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("null is not allowed; omit optional fields")
	}
	d, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch d {
	case '{':
		seen := make(map[string]bool)
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("invalid object key")
			}
			if seen[name] {
				return fmt.Errorf("duplicate JSON key %q", name)
			}
			seen[name] = true
			if err := checkValue(dec, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := checkValue(dec, depth+1); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
	_, err = dec.Token()
	return err
}

// encoding/json matches field names without regard to case and does not check
// required fields. Match the declared JSON names exactly before typed decoding.
func checkShape(raw []byte, typ reflect.Type) error {
	if typ.Kind() == reflect.Pointer {
		return checkShape(raw, typ.Elem())
	}
	switch typ.Kind() {
	case reflect.Struct:
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			return err
		}
		fields := make(map[string]bool)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")
			name := tag[0]
			fields[name] = true
			value, exists := object[name]
			optional := len(tag) > 1 && tag[1] == "omitempty"
			if !exists {
				if !optional {
					return fmt.Errorf("missing required field %q", name)
				}
				continue
			}
			if err := checkShape(value, field.Type); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		for name := range object {
			if !fields[name] {
				return fmt.Errorf("unknown field %q", name)
			}
		}
	case reflect.Slice:
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		for _, v := range values {
			if err := checkShape(v, typ.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}
