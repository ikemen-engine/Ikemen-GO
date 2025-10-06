package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

// -------------------------------------------------------------------
// Helper Functions
// -------------------------------------------------------------------

// queryPart represents a single part of a query path
type queryPart struct {
	name  string
	index *string
}

// parseQueryPath parses a query string into a slice of queryPart
func parseQueryPath(query string) []queryPart {
	var parts []queryPart
	for _, part := range strings.Split(query, ".") {
		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			name := part[:strings.Index(part, "[")]
			index := part[strings.Index(part, "[")+1 : len(part)-1]
			parts = append(parts, queryPart{name: name, index: &index})
		} else {
			parts = append(parts, queryPart{name: part, index: nil})
		}
	}
	return parts
}

// -------------------------------------------------------------------
// Field Retrieval Functions
// -------------------------------------------------------------------

// findFieldByINITag recursively searches for a struct field with a matching INI tag
func findFieldByINITag(v reflect.Value, tag string) (reflect.Value, reflect.StructField, bool) {
	val := v
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fVal := val.Field(i)
		iniTag := field.Tag.Get("ini")

		if iniTag != "" && strings.EqualFold(iniTag, tag) && field.PkgPath == "" {
			return fVal, field, true
		}
		// If the field is an embedded struct, search recursively
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			rv, rf, found := findFieldByINITag(fVal, tag)
			if found {
				return rv, rf, true
			}
		}
	}
	return reflect.Value{}, reflect.StructField{}, false
}

// findMapFieldWithTag searches for a map field in a struct with a specific INI tag
func findMapFieldWithTag(v reflect.Value, tag string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.EqualFold(field.Tag.Get("ini"), tag) && field.Type.Kind() == reflect.Map {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// findDefaultField returns the first struct field without an INI tag
func findDefaultField(v reflect.Value) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("ini") == "" {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// -------------------------------------------------------------------
// Default Tag Applying and Maps initialization Functions
// -------------------------------------------------------------------
// initMaps traverses all fields/slices/maps of the given struct Value
// and initializes maps if they are nil.
func initMaps(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			initMaps(v.Elem())
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			fieldVal := v.Field(i)
			fieldType := t.Field(i)

			// Only recurse on exported fields
			if fieldType.PkgPath != "" {
				continue
			}
			initMaps(fieldVal)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			initMaps(v.Index(i))
		}
	case reflect.Map:
		// If map is nil, we "make" it
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.MakeMap(v.Type()))
		}

		// Then we also want to recurse into each key/value pair
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			initMaps(val)
		}
	}
}

// applyDefaultsToValue applies default-tagged values to any zero fields.
// This works recursively on structs, pointers, slices, and arrays.
func applyDefaultsToValue(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		applyDefaultsToValue(v.Elem())

	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			fieldVal := v.Field(i)
			fieldType := t.Field(i)

			if fieldType.PkgPath != "" {
				continue
			}

			applyDefaultsToValue(fieldVal)

			defTag := fieldType.Tag.Get("default")
			if defTag != "" && isZeroValue(fieldVal) {
				setDefaultFieldValue(fieldVal, fieldType, defTag)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			applyDefaultsToValue(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			applyDefaultsToValue(val)
		}
	default:
		// Base kind: string, int, float, bool, etc. -> no recursion
	}
}

// setDefaultFieldValue tries to parse `defTag` and set it into fieldVal
func setDefaultFieldValue(fieldVal reflect.Value, fieldType reflect.StructField, defTag string) {
	if !fieldVal.CanSet() {
		return
	}
	kind := fieldVal.Kind()

	switch kind {
	case reflect.String:
		fieldVal.SetString(defTag)
	case reflect.Bool:
		if parsed, err := strconv.ParseBool(defTag); err == nil {
			fieldVal.SetBool(parsed)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if parsed, err := strconv.ParseInt(defTag, 10, 64); err == nil {
			fieldVal.SetInt(parsed)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if parsed, err := strconv.ParseUint(defTag, 10, 64); err == nil {
			fieldVal.SetUint(parsed)
		}
	case reflect.Float32, reflect.Float64:
		if parsed, err := strconv.ParseFloat(defTag, 64); err == nil {
			fieldVal.SetFloat(parsed)
		}
	case reflect.Array:
		parts := strings.Split(defTag, ",")
		for i := 0; i < fieldVal.Len() && i < len(parts); i++ {
			elem := fieldVal.Index(i)
			setDefaultFieldValue(elem, fieldType, strings.TrimSpace(parts[i]))
		}
	case reflect.Slice:
		parts := strings.Split(defTag, ",")
		newSlice := reflect.MakeSlice(fieldVal.Type(), 0, len(parts))
		for _, part := range parts {
			elem := reflect.New(fieldVal.Type().Elem()).Elem()
			setDefaultFieldValue(elem, fieldType, strings.TrimSpace(part))
			newSlice = reflect.Append(newSlice, elem)
		}
		fieldVal.Set(newSlice)
	case reflect.Struct:
		parts := strings.Split(defTag, ",")
		for i := 0; i < fieldVal.NumField() && i < len(parts); i++ {
			fv := fieldVal.Field(i)
			setDefaultFieldValue(fv, fieldType, strings.TrimSpace(parts[i]))
		}
	}
}

// isZeroValue returns true if the reflect.Value is a zero value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		z := reflect.Zero(v.Type()).Interface()
		return reflect.DeepEqual(v.Interface(), z)
	}
}

// -------------------------------------------------------------------
// Value Assignment Functions
// -------------------------------------------------------------------

// setFieldValue assigns a value to a field based on its type
func setFieldValue(fieldVal reflect.Value, value interface{}, defTag string, keyPath string) error {

	// Helper to zero-out the field
	setNil := func(fieldVal reflect.Value) {
		switch fieldVal.Kind() {
		case reflect.Map:
			// do not destroy existing map
			return
		case reflect.Slice:
			fieldVal.Set(reflect.MakeSlice(fieldVal.Type(), 0, 0))
		case reflect.Array:
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		case reflect.String:
			fieldVal.SetString("")
		case reflect.Bool:
			fieldVal.SetBool(false)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldVal.SetInt(0)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldVal.SetUint(0)
		case reflect.Float32, reflect.Float64:
			fieldVal.SetFloat(0.0)
		default:
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		}
	}

	// If "value" is literally nil, treat like empty string, but preserve maps.
	if value == nil {
		switch fieldVal.Kind() {
		case reflect.Map:
			return nil // no change
		default:
			setNil(fieldVal)
			return nil
		}
	}

	// If value is []string, handle that first
	if arr, ok := value.([]string); ok {
		switch fieldVal.Kind() {
		case reflect.Slice:
			newSlice := reflect.MakeSlice(fieldVal.Type(), 0, len(arr))
			for _, v := range arr {
				elem := reflect.New(fieldVal.Type().Elem()).Elem()
				if err := setFieldValue(elem, v, defTag, keyPath); err != nil {
					return err
				}
				newSlice = reflect.Append(newSlice, elem)
			}
			fieldVal.Set(newSlice)
			return nil
		case reflect.Array:
			for i := 0; i < fieldVal.Len() && i < len(arr); i++ {
				elem := fieldVal.Index(i)
				if err := setFieldValue(elem, arr[i], defTag, keyPath); err != nil {
					return err
				}
			}
			return nil
		default:
			// if there's at least one element, treat arr[0] as the final string
			if len(arr) > 0 {
				return setFieldValue(fieldVal, arr[0], defTag, keyPath)
			}
			// else empty => zero out
			setNil(fieldVal)
			return nil
		}
	}

	// Otherwise treat it as a string
	strVal, ok := value.(string)
	if !ok {
		strVal = fmt.Sprintf("%v", value)
	}
	trimmedValue := strings.TrimSpace(strVal)

	// Unwrap pointer
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			applyDefaultsToValue(fieldVal)
		}
		fieldVal = fieldVal.Elem()
	}

	// If the user explicitly gave an empty line => use the defTag if any
	if trimmedValue == "" {
		if defTag == "" {
			// If no default tag => zero out
			setNil(fieldVal)
			return nil
		}
		// If there's a default tag => parse it as if user typed it
		trimmedValue = defTag
	}

	// Now parse the final (non-empty) string
	switch fieldVal.Kind() {
	case reflect.Bool:
		parsed, err := strconv.ParseBool(trimmedValue)
		if err != nil {
			return fmt.Errorf("failed to parse bool for '%s': %v", keyPath, err)
		}
		fieldVal.SetBool(parsed)

	case reflect.Slice:
		var parts []string
		if strings.Contains(trimmedValue, ",") {
			parts = strings.Split(trimmedValue, ",")
		} else {
			parts = []string{trimmedValue}
		}
		newSlice := reflect.MakeSlice(fieldVal.Type(), 0, len(parts))

		for _, p := range parts {
			elem := reflect.New(fieldVal.Type().Elem()).Elem()
			if err := setFieldValue(elem, strings.TrimSpace(p), "", keyPath); err != nil {
				return err
			}
			newSlice = reflect.Append(newSlice, elem)
		}
		fieldVal.Set(newSlice)

	case reflect.Array:
		separator := ","
		if strings.Contains(trimmedValue, "&") {
			separator = "&"
		}
		parts := strings.Split(trimmedValue, separator)

		// Also parse defTag in case we do partial merges
		defParts := []string{}
		if defTag != "" && defTag != trimmedValue {
			// Only parse defParts if defTag != trimmedValue, to avoid loops
			defParts = strings.Split(defTag, ",")
		}

		// If user supplied partial => fill leftover elements from defTag (if present)
		for i := 0; i < fieldVal.Len(); i++ {
			var valStr string
			if i < len(parts) && strings.TrimSpace(parts[i]) != "" {
				// user explicitly set element i
				valStr = strings.TrimSpace(parts[i])
			} else if i < len(defParts) && strings.TrimSpace(defParts[i]) != "" {
				// fallback to defTag for element i
				valStr = strings.TrimSpace(defParts[i])
			} else {
				// if defTag doesn't have it => zero
				valStr = "0"
			}

			elem := fieldVal.Index(i)
			if err := setFieldValue(elem, valStr, "", keyPath); err != nil {
				return err
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(trimmedValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int: %v", err)
		}
		fieldVal.SetInt(parsed)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(trimmedValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse uint: %v", err)
		}
		fieldVal.SetUint(parsed)

	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(trimmedValue, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float: %v", err)
		}
		fieldVal.SetFloat(parsed)

	case reflect.String:
		parsed := strings.Trim(trimmedValue, "\"") // "
		fieldVal.SetString(parsed)

	case reflect.Struct:
		// If the struct has a default field, set its value
		defaultField, found := findDefaultField(fieldVal)
		if !found {
			applyDefaultsToValue(fieldVal)
			return nil
		}
		if err := setFieldValue(defaultField, trimmedValue, "", keyPath); err != nil {
			return err
		}
		applyDefaultsToValue(fieldVal)

	default:
		return fmt.Errorf("unsupported field kind: %s", fieldVal.Kind())
	}

	return nil
}

func assignToPatternMap(v reflect.Value, lastPartName string, value interface{}, final bool) (bool, reflect.Value, error) {
	if v.Kind() != reflect.Struct {
		return false, v, nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		iniTag := field.Tag.Get("ini")
		if iniTag == "" || field.PkgPath != "" {
			continue
		}

		if strings.HasPrefix(strings.ToLower(iniTag), "map:") {
			pattern := iniTag[4:]
			if pattern == "" {
				continue
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			if re.MatchString(lastPartName) {
				fieldVal := v.Field(i)
				if fieldVal.Kind() == reflect.Map && fieldVal.Type().Key().Kind() == reflect.String {
					if fieldVal.IsNil() {
						fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
					}
					mapKey := reflect.ValueOf(strings.ToLower(lastPartName))
					elemType := fieldVal.Type().Elem()
					mapElem := fieldVal.MapIndex(mapKey)

					// If we're applying a final value, we create/set a fully updated entry
					if final {
						var newVal reflect.Value
						if elemType.Kind() == reflect.Ptr {
							// If it's pointer-based, copy existing data if present
							if mapElem.IsValid() {
								newVal = reflect.New(elemType.Elem())
								newVal.Elem().Set(mapElem.Elem())
							} else {
								newVal = reflect.New(elemType.Elem())
							}
						} else {
							// Non-pointer map element
							if mapElem.IsValid() {
								newVal = reflect.New(elemType).Elem()
								newVal.Set(mapElem)
							} else {
								newVal = reflect.New(elemType).Elem()
							}
						}
						applyDefaultsToValue(newVal)

						// setFieldValue takes (fieldVal, value, keyPath, lastPartName)
						if err := setFieldValue(newVal, value, "", lastPartName); err != nil {
							return false, v, err
						}
						// Store the updated element back into the map
						fieldVal.SetMapIndex(mapKey, newVal)
						mapElem = fieldVal.MapIndex(mapKey)

					} else {
						// Not final: just create a fresh entry if the key doesn't exist
						if !mapElem.IsValid() {
							var newVal reflect.Value
							if elemType.Kind() == reflect.Ptr {
								newVal = reflect.New(elemType.Elem())
							} else {
								newVal = reflect.New(elemType).Elem()
							}
							applyDefaultsToValue(newVal)
							fieldVal.SetMapIndex(mapKey, newVal)
							mapElem = fieldVal.MapIndex(mapKey)
						} else {
							// We still apply defaults if the mapElem is valid
							applyDefaultsToValue(mapElem)
						}
					}
					return true, mapElem, nil
				}
			}
		}
	}
	return false, v, nil
}

// assignField assigns a value to a struct field based on query parts
func assignField(structPtr interface{}, parts []queryPart, value interface{}) error {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		fmt.Println("Error: structPtr must be a non-nil pointer")
		return fmt.Errorf("structPtr must be a non-nil pointer")
	}

	extractNames := func(parts []queryPart) []string {
		names := make([]string, len(parts))
		for i, part := range parts {
			names[i] = part.name
			if part.index != nil {
				names[i] += "[" + *part.index + "]"
			}
		}
		return names
	}

	assignDirect := func(v reflect.Value, part queryPart, value interface{}) (bool, error) {
		if v.Kind() == reflect.Struct {
			fieldVal, fieldType, found := findFieldByINITag(v, part.name)
			if found {
				defTag := fieldType.Tag.Get("default") // read the struct's default tag
				err := setFieldValue(fieldVal, value, defTag, strings.Join(extractNames(parts), "."))
				return true, err
			}
		}
		return false, nil
	}

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
				applyDefaultsToValue(v)
			}
			v = v.Elem()
		}

		if i == len(parts)-1 {
			// final part
			if v.Kind() == reflect.Struct {
				assigned, err := assignDirect(v, part, value)
				if err != nil {
					return err
				}
				if assigned {
					return nil
				}

				assigned, _, err = assignToPatternMap(v, part.name, value, true)
				if err != nil {
					return err
				}
				if assigned {
					return nil
				}

				return fmt.Errorf("field '%s' not found in struct or pattern map field", part.name)
			} else if v.Kind() == reflect.Map {
				if value == nil {
					mapKey := reflect.ValueOf(part.name)
					v.SetMapIndex(mapKey, reflect.Value{})
					return nil
				}
				mapKey := reflect.ValueOf(part.name)
				elemType := v.Type().Elem()

				var newVal reflect.Value
				if elemType.Kind() == reflect.Ptr {
					newVal = reflect.New(elemType.Elem())
				} else {
					newVal = reflect.New(elemType).Elem()
				}
				applyDefaultsToValue(newVal)

				if err := setFieldValue(newVal, value, "", strings.Join(extractNames(parts), ".")); err != nil {
					return err
				}
				v.SetMapIndex(mapKey, newVal)
				return nil
			} else {
				return fmt.Errorf("cannot set value on non-struct and non-map kind '%s'", v.Kind())
			}
		}

		if v.Kind() == reflect.Struct {
			fieldVal, _, found := findFieldByINITag(v, part.name)
			if found {
				v = fieldVal
				continue
			}

			mapFieldVal, found := findMapFieldWithTag(v, part.name)
			if found && mapFieldVal.Kind() == reflect.Map {
				if i+1 >= len(parts) {
					return fmt.Errorf("expected map key after '%s'", part.name)
				}
				mapKey := parts[i+1].name
				if mapFieldVal.IsNil() {
					mapFieldVal.Set(reflect.MakeMap(mapFieldVal.Type()))
				}
				mapKeyVal := reflect.ValueOf(mapKey)
				elemType := mapFieldVal.Type().Elem()
				mapElem := mapFieldVal.MapIndex(mapKeyVal)
				var elem reflect.Value
				if mapElem.IsValid() {
					elem = mapElem
				} else {
					if elemType.Kind() == reflect.Ptr {
						elem = reflect.New(elemType.Elem())
					} else {
						elem = reflect.New(elemType).Elem()
					}
					applyDefaultsToValue(elem)
					mapFieldVal.SetMapIndex(mapKeyVal, elem)
				}
				v = elem
				i++
				continue
			}

			assigned, newV, err := assignToPatternMap(v, part.name, value, false)
			if err != nil {
				return err
			}
			if assigned {
				v = newV
				continue
			}

			return fmt.Errorf("field '%s' not found as struct or map field", part.name)

		} else if v.Kind() == reflect.Map {
			if i+1 >= len(parts) {
				return fmt.Errorf("expected map key after '%s'", part.name)
			}
			mapKey := part.name
			mapKeyVal := reflect.ValueOf(mapKey)
			if v.IsNil() {
				v.Set(reflect.MakeMap(v.Type()))
			}
			elemType := v.Type().Elem()
			mapElem := v.MapIndex(mapKeyVal)
			var elem reflect.Value
			if mapElem.IsValid() {
				elem = mapElem
			} else {
				if elemType.Kind() == reflect.Ptr {
					elem = reflect.New(elemType.Elem())
				} else {
					elem = reflect.New(elemType).Elem()
				}
				applyDefaultsToValue(elem)
				v.SetMapIndex(mapKeyVal, elem)
			}
			v = elem
			continue
		} else {
			return fmt.Errorf("cannot traverse into non-struct and non-map kind '%s'", v.Kind())
		}
	}

	return nil
}

// updateINIFile updates the INI file based on the struct, query, and value
func updateINIFile(obj interface{}, iniFile *ini.File, query string, value string) error {
	parts := parseQueryPath(query)
	if len(parts) == 0 {
		return fmt.Errorf("invalid query: '%s'", query)
	}

	getSectionAndKeyForPatternMap := func(v reflect.Value, partName string) bool {
		if v.Kind() != reflect.Struct {
			return false
		}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			iniTag := field.Tag.Get("ini")
			if iniTag == "" || field.PkgPath != "" {
				continue
			}
			if strings.HasPrefix(strings.ToLower(iniTag), "map:") {
				pattern := iniTag[4:]
				re, err := regexp.Compile(pattern)
				if err != nil {
					continue
				}
				if re.MatchString(partName) {
					return true
				}
			}
		}
		return false
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	currentValue := v
	currentType := v.Type()

	var sectionName string    // which [section] in the .ini
	var keyNameParts []string // the final "Key.SubKey..." part
	sectionSet := false
	i := 0

	for i < len(parts) {
		part := parts[i]

		// unwrap pointers
		for currentValue.Kind() == reflect.Ptr {
			if currentValue.IsNil() {
				currentValue.Set(reflect.New(currentValue.Type().Elem()))
				applyDefaultsToValue(currentValue)
			}
			currentValue = currentValue.Elem()
			currentType = currentValue.Type()
		}

		switch currentValue.Kind() {
		case reflect.Struct:
			// First, see if we can find a field with ini:"part.name"
			fieldVal, _, found := findFieldByINITag(currentValue, part.name)
			if found {
				// We found a direct field. We'll record its iniTag as the "section" or "subKey".
				var structField reflect.StructField
				fieldFound := false
				for idx := 0; idx < currentType.NumField(); idx++ {
					f := currentType.Field(idx)
					iniTag := f.Tag.Get("ini")
					if strings.EqualFold(iniTag, part.name) {
						structField = f
						fieldFound = true
						break
					}
					// fallback: check if the field name itself matches
					if !fieldFound && strings.EqualFold(f.Name, part.name) {
						structField = f
						fieldFound = true
					}
				}

				if !fieldFound {
					return fmt.Errorf("field '%s' not found in struct for query '%s'", part.name, query)
				}

				tag := structField.Tag.Get("ini")
				if tag == "" {
					tag = structField.Name
				}

				if !sectionSet {
					sectionName = tag
					sectionSet = true
				} else {
					keyNameParts = append(keyNameParts, tag)
				}

				currentValue = fieldVal
				currentType = currentValue.Type()

				if part.index != nil {
					if currentValue.Kind() == reflect.Array || currentValue.Kind() == reflect.Slice {
						idx, err := strconv.Atoi(*part.index)
						if err != nil || idx < 0 || idx >= currentValue.Len() {
							return fmt.Errorf("invalid index '%s' for field '%s' in query '%s'", *part.index, part.name, query)
						}
						currentValue = currentValue.Index(idx)
						currentType = currentValue.Type()
					} else {
						return fmt.Errorf("field '%s' is not an array or slice in query '%s'", part.name, query)
					}
				}
				i++

			} else {
				// Not a direct field -> check if it's a direct map field with the same tag
				mapFieldVal, mapFound := findMapFieldWithTag(currentValue, part.name)
				if mapFound && mapFieldVal.Kind() == reflect.Map {
					if i+1 >= len(parts) {
						return fmt.Errorf("expected map key after '%s'", part.name)
					}
					if !sectionSet {
						sectionName = part.name
						sectionSet = true
					} else {
						keyNameParts = append(keyNameParts, part.name)
					}
					i++
					mapKey := parts[i].name
					keyNameParts = append(keyNameParts, mapKey)
					i++
					goto finalize

				} else {
					// Now do the pattern-based approach
					foundPattern := getSectionAndKeyForPatternMap(currentValue, part.name)
					if !foundPattern {
						return fmt.Errorf("field '%s' not found as struct or map field", part.name)
					}
					if !sectionSet {
						sectionName = part.name
						sectionSet = true
					} else {
						keyNameParts = append(keyNameParts, part.name)
					}

					i++
					// If there's another part after e.g. Up, that is the actual final key
					if i < len(parts) {
						// e.g. next part is "Up"
						keyNameParts = append(keyNameParts, parts[i].name)
						i++
					}
					goto finalize
				}
			}

		case reflect.Map:
			// If we've stepped into a map, the next part is the map key
			if part.name == "" {
				return fmt.Errorf("expected key for map in query '%s'", query)
			}
			if !sectionSet {
				sectionName = part.name
				sectionSet = true
			} else {
				keyNameParts = append(keyNameParts, part.name)
			}
			i++

		default:
			return fmt.Errorf("unexpected kind '%s' at part '%s'", currentValue.Kind(), part.name)
		}
	}

finalize:
	if len(keyNameParts) == 0 {
		return fmt.Errorf("unable to determine key name from query '%s'", query)
	}

	keyName := strings.Join(keyNameParts, ".")

	section, err := iniFile.GetSection(sectionName)
	if err != nil {
		section, err = iniFile.NewSection(sectionName)
		if err != nil {
			return fmt.Errorf("failed to create section '%s': %v", sectionName, err)
		}
	}

	// If the value is set to nil, remove the key
	if value == "nil" {
		section.DeleteKey(keyName)
		return nil
	}

	key, err := section.GetKey(keyName)
	if err != nil {
		key, err = section.NewKey(keyName, value)
		if err != nil {
			return fmt.Errorf("failed to create key '%s' in section '%s': %v", keyName, sectionName, err)
		}
	} else {
		key.SetValue(value)
	}

	return nil
}

// -------------------------------------------------------------------
// Core Operations
// -------------------------------------------------------------------

// getValueFromPatternMap checks if the struct has a map field with ini:"map:REGEX"
func getValueFromPatternMap(v reflect.Value, partName string) (bool, reflect.Value) {
	if v.Kind() != reflect.Struct {
		return false, reflect.Value{}
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		iniTag := field.Tag.Get("ini")
		if iniTag == "" || field.PkgPath != "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(iniTag), "map:") {
			pattern := iniTag[4:]
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			// Check if partName matches the map regex
			if re.MatchString(partName) {
				fieldVal := v.Field(i)
				if fieldVal.Kind() == reflect.Map && fieldVal.Type().Key().Kind() == reflect.String {
					if fieldVal.IsNil() {
						fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
					}
					// Convert to lower or keep original
					mapKey := reflect.ValueOf(strings.ToLower(partName))
					mapItem := fieldVal.MapIndex(mapKey)
					if !mapItem.IsValid() {
						// Create a new entry
						elemType := fieldVal.Type().Elem()
						var newVal reflect.Value
						if elemType.Kind() == reflect.Ptr {
							newVal = reflect.New(elemType.Elem())
						} else {
							newVal = reflect.New(elemType).Elem()
						}
						applyDefaultsToValue(newVal)
						fieldVal.SetMapIndex(mapKey, newVal)
						mapItem = fieldVal.MapIndex(mapKey)
					}
					return true, mapItem
				}
			}
		}
	}
	return false, reflect.Value{}
}

// GetValue retrieves a value from the struct based on the query and returns it as an interface{}
func GetValue(structPtr interface{}, query string) (interface{}, error) {
	parts := parseQueryPath(query)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid query: '%s'", query)
	}

	var current reflect.Value
	current = reflect.ValueOf(structPtr)

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil, fmt.Errorf("value not set for query: '%s'", query)
			}
			current = current.Elem()
		}

		switch current.Kind() {
		case reflect.Struct:
			// 1) Try normal ini tag
			fieldVal, _, found := findFieldByINITag(current, part.name)
			if found {
				current = fieldVal
			} else {
				// 2) Try pattern-based map
				matched, mapVal := getValueFromPatternMap(current, part.name)
				if matched {
					current = mapVal
				} else {
					return nil, fmt.Errorf("field '%s' not found in struct or pattern map for query '%s'", part.name, query)
				}
			}

			// if we have an index, check if it's slice/array
			if part.index != nil {
				if current.Kind() == reflect.Array || current.Kind() == reflect.Slice {
					idx, err := strconv.Atoi(*part.index)
					if err != nil || idx < 0 || idx >= current.Len() {
						return nil, fmt.Errorf("invalid index '%s' for field '%s' in query '%s'", *part.index, part.name, query)
					}
					current = current.Index(idx)
				} else {
					return nil, fmt.Errorf("field '%s' is not an array or slice in query '%s'", part.name, query)
				}
			}

		case reflect.Map:
			if part.name == "" {
				return nil, fmt.Errorf("expected key for map in query '%s'", query)
			}
			mapKey := reflect.ValueOf(part.name)
			mapVal := current.MapIndex(mapKey)
			if !mapVal.IsValid() {
				return nil, fmt.Errorf("key '%s' not found in map for query '%s'", part.name, query)
			}
			current = mapVal

			if part.index != nil {
				if current.Kind() == reflect.Array || current.Kind() == reflect.Slice {
					idx, err := strconv.Atoi(*part.index)
					if err != nil || idx < 0 || idx >= current.Len() {
						return nil, fmt.Errorf("invalid index '%s' for key '%s' in query '%s'", *part.index, part.name, query)
					}
					current = current.Index(idx)
				} else {
					return nil, fmt.Errorf("value for key '%s' is not an array or slice in query '%s'", part.name, query)
				}
			}

		default:
			return nil, fmt.Errorf("unsupported kind '%s' in query '%s'", current.Kind(), query)
		}
	}

	switch current.Kind() {
	case reflect.Int32, reflect.Int, reflect.Int64:
		return current.Int(), nil
	case reflect.Float32, reflect.Float64:
		return RoundFloat(current.Float(), 6), nil
	case reflect.String:
		return current.String(), nil
	case reflect.Bool:
		return current.Bool(), nil
	default:
		return current.Interface(), nil
	}
}

// SetValue assigns a value to a struct field based on the query
func SetValue(structPtr interface{}, query string, val interface{}) error {
	parts := parseQueryPath(query)
	if len(parts) == 0 {
		return fmt.Errorf("invalid query: '%s'", query)
	}
	return assignField(structPtr, parts, val)
}

// SetValueUpdate sets a value and updates the INI file accordingly
func SetValueUpdate(obj interface{}, iniFile *ini.File, query string, value interface{}) error {
	err := SetValue(obj, query, value)
	if err != nil {
		return err
	}

	// Convert slices to comma-separated strings
	var valStr string
	switch v := value.(type) {
	case nil:
		valStr = "nil"
	case bool:
		if v {
			valStr = "1"
		} else {
			valStr = "0"
		}
	case []string:
		valStr = strings.Join(v, ", ")
	default:
		valStr = fmt.Sprintf("%v", v)
	}

	err = updateINIFile(obj, iniFile, query, valStr)
	if err != nil {
		return err
	}

	return nil
}

// SaveINI saves the INI file to the specified path
func SaveINI(iniFile *ini.File, filePath string) error {
	if iniFile == nil {
		return fmt.Errorf("iniFile is not initialized")
	}
	// Normalize all true/false to 1/0
	for _, section := range iniFile.Sections() {
		for _, key := range section.Keys() {
			if key.Value() == "true" {
				key.SetValue("1")
			} else if key.Value() == "false" {
				key.SetValue("0")
			}
		}
	}
	return iniFile.SaveTo(filePath)
}
