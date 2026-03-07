package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

// -------------------------------------------------------------------
// Language helpers
// -------------------------------------------------------------------

// SelectedLanguage returns the effective 2-letter language code to use.
func SelectedLanguage() string {
	lang := strings.ToLower(strings.TrimSpace(sys.cfg.Config.Language))
	if lang == "system" {
		s := strings.ToLower(strings.TrimSpace(osPreferredLanguage()))
		// Treat empty/C/POSIX as unset.
		if s == "" || s == "c" || s == "posix" {
			return "en"
		}
		// Extract 2-letter code from start (e.g., "en_US", "en-US.UTF-8").
		if len(s) >= 2 && s[0] >= 'a' && s[0] <= 'z' && s[1] >= 'a' && s[1] <= 'z' {
			lang = s[:2]
		} else {
			return "en"
		}
	}
	// Keep only first 2 letters if longer (e.g., "pt-BR" → "pt").
	if len(lang) >= 2 {
		lang = lang[:2]
	}
	if lang == "" {
		lang = "en"
	}
	return lang
}

// ResolveLangSectionName returns the section name to read, honoring language prefixes.
func ResolveLangSectionName(f *ini.File, section string, lang string) string {
	if f == nil || section == "" {
		return section
	}
	if lang = strings.ToLower(strings.TrimSpace(lang)); lang != "" {
		if _, err := f.GetSection(lang + "." + section); err == nil {
			return lang + "." + section
		}
	}
	if _, err := f.GetSection(section); err == nil {
		return section
	}
	if _, err := f.GetSection("en." + section); err == nil {
		return "en." + section
	}
	return section
}

// ResolveLangSectionNames returns section names in overlay order:
// base [Section] first, then [<lang>.Section] (if present).
func ResolveLangSectionNames(f *ini.File, section string, lang string) []string {
	if f == nil || section == "" {
		return []string{section}
	}
	lang = strings.ToLower(strings.TrimSpace(lang))
	out := make([]string, 0, 2)

	// Base section first (if it exists)
	if _, err := f.GetSection(section); err == nil {
		out = append(out, section)
	}

	// Language override section second (if it exists)
	if lang != "" {
		ls := lang + "." + section
		if _, err := f.GetSection(ls); err == nil {
			out = append(out, ls)
		}
	}

	// If neither exists, keep original name as a best-effort fallback.
	if len(out) == 0 {
		out = append(out, section)
	}
	return out
}

// pickLangSectionMerged returns a synthetic section that behaves like:
// [Section] with keys overwritten by [<lang>.Section] (if present).
func pickLangSectionMerged(f *ini.File, sec string) *ini.Section {
	if f == nil || sec == "" {
		return nil
	}
	lang := strings.ToLower(strings.TrimSpace(SelectedLanguage()))
	var baseSec, langSec *ini.Section
	if s, err := f.GetSection(sec); err == nil && s != nil {
		baseSec = s
	}
	if lang != "" {
		if s, err := f.GetSection(lang + "." + sec); err == nil && s != nil {
			langSec = s
		}
	}
	if baseSec == nil {
		return langSec
	}
	if langSec == nil {
		return baseSec
	}

	// Build a synthetic merged section: base keys first, then lang overrides.
	tmp := ini.Empty()
	merged, _ := tmp.NewSection(sec)
	for _, k := range baseSec.Keys() {
		_, _ = merged.NewKey(k.Name(), k.Value())
	}
	for _, k := range langSec.Keys() {
		if merged.HasKey(k.Name()) {
			mk, _ := merged.GetKey(k.Name())
			if mk != nil {
				mk.SetValue(k.Value())
			}
		} else {
			_, _ = merged.NewKey(k.Name(), k.Value())
		}
	}
	return merged
}

// pickLangSection returns the INI section to use for a logical section name,
// honoring language-specific overrides if present.
func pickLangSection(f *ini.File, sec string) *ini.Section {
	if f == nil || sec == "" {
		return nil
	}
	name := ResolveLangSectionName(f, sec, SelectedLanguage())
	if s, err := f.GetSection(name); err == nil && s != nil {
		return s
	}
	// Best-effort fallback.
	return f.Section(sec)
}

// normalize/strip a potential "xx." language prefix. Returns (lang, base, hasPrefix).
func splitLangPrefix(name string) (string, string, bool) {
	re := regexp.MustCompile(`(?i)^([a-z]{2})\.(.+)$`)
	if m := re.FindStringSubmatch(name); len(m) == 3 {
		return strings.ToLower(m[1]), m[2], true
	}
	return "", name, false
}

// -------------------------------------------------------------------
// Duplicate-key helpers (first instance wins)
// -------------------------------------------------------------------

// returns the first value for a key when duplicates exist (requires LoadOptions.AllowShadows = true).
// dupCount is the number of additional duplicate values ignored; handles empty-value edge cases in go-ini.
func iniFirstValue(k *ini.Key) (val string, dupCount int) {
	if k == nil {
		return "", 0
	}
	vs := k.ValueWithShadows()
	if len(vs) == 0 {
		// No shadows support or key had an empty value (go-ini quirk)
		return k.Value(), 0
	}
	// First-defined value is the first element.
	val = vs[0]
	if len(vs) > 1 {
		dupCount = len(vs) - 1
	}
	return val, dupCount
}

// overlayUserFirstWins overlays values from user into dst.
// If user contains duplicate keys (AllowShadows=true), the first occurrence wins.
func overlayUserFirstWins(dst, user *ini.File) {
	if dst == nil || user == nil {
		return
	}
	for _, us := range user.Sections() {
		ds := dst.Section(us.Name()) // creates if missing
		for _, uk := range us.Keys() {
			v, _ := iniFirstValue(uk)
			ds.Key(uk.Name()).SetValue(v)
		}
	}
}

// -------------------------------------------------------------------
// Helper Functions
// -------------------------------------------------------------------

// queryPart represents a single part of a query path
type queryPart struct {
	name  string
	index *string
}

// returns whether a map field should treat its keys case-insensitively.
// Defaults to true unless the struct tag explicitly disables it with insensitivekeys:"false".
func tagInsensitiveKeys(sf reflect.StructField) bool {
	v := strings.TrimSpace(strings.ToLower(sf.Tag.Get("insensitivekeys")))
	return !(v == "false" || v == "0" || v == "no")
}

// tagKeyFirstMap returns true if a map field opts into "key-first" query syntax:
//
//	"<key>.<mapField>[.<subfield>...]"  ->  "<mapField>.<key>[.<subfield>...]"
//
// Enable with: keyfirst:"true".
func tagKeyFirstMap(sf reflect.StructField) bool {
	v := strings.TrimSpace(strings.ToLower(sf.Tag.Get("keyfirst")))
	return v == "true" || v == "1" || v == "yes"
}

// returns true if root has a top-level ini:"<sec>" field with `literal:"yes"`.
func isLiteralSectionFor(root interface{}, sec string) bool {
	v := reflect.ValueOf(root)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	needle := strings.ToLower(sec)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if iniTag := strings.ToLower(f.Tag.Get("ini")); iniTag != "" && iniTag == needle {
			lit := strings.ToLower(f.Tag.Get("literal"))
			return lit == "yes" || lit == "true" || lit == "1"
		}
	}
	return false
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

// findFieldByINITag locates a field by its `ini` tag, supporting anonymous
// embedded structs. If no `ini` tag is present, it falls back to matching by field name.
func findFieldByINITag(v reflect.Value, tag string) (reflect.Value, reflect.StructField, bool) {
	val := v
	if !val.IsValid() {
		return reflect.Value{}, reflect.StructField{}, false
	}

	// Unwrap pointers
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return reflect.Value{}, reflect.StructField{}, false
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.StructField{}, false
	}

	typ := val.Type()

	// 1) Direct ini tag matches on this struct.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if iniTag := field.Tag.Get("ini"); iniTag != "" &&
			strings.EqualFold(iniTag, tag) {
			return val.Field(i), field, true
		}
	}

	// 2) Anonymous embedded structs: recurse.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" || !field.Anonymous {
			continue
		}
		fv := val.Field(i)
		if !fv.IsValid() {
			continue
		}
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			if rv, rf, found := findFieldByINITag(fv, tag); found {
				return rv, rf, true
			}
		}
	}
	// 3) Fallback: exported field name with no ini tag.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if field.Tag.Get("ini") == "" &&
			strings.EqualFold(field.Name, tag) {
			return val.Field(i), field, true
		}
	}

	return reflect.Value{}, reflect.StructField{}, false
}

// findMapFieldWithTag searches for a map field in a struct with a specific INI tag
func findMapFieldWithTag(v reflect.Value, tag string) (reflect.Value, reflect.StructField, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// skip unexported fields to avoid producing non-settable values
		if field.PkgPath != "" {
			continue
		}
		if strings.EqualFold(field.Tag.Get("ini"), tag) && field.Type.Kind() == reflect.Map {
			return v.Field(i), field, true
		}
	}
	return reflect.Value{}, reflect.StructField{}, false
}

// findDefaultField returns the first struct field without an INI tag
func findDefaultField(v reflect.Value) (reflect.Value, reflect.StructField, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// only consider exported fields
		if field.PkgPath != "" {
			continue
		}
		if field.Tag.Get("ini") == "" {
			return v.Field(i), field, true
		}
	}
	return reflect.Value{}, reflect.StructField{}, false
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

// buildLookupDirs expands a lookup tag like "def,,data/" into concrete dirs.
func buildLookupDirs(lookupTag string, baseDef string) []string {
	if lookupTag == "" {
		return nil
	}
	raw := strings.Split(lookupTag, ",")
	dirs := make([]string, 0, len(raw))
	for _, r := range raw {
		t := strings.TrimSpace(r)
		switch {
		case strings.EqualFold(t, "def"):
			dirs = append(dirs, baseDef)
		case t == "":
			dirs = append(dirs, "")
		default:
			dirs = append(dirs, t)
		}
	}
	return dirs
}

// resolveWithLookup runs SearchFile using dirs derived from lookupTag.
func resolveWithLookup(in string, lookupTag string, baseDef string) string {
	if lookupTag == "" {
		return in
	}
	dirs := buildLookupDirs(lookupTag, baseDef)
	if len(dirs) == 0 {
		return in
	}
	return SearchFile(in, dirs)
}

// setFieldValue assigns a value to a field based on its type, honoring default and lookup tags.
func setFieldValue(fieldVal reflect.Value, value interface{}, defTag string, keyPath string, lookupTag string, baseDef string) error {
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
				if err := setFieldValue(elem, v, defTag, keyPath, lookupTag, baseDef); err != nil {
					return err
				}
				newSlice = reflect.Append(newSlice, elem)
			}
			fieldVal.Set(newSlice)
			return nil
		case reflect.Array:
			for i := 0; i < fieldVal.Len() && i < len(arr); i++ {
				elem := fieldVal.Index(i)
				if err := setFieldValue(elem, arr[i], defTag, keyPath, lookupTag, baseDef); err != nil {
					return err
				}
			}
			return nil
		default:
			// if there's at least one element, treat arr[0] as the final string
			if len(arr) > 0 {
				return setFieldValue(fieldVal, arr[0], defTag, keyPath, lookupTag, baseDef)
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

	// Normalize legacy key syntax: replace & separator with commas
	if strings.HasSuffix(strings.ToLower(keyPath), ".key") {
		trimmedValue = strings.ReplaceAll(trimmedValue, "&", ",")
	}

	// Unwrap pointer
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			applyDefaultsToValue(fieldVal)
		}
		fieldVal = fieldVal.Elem()
	}

	// Empty INI value semantics:
	// - Scalars: treat as "unset" (do not override whatever is already there, e.g. defaults)
	// - Arrays/Slices: treat as "clear" (override defaults with no entries)
	if trimmedValue == "" {
		switch fieldVal.Kind() {
		case reflect.Slice, reflect.Array:
			// Explicitly clear list/array
			setNil(fieldVal)
			return nil
		default:
			// Leave existing
			if defTag == "" {
				return nil
			}
			// If a struct-level default tag exists, apply it
			trimmedValue = defTag
		}
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
			if err := setFieldValue(elem, strings.TrimSpace(p), "", keyPath, lookupTag, baseDef); err != nil {
				return err
			}
			newSlice = reflect.Append(newSlice, elem)
		}
		fieldVal.Set(newSlice)

	case reflect.Array:
		separator := ","
		parts := strings.Split(trimmedValue, separator)

		// For font arrays, the first element may be a path instead of a numeric index.
		// In that case we keep the index as -1 so it can be resolved later by the inline font loader.
		lkp := strings.ToLower(keyPath)
		isFontArray := strings.HasSuffix(lkp, ".font") || lkp == "font"
		if isFontArray && len(parts) > 0 {
			first := strings.TrimSpace(parts[0])
			if first != "" {
				if _, err := strconv.ParseInt(first, 10, 64); err != nil {
					// Non-numeric font id (inline font path) – keep index as -1,it will be resolved later.
					parts[0] = "-1"
				}
			}
		}

		// Treat the assignment as partial when not all array elements are provided or some entries are blank.
		// For such partial updates we rebuild the array from default tag/zeros instead of keeping older values (e.g. from defaultMotif.ini)
		isPartial := len(parts) < fieldVal.Len()
		if !isPartial {
			for _, raw := range parts {
				if strings.TrimSpace(raw) == "" {
					isPartial = true
					break
				}
			}
		}

		if isPartial {
			if defTag != "" {
				// Rebase the whole array to the struct-level `default` tag string.
				if err := setFieldValue(fieldVal, defTag, "", keyPath, lookupTag, baseDef); err != nil {
					return err
				}
			} else {
				// If there is no struct `default`, rebase the array to zero values.
				for i := 0; i < fieldVal.Len(); i++ {
					elem := fieldVal.Index(i)
					elem.Set(reflect.Zero(elem.Type()))
				}
			}
		}

		// Apply user / source values on top of the freshly rebased array.
		for i := 0; i < fieldVal.Len(); i++ {
			var p string
			if i < len(parts) {
				p = strings.TrimSpace(parts[i])
			}
			if p == "" {
				// leave existing element untouched
				continue
			}
			elem := fieldVal.Index(i)
			if err := setFieldValue(elem, p, "", keyPath, lookupTag, baseDef); err != nil {
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
		// If a lookup tag is present on this field, resolve the path immediately.
		if lookupTag != "" {
			parsed = resolveWithLookup(parsed, lookupTag, baseDef)
		}
		fieldVal.SetString(parsed)

	case reflect.Struct:
		// If the struct has a default field, set its value
		defaultField, defaultSf, found := findDefaultField(fieldVal)
		if !found {
			applyDefaultsToValue(fieldVal)
			return nil
		}
		// Use the default field's own lookup tag (if any) when setting it.
		dfLookup := defaultSf.Tag.Get("lookup")
		if err := setFieldValue(defaultField, trimmedValue, "", keyPath, dfLookup, baseDef); err != nil {
			return err
		}
		applyDefaultsToValue(fieldVal)

	default:
		return fmt.Errorf("unsupported field kind: %s", fieldVal.Kind())
	}

	return nil
}

func assignToPatternMap(v reflect.Value, lastPartName string, value interface{}, final bool, baseDef string) (bool, reflect.Value, error) {
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
					if final {
						var newVal reflect.Value
						if elemType.Kind() == reflect.Ptr {
							if mapElem.IsValid() {
								newVal = reflect.New(elemType.Elem())
								newVal.Elem().Set(mapElem.Elem())
							} else {
								newVal = reflect.New(elemType.Elem())
							}
						} else {
							if mapElem.IsValid() {
								newVal = reflect.New(elemType).Elem()
								newVal.Set(mapElem)
							} else {
								newVal = reflect.New(elemType).Elem()
							}
						}
						applyDefaultsToValue(newVal)
						// For default-field structs, setFieldValue will pick up that field's lookup tag itself.
						if err := setFieldValue(newVal, value, "", lastPartName, "", baseDef); err != nil {
							return false, v, err
						}
						fieldVal.SetMapIndex(mapKey, newVal)
						mapElem = fieldVal.MapIndex(mapKey)
					} else {
						if !mapElem.IsValid() {
							var newVal reflect.Value
							if elemType.Kind() == reflect.Ptr {
								newVal = reflect.New(elemType.Elem())
							} else {
								newVal = reflect.New(elemType).Elem()
							}
							applyDefaultsToValue(newVal)
							fieldVal.SetMapIndex(mapKey, newVal)
							mapElem = newVal
						} else {
							// Ensure that the value we work with is addressable.
							var newVal reflect.Value
							if elemType.Kind() == reflect.Ptr {
								newVal = reflect.New(elemType.Elem())
								newVal.Elem().Set(mapElem.Elem())
							} else {
								newVal = reflect.New(elemType).Elem()
								newVal.Set(mapElem)
							}
							fieldVal.SetMapIndex(mapKey, newVal)
							mapElem = newVal
						}
					}
					return true, mapElem, nil
				}
			}
		}
	}
	return false, v, nil
}

// assignToFlattenMapField handles map fields with `flatten:"true"`.
func assignToFlattenMapField(fieldVal reflect.Value, sf reflect.StructField, parts []queryPart, currentIndex int, value interface{},
	baseDef string, extractNames func([]queryPart) []string) (bool, error) {

	flattenTag := strings.TrimSpace(strings.ToLower(sf.Tag.Get("flatten")))
	if !(flattenTag == "true" || flattenTag == "1" || flattenTag == "yes") {
		return false, nil
	}

	// Need at least: bg.<key>.field
	if len(parts) < currentIndex+3 {
		return false, nil
	}

	// Parts layout (indices):
	//   currentIndex   -> "bg"
	//   keyStart..keyEnd -> key segments to flatten with "_"
	//   propIdx        -> final struct field name (e.g. "spr", "offset", ...)
	keyStart := currentIndex + 1
	keyEnd := len(parts) - 2
	propIdx := len(parts) - 1

	if keyStart > keyEnd {
		return false, nil
	}

	// Build flattened key: menunetwork.serverhost -> "menunetwork_serverhost"
	keySegments := make([]string, 0, keyEnd-keyStart+1)
	for i := keyStart; i <= keyEnd; i++ {
		keySegments = append(keySegments, parts[i].name)
	}
	mapKey := strings.Join(keySegments, "_")
	if mapKey == "" {
		return false, nil
	}

	if fieldVal.IsNil() {
		fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
	}

	mapKeyVal := reflect.ValueOf(mapKey)
	elemType := fieldVal.Type().Elem()
	mapElem := fieldVal.MapIndex(mapKeyVal)

	var elem reflect.Value
	if mapElem.IsValid() {
		// Reuse existing element; for pointer types we can take it directly.
		if elemType.Kind() == reflect.Ptr {
			elem = mapElem
		} else {
			elem = reflect.New(elemType).Elem()
			elem.Set(mapElem)
		}
	} else {
		// Create new element with defaults.
		if elemType.Kind() == reflect.Ptr {
			elem = reflect.New(elemType.Elem())
		} else {
			elem = reflect.New(elemType).Elem()
		}
		applyDefaultsToValue(elem)
	}

	// Dereference pointer to get the underlying struct.
	target := elem
	for target.Kind() == reflect.Ptr {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
			applyDefaultsToValue(target)
		}
		target = target.Elem()
	}

	if target.Kind() != reflect.Struct {
		return false, fmt.Errorf("flatten map element for '%s' is not a struct", sf.Name)
	}

	propPart := parts[propIdx]
	fieldVal2, fieldType, foundField := findFieldByINITag(target, propPart.name)
	if !foundField {
		return false, fmt.Errorf("field '%s' not found in flattened map element '%s'", propPart.name, mapKey)
	}

	defTag := fieldType.Tag.Get("default")
	lookupTag := fieldType.Tag.Get("lookup")
	keyPath := strings.Join(extractNames(parts), ".")

	if err := setFieldValue(fieldVal2, value, defTag, keyPath, lookupTag, baseDef); err != nil {
		return false, err
	}

	// Store back into the map.
	fieldVal.SetMapIndex(mapKeyVal, elem)

	return true, nil
}

// elemHasFieldWithINITag reports whether elemType (or its Elem for pointers)
// has an exported field whose ini:"..." tag equals 'name' (case-insensitive).
func elemHasFieldWithINITag(elemType reflect.Type, name string) bool {
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return false
	}
	// Use the same resolution logic as findFieldByINITag, so this works
	// even when fields come from anonymously embedded structs.
	dummy := reflect.New(elemType).Elem()
	_, _, found := findFieldByINITag(dummy, name)
	return found
}

// assignField assigns a value to a struct field based on query parts
func assignField(structPtr interface{}, parts []queryPart, value interface{}, baseDef string) error {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		fmt.Println("Error: structPtr must be a non-nil pointer")
		return fmt.Errorf("structPtr must be a non-nil pointer")
	}

	// Tracks whether the *current* map field (if any) should use case-insensitive keys.
	currentMapInsensitiveKeys := true

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
				// If the matched field is a MAP and we're at the final token, assign to the map["default"] element.
				if fieldVal.Kind() == reflect.Map {
					if fieldVal.IsNil() {
						fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
					}
					if fieldVal.Type().Key().Kind() != reflect.String {
						return true, fmt.Errorf("map key type must be string for '%s'", part.name)
					}
					defKey := reflect.ValueOf("default")
					elemType := fieldVal.Type().Elem()
					var newVal reflect.Value
					if elemType.Kind() == reflect.Ptr {
						newVal = reflect.New(elemType.Elem())
					} else {
						newVal = reflect.New(elemType).Elem()
					}
					applyDefaultsToValue(newVal)
					// setFieldValue will parse the string into the element (string/struct/etc)
					if err := setFieldValue(newVal, value, "", strings.Join(extractNames(parts), "."), "", baseDef); err != nil {
						return true, err
					}
					fieldVal.SetMapIndex(defKey, newVal)
					return true, nil
				}

				defTag := fieldType.Tag.Get("default") // read the struct's default tag
				lookupTag := fieldType.Tag.Get("lookup")
				if err := setFieldValue(fieldVal, value, defTag, strings.Join(extractNames(parts), "."), lookupTag, baseDef); err != nil {
					return true, err
				}
				// Support for preload:"char" / preload:"stage" / preload:"pal"
				preload := strings.ToLower(fieldType.Tag.Get("preload"))
				switch preload {
				case "char", "stage":
					switch fieldType.Name {
					case "Anim":
						if fieldVal.Kind() == reflect.Int32 {
							av := int32(fieldVal.Int())
							if av >= 0 {
								if preload == "char" {
									sys.sel.charAnimPreload[av] = true
								} else {
									sys.sel.stageAnimPreload[av] = true
								}
							}
						}
					case "Spr":
						// Expect a [2]int32 array
						if fieldVal.Kind() == reflect.Array && fieldVal.Len() == 2 {
							// Read both elements as int32
							a0 := uint16(fieldVal.Index(0).Int())
							a1 := uint16(fieldVal.Index(1).Int())
							if a0 >= 0 && a1 >= 0 {
								key := [2]uint16{a0, a1}
								if preload == "char" {
									sys.sel.charSpritePreload[key] = true
								} else {
									sys.sel.stageSpritePreload[key] = true
								}
							}
						}
					}
				case "pal":
					// When a parameter tagged with preload:"pal" is true, enable palette-based loading globally.
					if fieldVal.Kind() == reflect.Bool && fieldVal.Bool() {
						sys.usePalette = true
					}
				}
				return true, nil
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

				assigned, _, err = assignToPatternMap(v, part.name, value, true, baseDef)
				if err != nil {
					return err
				}
				if assigned {
					return nil
				}

				return fmt.Errorf("field '%s' not found in struct or pattern map field", part.name)
			} else if v.Kind() == reflect.Map {
				if v.IsNil() {
					if !v.CanSet() {
						return fmt.Errorf("cannot assign into nil map at '%s' (map is not settable)", strings.Join(extractNames(parts), "."))
					}
					v.Set(reflect.MakeMap(v.Type()))
				}
				if value == nil {
					keyName := part.name
					if currentMapInsensitiveKeys {
						keyName = strings.ToLower(keyName)
					}
					mapKey := reflect.ValueOf(keyName)
					v.SetMapIndex(mapKey, reflect.Value{})
					return nil
				}
				keyName := part.name
				if currentMapInsensitiveKeys {
					keyName = strings.ToLower(keyName)
				}
				mapKey := reflect.ValueOf(keyName)
				elemType := v.Type().Elem()

				var newVal reflect.Value
				if elemType.Kind() == reflect.Ptr {
					newVal = reflect.New(elemType.Elem())
				} else {
					newVal = reflect.New(elemType).Elem()
				}
				applyDefaultsToValue(newVal)

				if err := setFieldValue(newVal, value, "", strings.Join(extractNames(parts), "."), "", baseDef); err != nil {
					return err
				}
				v.SetMapIndex(mapKey, newVal)
				return nil
			} else {
				return fmt.Errorf("cannot set value on non-struct and non-map kind '%s'", v.Kind())
			}
		}

		if v.Kind() == reflect.Struct {
			fieldVal, sf, found := findFieldByINITag(v, part.name)
			if found {
				if fieldVal.Kind() == reflect.Map {
					// First, try flattened map semantics if requested by tag.
					if ok, err := assignToFlattenMapField(fieldVal, sf, parts, i, value, baseDef, extractNames); err != nil {
						return err
					} else if ok {

						return nil
					}

					// Normal map-of-struct behavior (using "default" key).
					currentMapInsensitiveKeys = tagInsensitiveKeys(sf)

					if i+1 < len(parts) {
						elemType := fieldVal.Type().Elem()
						if elemHasFieldWithINITag(elemType, parts[i+1].name) {
							if fieldVal.IsNil() {
								fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
							}
							defKey := reflect.ValueOf("default")
							mapElem := fieldVal.MapIndex(defKey)
							var newVal reflect.Value
							if !mapElem.IsValid() {
								if elemType.Kind() == reflect.Ptr {
									newVal = reflect.New(elemType.Elem())
								} else {
									newVal = reflect.New(elemType).Elem()
								}
								applyDefaultsToValue(newVal)
								fieldVal.SetMapIndex(defKey, newVal)

								v = newVal
							} else {

								if elemType.Kind() == reflect.Ptr {
									newVal = reflect.New(elemType.Elem())
									newVal.Elem().Set(mapElem.Elem())
								} else {
									newVal = reflect.New(elemType).Elem()
									newVal.Set(mapElem)
								}
								fieldVal.SetMapIndex(defKey, newVal)
								v = newVal
							}
							continue
						}
					}
				}
				// normal struct field
				v = fieldVal
				continue
			}

			mapFieldVal, mapSf, found := findMapFieldWithTag(v, part.name)
			if found && mapFieldVal.Kind() == reflect.Map {
				// Record key-sensitivity for this map field.
				mapInsensitive := tagInsensitiveKeys(mapSf)
				if i+1 >= len(parts) {
					return fmt.Errorf("expected map key after '%s'", part.name)
				}
				// Decide whether parts[i+1] is a real map key, or an element-field → default
				elemType := mapFieldVal.Type().Elem()
				if elemHasFieldWithINITag(elemType, parts[i+1].name) {
					currentMapInsensitiveKeys = mapInsensitive
					// default element
					if mapFieldVal.IsNil() {
						mapFieldVal.Set(reflect.MakeMap(mapFieldVal.Type()))
					}
					defKey := reflect.ValueOf("default")
					mapElem := mapFieldVal.MapIndex(defKey)
					var elem reflect.Value
					if mapElem.IsValid() {
						if elemType.Kind() == reflect.Ptr {
							elem = reflect.New(elemType.Elem())
							elem.Elem().Set(mapElem.Elem())
						} else {
							elem = reflect.New(elemType).Elem()
							elem.Set(mapElem)
						}
					} else {
						if elemType.Kind() == reflect.Ptr {
							elem = reflect.New(elemType.Elem())
						} else {
							elem = reflect.New(elemType).Elem()
						}
						applyDefaultsToValue(elem)
					}
					mapFieldVal.SetMapIndex(defKey, elem)
					v = elem
					// do NOT consume parts[i+1] here; it belongs to the element
					continue
				}
				// explicit map key path
				mapKey := parts[i+1].name
				if mapInsensitive {
					mapKey = strings.ToLower(mapKey)
				}
				if mapFieldVal.IsNil() {
					mapFieldVal.Set(reflect.MakeMap(mapFieldVal.Type()))
				}
				mapKeyVal := reflect.ValueOf(mapKey)
				elemType2 := mapFieldVal.Type().Elem()
				mapElem := mapFieldVal.MapIndex(mapKeyVal)
				var elem reflect.Value
				if mapElem.IsValid() {
					elem = mapElem
				} else {
					if elemType2.Kind() == reflect.Ptr {
						elem = reflect.New(elemType2.Elem())
					} else {
						elem = reflect.New(elemType2).Elem()
					}
					applyDefaultsToValue(elem)
					mapFieldVal.SetMapIndex(mapKeyVal, elem)
				}
				v = elem
				i++
				continue
			}

			assigned, newV, err := assignToPatternMap(v, part.name, value, false, baseDef)
			if err != nil {
				return err
			}
			if assigned {
				v = newV
				continue
			}

			// Handle key-first map fields:
			// rewrite "<key>.<mapField>[...]" to "<mapField>.<key>[...]" when the map field has `keyfirst:"true"`.
			if i+1 < len(parts) {
				if fieldValMap, sfMap, foundMap := findFieldByINITag(v, parts[i+1].name); foundMap &&
					fieldValMap.Kind() == reflect.Map &&
					fieldValMap.Type().Key().Kind() == reflect.String &&
					tagKeyFirstMap(sfMap) {
					newParts := make([]queryPart, len(parts))
					copy(newParts, parts)
					newParts[i], newParts[i+1] = newParts[i+1], newParts[i]
					return assignField(structPtr, newParts, value, baseDef)
				}
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

// resolveSectionForWrite tries to map an internal/tag section name (using underscores instead of spaces)
// onto the actual section name present in the INI (which can have spaces).
// If a matching existing section is found, it returns that canonical name.
// Otherwise it returns a best-effort name for creation (preferring spaces).
func resolveSectionForWrite(f *ini.File, sec string) string {
	if f == nil || sec == "" {
		return sec
	}
	// Exact name (case-insensitive lookup is handled by ini when enabled)
	if _, err := f.GetSection(sec); err == nil {
		// Keep the canonical stored name (preserves original casing/spaces)
		s, _ := f.GetSection(sec)
		if s != nil {
			return s.Name()
		}
		return sec
	}
	// Underscore -> space fallback
	if strings.Contains(sec, "_") {
		alt := strings.ReplaceAll(sec, "_", " ")
		if s, err := f.GetSection(alt); err == nil && s != nil {
			return s.Name()
		}
		// Prefer creating the spaced name so we don't split the config across two sections.
		return alt
	}
	return sec
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
		}

		switch currentValue.Kind() {
		case reflect.Struct:
			// Try to resolve as a normal struct field (ini tag / embedded aware).
			fieldVal, structField, found := findFieldByINITag(currentValue, part.name)
			if found {
				// Map-of-struct special case: treat "map.field" as "map.default.field"
				if fieldVal.Kind() == reflect.Map && i+1 < len(parts) {
					elemType := fieldVal.Type().Elem()
					if elemHasFieldWithINITag(elemType, parts[i+1].name) {
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
						// append "default" and the remainder (element field path)
						keyNameParts = append(keyNameParts, "default")
						for j := i + 1; j < len(parts); j++ {
							keyNameParts = append(keyNameParts, parts[j].name)
						}
						i = len(parts)
						goto finalize
					}
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

				if part.index != nil {
					if currentValue.Kind() == reflect.Array || currentValue.Kind() == reflect.Slice {
						idx, err := strconv.Atoi(*part.index)
						if err != nil || idx < 0 || idx >= currentValue.Len() {
							return fmt.Errorf("invalid index '%s' for field '%s' in query '%s'", *part.index, part.name, query)
						}
						currentValue = currentValue.Index(idx)
					} else {
						return fmt.Errorf("field '%s' is not an array or slice in query '%s'", part.name, query)
					}
				}
				i++

			} else {
				// Key-first map syntax support:
				// If the next token names a map field tagged `keyfirst:"true"`, swap and retry.
				// Example: "<key>.text" -> "text.<key>" (but generalized to any tagged map field).
				if i+1 < len(parts) {
					if fvMap, sfMap, ok := findFieldByINITag(currentValue, parts[i+1].name); ok &&
						fvMap.Kind() == reflect.Map &&
						fvMap.Type().Key().Kind() == reflect.String &&
						tagKeyFirstMap(sfMap) {
						parts[i], parts[i+1] = parts[i+1], parts[i]
						continue
					}
				}
				// Not a direct field -> check if it's a direct map field with the same tag
				mapFieldVal, _, mapFound := findMapFieldWithTag(currentValue, part.name)
				if mapFound && mapFieldVal.Kind() == reflect.Map {
					if i+1 >= len(parts) {
						return fmt.Errorf("expected map key after '%s'", part.name)
					}
					// Decide default vs explicit
					elemType := mapFieldVal.Type().Elem()
					if elemHasFieldWithINITag(elemType, parts[i+1].name) {
						if !sectionSet {
							sectionName = part.name
							sectionSet = true
						} else {
							keyNameParts = append(keyNameParts, part.name)
						}
						// default element; include default and the remainder as keys
						keyNameParts = append(keyNameParts, "default")
						for j := i + 1; j < len(parts); j++ {
							keyNameParts = append(keyNameParts, parts[j].name)
						}
						i = len(parts)
						goto finalize
					} else {
						// explicit map key
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
					}

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

	// Map internal/tag section names to the real INI section name.
	sectionName = resolveSectionForWrite(iniFile, sectionName)
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
					// Key-first map syntax support (read path):
					// If the next token names a map field tagged `keyfirst:"true"`, swap and retry.
					if i+1 < len(parts) {
						if fvMap, sfMap, ok := findFieldByINITag(current, parts[i+1].name); ok &&
							fvMap.Kind() == reflect.Map &&
							fvMap.Type().Key().Kind() == reflect.String &&
							tagKeyFirstMap(sfMap) {
							parts[i], parts[i+1] = parts[i+1], parts[i]
							i--
							continue
						}
					}
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
	// Try to extract a base "Def" path from the target struct.
	extractDef := func(ptr interface{}) string {
		v := reflect.ValueOf(ptr)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			f := v.FieldByName("Def")
			if f.IsValid() && f.Kind() == reflect.String {
				// Normalize to slash form to match existing SearchFile usage.
				return filepath.ToSlash(f.String())
			}
		}
		return ""
	}
	baseDef := extractDef(structPtr)

	return assignField(structPtr, parts, val, baseDef)
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

// -------------------------------------------------------------------
// Font utilities
// -------------------------------------------------------------------

// syncFontsMap ensures the destination FontProperties map mirrors the actual
// loaded fonts (Fnt) and any inline-defined fonts that were appended via
// resolveInlineFonts. It fills filename/height when possible and copies
// extended properties from the Fnt into FontProperties.
func syncFontsMap(dst *map[string]*FontProperties, fonts map[int]*Fnt, fntIndexByKey map[string]int) {
	if *dst == nil {
		*dst = make(map[string]*FontProperties)
	}
	// Build reverse: index -> "path|height" (keep first encountered per index)
	reverse := make(map[int]string)
	for k, idx := range fntIndexByKey {
		if _, ok := reverse[idx]; !ok {
			reverse[idx] = k
		}
	}
	for idx, f := range fonts {
		key := fmt.Sprintf("font%d", idx)
		fp, ok := (*dst)[key]
		if !ok {
			fp = &FontProperties{Height: -1}
			(*dst)[key] = fp
		}
		// Recover filename and height from reverse map ("path|height")
		if s, ok := reverse[idx]; ok && s != "" {
			parts := strings.SplitN(s, "|", 2)
			if len(parts) > 0 && parts[0] != "" {
				fp.Font = parts[0]
			}
			if len(parts) == 2 {
				fp.Height = int32(Atoi(parts[1]))
			}
		}
		// Copy extended properties from the actual font
		if f != nil {
			fp.Type = f.Type
			fp.Size = f.Size
			fp.Spacing = f.Spacing
			fp.Offset = f.offset
		}
	}
}

// -------------------------------------------------------------------
// Generic Setter Functions
// -------------------------------------------------------------------

// getFieldFromHierarchy finds field name on primary or, if missing, on parent.
// This lets SetAnim/SetTextSprite/etc see both embedded and wrapper fields.
func getFieldFromHierarchy(primary, parent reflect.Value, name string) (reflect.Value, bool) {
	// Unwrap pointers.
	for primary.IsValid() && primary.Kind() == reflect.Ptr && !primary.IsNil() {
		primary = primary.Elem()
	}
	if primary.IsValid() && primary.Kind() == reflect.Struct {
		if fv := primary.FieldByName(name); fv.IsValid() {
			return fv, true
		}
	}

	for parent.IsValid() && parent.Kind() == reflect.Ptr && !parent.IsNil() {
		parent = parent.Elem()
	}
	if parent.IsValid() && parent.Kind() == reflect.Struct {
		if fv := parent.FieldByName(name); fv.IsValid() {
			return fv, true
		}
	}
	return reflect.Value{}, false
}

// SetAnim sets the Anim field generically for any struct that contains AnimTable and Sff fields.
func SetAnim(obj interface{}, fVal, structVal, parent reflect.Value, sffOverride *Sff) {
	anim := int32(-1)
	spr := [2]int32{-1, 0}
	offset := [2]float32{0, 0}
	scale := [2]float32{1, 1}
	facing := float32(1)
	xshear := float32(0)
	angle := float32(0)
	xangle := float32(0)
	yangle := float32(0)
	focallength := float32(2048)
	projection := int32(0)
	layerno := int16(1)
	localcoord := [2]float32{0, 0}
	window := [4]float32{0, 0, 0, 0}
	velocity := [2]float32{0, 0}
	maxDist := [2]float32{0, 0}
	accel := [2]float32{0, 0}
	friction := [2]float32{1, 1}
	hasSpr := false

	get := func(name string) (reflect.Value, bool) {
		return getFieldFromHierarchy(structVal, parent, name)
	}

	if fv, ok := get("Anim"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		anim = int32(fv.Int())
	}
	if fv, ok := get("Spr"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		spr[0] = int32(fv.Index(0).Int())
		spr[1] = int32(fv.Index(1).Int())
		hasSpr = true
	}
	if fv, ok := get("Offset"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		offset[0] = float32(fv.Index(0).Float())
		offset[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Scale"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		scale[0] = float32(fv.Index(0).Float())
		scale[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Facing"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		facing = float32(fv.Int())
	}
	if fv, ok := get("Xshear"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		xshear = float32(fv.Float())
	}
	if fv, ok := get("Angle"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		angle = float32(fv.Float())
	}
	if fv, ok := get("XAngle"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		xangle = float32(fv.Float())
	}
	if fv, ok := get("YAngle"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		yangle = float32(fv.Float())
	}
	if fv, ok := get("Focallength"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		focallength = float32(fv.Float())
	}
	if fv, ok := get("Projection"); ok {
		switch fv.Kind() {

		case reflect.String:
			switch strings.ToLower(strings.TrimSpace(fv.String())) {
			case "orthographic":
				projection = int32(Projection_Orthographic)
			case "perspective":
				projection = int32(Projection_Perspective)
			case "perspective2":
				projection = int32(Projection_Perspective2)
			default:
				if i, err := strconv.Atoi(fv.String()); err == nil {
					projection = int32(i)
				}
			}
		case reflect.Int32, reflect.Int64:
			projection = int32(fv.Int())
		}
	}
	if fv, ok := get("Layerno"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		layerno = int16(fv.Int())
	}
	if fv, ok := get("Window"); ok && fv.Kind() == reflect.Array && fv.Len() == 4 {
		for k := 0; k < 4; k++ {
			window[k] = float32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("Localcoord"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		for k := 0; k < 2; k++ {
			localcoord[k] = float32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("Velocity"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		velocity[0] = float32(fv.Index(0).Float())
		velocity[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("MaxDist"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		maxDist[0] = float32(fv.Index(0).Float())
		maxDist[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Accel"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		accel[0] = float32(fv.Index(0).Float())
		accel[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Friction"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		friction[0] = float32(fv.Index(0).Float())
		friction[1] = float32(fv.Index(1).Float())
	}
	/*
		We still keep the original structVal.Localcoord fallback for when the
		localcoord field is zeroed; handled in PopulateDataPointers.
	*/

	// Access obj fields generically
	objVal := reflect.ValueOf(obj).Elem()
	atField := objVal.FieldByName("AnimTable")

	// Resolve SFF to use: override if provided, else obj.Sff
	var sffPtr *Sff
	if sffOverride != nil {
		sffPtr = sffOverride
	} else {
		if sf := objVal.FieldByName("Sff"); sf.IsValid() && !sf.IsNil() {
			if s, ok := sf.Interface().(*Sff); ok {
				sffPtr = s
			}
		}
	}

	if !atField.IsValid() {
		fmt.Println("Error: object does not have required AnimTable or Sff fields")
		return
	}

	animMap, ok := atField.Interface().(AnimationTable)
	if !ok {
		fmt.Println("Error: AnimTable field is not of type AnimationTable")
		return
	}

	// animNew
	a := NewAnim(sffPtr, fmt.Sprintf("%d,%d, 0,0, -1", spr[0], spr[1]))
	if animData, exists := animMap[anim]; exists {
		a.anim = animData
	} else if hasSpr {
		//a.anim.SetAnimElem(1, 0)
		a.anim.UpdateSprite()
	} else {
		a.anim = nil
	}
	// animSetLocalcoord
	a.SetLocalcoord(localcoord[0], localcoord[1])
	// animSetPos
	a.SetPos(offset[0], offset[1])
	// animSetScale
	a.SetScale(scale[0], scale[1])
	// animSetFacing
	a.facing = facing
	// animSetXShear
	a.xshear = xshear
	// animSetAngle
	a.rot.angle = angle
	// animSetXAngle
	a.rot.xangle = xangle
	// animSetYAngle
	a.rot.yangle = yangle
	// animSetProjection
	a.projection = projection
	// animSetFocalLength
	a.fLength = focallength
	// animSetWindow
	a.SetWindow(window)
	// animSetLayerno
	a.layerno = layerno
	// animSetVelocity
	a.SetVelocity(velocity[0], velocity[1])
	// animSetMaxDist
	a.SetMaxDist(maxDist[0], maxDist[1])
	// animSetAccel
	a.SetAccel(accel[0], accel[1])
	// animSetFriction
	a.friction = friction
	// animSetTile
	// animSetAlpha
	// animSetColorKey
	// animSetPalFX

	// Propagate sprite size back into the config struct if possible.
	// Use the "primary" structVal here (the struct that owns the
	// pointer field) so this still works for embedded wrappers.
	for structVal.IsValid() && structVal.Kind() == reflect.Ptr && !structVal.IsNil() {
		structVal = structVal.Elem()
	}
	if a.anim != nil && a.anim.spr != nil && structVal.IsValid() && structVal.CanAddr() {
		if f := structVal.FieldByName("Size"); f.IsValid() && f.CanSet() && f.Kind() == reflect.Array && f.Len() == 2 {
			f.Index(0).SetInt(int64(a.anim.spr.Size[0]))
			f.Index(1).SetInt(int64(a.anim.spr.Size[1]))
		}
	}

	fVal.Set(reflect.ValueOf(a))
}

// SetTextSprite sets the TextSprite field generically for any struct.
func SetTextSprite(obj interface{}, fVal, structVal, parent reflect.Value) {
	offset := [2]float32{0, 0}
	font := [8]int32{-1, 0, 0, 255, 255, 255, 255, -1}
	scale := [2]float32{1, 1}
	xshear := float32(0)
	angle := float32(0)
	xangle := float32(0)
	yangle := float32(0)
	focallength := float32(2048)
	projection := int32(0)
	text := ""
	layerno := int16(1)
	localcoord := [2]float32{0, 0}
	window := [4]float32{0, 0, 0, 0}
	textDelay := float32(0)
	textSpacing := [2]float32{0, 0}
	textWrap := ""
	velocity := [2]float32{0, 0}
	maxDist := [2]float32{0, 0}
	accel := [2]float32{0, 0}
	friction := [2]float32{1, 1}

	get := func(name string) (reflect.Value, bool) {
		return getFieldFromHierarchy(structVal, parent, name)
	}

	if fv, ok := get("Offset"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		offset[0] = float32(fv.Index(0).Float())
		offset[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Font"); ok && fv.Kind() == reflect.Array && fv.Len() == 8 {
		for k := 0; k < 8; k++ {
			font[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("Scale"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		scale[0] = float32(fv.Index(0).Float())
		scale[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Xshear"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		xshear = float32(fv.Float())
	}
	if fv, ok := get("Angle"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		angle = float32(fv.Float())
	}
	if fv, ok := get("XAngle"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		xangle = float32(fv.Float())
	}
	if fv, ok := get("YAngle"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		yangle = float32(fv.Float())
	}
	if fv, ok := get("Focallength"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		focallength = float32(fv.Float())
	}
	if fv, ok := get("Projection"); ok {
		switch fv.Kind() {

		case reflect.String:
			switch strings.ToLower(strings.TrimSpace(fv.String())) {
			case "orthographic":
				projection = int32(Projection_Orthographic)
			case "perspective":
				projection = int32(Projection_Perspective)
			case "perspective2":
				projection = int32(Projection_Perspective2)
			default:
				if i, err := strconv.Atoi(fv.String()); err == nil {
					projection = int32(i)
				}
			}
		case reflect.Int32, reflect.Int64:
			projection = int32(fv.Int())
		}
	}
	if fv, ok := get("Text"); ok && fv.Kind() == reflect.String {
		text = fv.String()
	}
	if fv, ok := get("Layerno"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		layerno = int16(fv.Int())
	}
	if fv, ok := get("Window"); ok && fv.Kind() == reflect.Array && fv.Len() == 4 {
		for k := 0; k < 4; k++ {
			window[k] = float32(fv.Index(k).Int())
		}
	}
	// Some callers use a dedicated TextWindow field; let it override Window.
	if fv, ok := get("TextWindow"); ok && fv.Kind() == reflect.Array && fv.Len() == 4 {
		for k := 0; k < 4; k++ {
			window[k] = float32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("Localcoord"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		localcoord[0] = float32(fv.Index(0).Int())
		localcoord[1] = float32(fv.Index(1).Int())
	}
	if fv, ok := get("TextDelay"); ok {
		switch fv.Kind() {
		case reflect.Float32, reflect.Float64:
			textDelay = float32(fv.Float())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			textDelay = float32(fv.Int())
		}
	}
	if fv, ok := get("TextSpacing"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		textSpacing[0] = float32(fv.Index(0).Float())
		textSpacing[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("TextWrap"); ok && fv.Kind() == reflect.String {
		textWrap = fv.String()
	}
	if fv, ok := get("Velocity"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		velocity[0] = float32(fv.Index(0).Float())
		velocity[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("MaxDist"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		maxDist[0] = float32(fv.Index(0).Float())
		maxDist[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Accel"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		accel[0] = float32(fv.Index(0).Float())
		accel[1] = float32(fv.Index(1).Float())
	}
	if fv, ok := get("Friction"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		friction[0] = float32(fv.Index(0).Float())
		friction[1] = float32(fv.Index(1).Float())
	}

	objVal := reflect.ValueOf(obj).Elem()

	// textImgNew
	ts := NewTextSprite()
	// textImgSetFont
	fntField := objVal.FieldByName("Fnt")
	key := reflect.ValueOf(int(font[0]))
	v := fntField.MapIndex(key)
	if v.IsValid() {
		if fnt, ok := v.Interface().(*Fnt); ok {
			ts.fnt = fnt
		}
	}
	// textImgSetBank
	ts.bank = font[1]
	// textImgSetAlign
	ts.align = font[2]
	// textImgSetText
	ts.text = text
	ts.textInit = text
	// textImgSetColor
	ts.SetColor(font[3], font[4], font[5], font[6])
	// textImgSetLocalcoord
	ts.SetLocalcoord(localcoord[0], localcoord[1])
	// textImgSetPos
	ts.SetPos(offset[0], offset[1])
	// textImgSetScale
	ts.SetScale(scale[0], scale[1])
	// textImgSetXShear
	ts.xshear = xshear
	// textImgSetAngle
	ts.rot.angle = angle
	// textImgSetXAngle
	ts.rot.xangle = xangle
	// textImgSetYAngle
	ts.rot.yangle = yangle
	// textImgSetProjection
	ts.projection = projection
	// textImgSetFocalLength
	ts.fLength = focallength
	// textImgSetWindow
	ts.SetWindow(window)
	// textImgSetLayerno
	ts.layerno = layerno
	// textImgSetTextDelay
	ts.textDelay = textDelay
	// textImgSetTextSpacing
	ts.SetTextSpacing(textSpacing[0], textSpacing[1])
	// textImgSetTextWrap
	ts.textWrap = textWrap == "w" || textWrap == "1"
	// textImgSetVelocity
	ts.SetVelocity(velocity[0], velocity[1])
	// textImgSetMaxDist
	ts.SetMaxDist(maxDist[0], maxDist[1])
	// textImgSetAccel
	ts.SetAccel(accel[0], accel[1])
	// textImgSetFriction
	ts.friction = friction

	fVal.Set(reflect.ValueOf(ts))
}

// SetPalFx sets the PalFX field generically for any struct.
func SetPalFx(obj interface{}, fVal, structVal, parent reflect.Value) {
	time := int32(-1)
	color := float32(256)
	hue := float32(0)
	add := [3]int32{}
	mul := [3]int32{256, 256, 256}
	sinAdd := [4]int32{}
	sinMul := [4]int32{}
	sinColor := [2]int32{}
	sinHue := [2]int32{}
	invertAll := false
	invertBlend := int32(0)

	get := func(name string) (reflect.Value, bool) {
		return getFieldFromHierarchy(structVal, parent, name)
	}

	if fv, ok := get("Time"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		time = int32(fv.Int())
	}
	if fv, ok := get("Color"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		color = float32(fv.Float())
	}
	if fv, ok := get("Hue"); ok && (fv.Kind() == reflect.Float32 || fv.Kind() == reflect.Float64) {
		hue = float32(fv.Float())
	}
	if fv, ok := get("Add"); ok && fv.Kind() == reflect.Array && fv.Len() == 3 {
		for k := 0; k < 3; k++ {
			add[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("Mul"); ok && fv.Kind() == reflect.Array && fv.Len() == 3 {
		for k := 0; k < 3; k++ {
			mul[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("SinAdd"); ok && fv.Kind() == reflect.Array && fv.Len() == 4 {
		for k := 0; k < 4; k++ {
			sinAdd[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("SinMul"); ok && fv.Kind() == reflect.Array && fv.Len() == 4 {
		for k := 0; k < 4; k++ {
			sinMul[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("SinColor"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		for k := 0; k < 2; k++ {
			sinColor[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("SinHue"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		for k := 0; k < 2; k++ {
			sinHue[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("InvertAll"); ok && fv.Kind() == reflect.Bool {
		invertAll = fv.Bool()
	}
	if fv, ok := get("InvertBlend"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		invertBlend = int32(fv.Int())
	}

	palfx := newPalFX()
	palfx.time = time
	palfx.color = color / 256
	palfx.hue = hue / 256
	palfx.add = add
	palfx.mul = mul
	if sinAdd[3] < 0 {
		palfx.sinadd[0] = -sinAdd[0]
		palfx.sinadd[1] = -sinAdd[1]
		palfx.sinadd[2] = -sinAdd[2]
		palfx.cycletime[0] = -sinAdd[3]
	} else {
		palfx.sinadd[0] = sinAdd[0]
		palfx.sinadd[1] = sinAdd[1]
		palfx.sinadd[2] = sinAdd[2]
		palfx.cycletime[0] = sinAdd[3]
	}
	if sinMul[3] < 0 {
		palfx.sinmul[0] = -sinMul[0]
		palfx.sinmul[1] = -sinMul[1]
		palfx.sinmul[2] = -sinMul[2]
		palfx.cycletime[1] = -sinMul[3]
	} else {
		palfx.sinmul[0] = sinMul[0]
		palfx.sinmul[1] = sinMul[1]
		palfx.sinmul[2] = sinMul[2]
		palfx.cycletime[1] = sinMul[3]
	}
	if sinColor[1] < 0 {
		palfx.sincolor = -sinColor[0]
		palfx.cycletime[2] = -sinColor[1]
	} else {
		palfx.sincolor = sinColor[0]
		palfx.cycletime[2] = sinColor[1]
	}
	if sinHue[1] < 0 {
		palfx.sinhue = -sinHue[0]
		palfx.cycletime[3] = -sinHue[1]
	} else {
		palfx.sinhue = sinHue[0]
		palfx.cycletime[3] = sinHue[1]
	}
	palfx.invertall = invertAll
	palfx.invertblend = invertBlend

	fVal.Set(reflect.ValueOf(palfx))
}

// SetRect sets the Rect field generically for any struct.
func SetRect(obj interface{}, fVal, structVal, parent reflect.Value) {
	time := int32(0)
	layerno := int16(0)
	localcoord := [2]float32{0, 0}
	window := [4]float32{0, 0, 0, 0}
	col := [3]int32{0, 0, 0}
	alpha := [2]int32{0, 0}
	pulse := [3]int32{0, 0, 0}
	var usePulse bool
	//var attachPfx *PalFX

	get := func(name string) (reflect.Value, bool) {
		return getFieldFromHierarchy(structVal, parent, name)
	}

	if fv, ok := get("Time"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		time = int32(fv.Int())
	}
	// Layerno can be stored under different field names depending on context.
	for _, name := range []string{"Layerno", "ClearLayerno", "BgClearLayerno"} {
		if fv, ok := get(name); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
			layerno = int16(fv.Int())
		}
	}
	if fv, ok := get("Window"); ok && fv.Kind() == reflect.Array && fv.Len() == 4 {
		for k := 0; k < 4; k++ {
			window[k] = float32(fv.Index(k).Int())
		}
	}
	for _, name := range []string{"Col", "ClearColor", "BgClearColor"} {
		if fv, ok := get(name); ok && fv.Kind() == reflect.Array && fv.Len() == 3 {
			for k := 0; k < 3; k++ {
				col[k] = int32(fv.Index(k).Int())
			}
		}
	}
	for _, name := range []string{"Alpha", "ClearAlpha", "BgClearAlpha"} {
		if fv, ok := get(name); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
			for k := 0; k < 2; k++ {
				alpha[k] = int32(fv.Index(k).Int())
			}
		}
	}
	if fv, ok := get("Pulse"); ok && fv.Kind() == reflect.Array && fv.Len() == 3 {
		for k := 0; k < 3; k++ {
			pulse[k] = int32(fv.Index(k).Int())
		}
		usePulse = true
	}
	if fv, ok := get("Localcoord"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		localcoord[0] = float32(fv.Index(0).Int())
		localcoord[1] = float32(fv.Index(1).Int())
	}

	// rectNew
	rect := NewRect()
	// rectSetLocalcoord
	rect.SetLocalcoord(localcoord[0], localcoord[1])
	// rectSetTime
	rect.time = time
	// rectSetWindow
	rect.SetWindow(window)
	// rectSetColor
	rect.SetColor(col)
	// rectSetAlpha / rectSetAlphaPulse
	if usePulse {
		rect.SetAlphaPulse(pulse[0], pulse[1], pulse[2])
	} else {
		rect.SetAlpha(alpha)
	}
	// rectSetLayerno
	rect.layerno = layerno
	// rectSetPalFx
	//if attachPfx != nil {
	//	rect.SetPalFx(attachPfx)
	//}

	fVal.Set(reflect.ValueOf(rect))
}

// SetFade sets the Fade field generically for any struct.
func SetFade(obj interface{}, fVal, structVal, parent reflect.Value) {
	time := int32(0)
	col := [3]int32{0, 0, 0}
	animData := &Anim{}
	snd := [2]int32{-1, 0}

	get := func(name string) (reflect.Value, bool) {
		return getFieldFromHierarchy(structVal, parent, name)
	}

	if fv, ok := get("Time"); ok && fv.Kind() >= reflect.Int && fv.Kind() <= reflect.Int64 {
		time = int32(fv.Int())
	}
	if fv, ok := get("Col"); ok && fv.Kind() == reflect.Array && fv.Len() == 3 {
		for k := 0; k < 3; k++ {
			col[k] = int32(fv.Index(k).Int())
		}
	}
	if fv, ok := get("AnimData"); ok {
		switch fv.Kind() {
		case reflect.Ptr:
			if !fv.IsNil() {
				if a, ok := fv.Interface().(*Anim); ok {
					animData = a
				}
			}
		case reflect.Struct:
			if fv.CanAddr() {
				if a, ok := fv.Addr().Interface().(*Anim); ok {
					animData = a
				}
			} else {
				if a, ok := fv.Interface().(Anim); ok {
					animData = &a
				} else {
					tmp := fv.Interface().(Anim)
					animData = &tmp
				}
			}
		}
	}
	if fv, ok := get("Snd"); ok && fv.Kind() == reflect.Array && fv.Len() == 2 {
		for k := 0; k < 2; k++ {
			snd[k] = int32(fv.Index(k).Int())
		}
	}

	fade := newFade()
	fade.time = time
	fade.col = col
	fade.animData = animData
	fade.snd = snd

	fVal.Set(reflect.ValueOf(fade))
}

// PopulateDataPointers initializes all *Anim, *TextSprite, *PalFX, *Rect, *Fade pointers in the struct.
func PopulateDataPointers(obj interface{}, rootLocalcoord [2]int32) {
	// Resolve root SFF once (defaults for subtrees without an override).
	rootObjVal := reflect.ValueOf(obj).Elem()
	var rootSff *Sff
	if sf := rootObjVal.FieldByName("Sff"); sf.IsValid() && !sf.IsNil() {
		if s, ok := sf.Interface().(*Sff); ok {
			rootSff = s
		}
	}

	// Thread both localcoord and effective SFF through recursion.
	var populate func(v reflect.Value, parent reflect.Value, currentLocalCoord [2]int32, currentSff *Sff)

	// Types we look for
	animPtrType := reflect.TypeOf((*Anim)(nil))
	textSpritePtrType := reflect.TypeOf((*TextSprite)(nil))
	palFxPtrType := reflect.TypeOf((*PalFX)(nil))
	rectPtrType := reflect.TypeOf((*Rect)(nil))
	fadePtrType := reflect.TypeOf((*Fade)(nil))

	// The recursive function
	populate = func(v reflect.Value, parent reflect.Value, currentLocalCoord [2]int32, currentSff *Sff) {
		if !v.IsValid() {
			return
		}
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return
			}
			populate(v.Elem(), parent, currentLocalCoord, currentSff)
			return
		}

		switch v.Kind() {
		// If we have a struct, we iterate through its fields.
		case reflect.Struct:
			t := v.Type()
			// We make a local copy of the currentLocalCoord to allow an override
			localCoordForThisStruct := currentLocalCoord
			// Effective SFF for this struct (may be overridden by the field tag of children)
			effectiveSffForThisStruct := currentSff

			for i := 0; i < v.NumField(); i++ {
				fVal := v.Field(i)
				fType := t.Field(i)

				// Skip unexported fields
				if fType.PkgPath != "" {
					continue
				}

				if fType.Name == "Localcoord" && fVal.CanSet() &&
					fVal.Kind() == reflect.Array && fVal.Len() == 2 &&
					fVal.Index(0).CanSet() && fVal.Index(1).CanSet() {
					if fVal.Index(0).Int() == 0 && fVal.Index(1).Int() == 0 {
						fVal.Index(0).SetInt(int64(localCoordForThisStruct[0]))
						fVal.Index(1).SetInt(int64(localCoordForThisStruct[1]))
					}
				}

				// Check if this field declares an SFF override via struct tag: sff:"FieldName"
				nextSff := effectiveSffForThisStruct
				if tag := fType.Tag.Get("sff"); tag != "" {
					sf := rootObjVal.FieldByName(tag)
					if sf.IsValid() && !sf.IsNil() {
						if s, ok := sf.Interface().(*Sff); ok {
							nextSff = s
						}
					}
				}

				kind := fVal.Kind()

				// Recurse for struct or non-nil ptr
				if kind == reflect.Struct || (kind == reflect.Ptr && !fVal.IsNil()) {
					populate(fVal, v, localCoordForThisStruct, nextSff)
					continue
				}
				// If it's a map, recurse as well
				if kind == reflect.Map {
					// Pass along the possibly overridden SFF from this field's tag.
					populate(fVal, parent, localCoordForThisStruct, nextSff)
					continue
				}
				// If it's *Anim
				if kind == reflect.Ptr && fVal.Type().AssignableTo(animPtrType) {
					if fVal.IsNil() && fVal.CanSet() {
						SetAnim(obj, fVal, v, parent, effectiveSffForThisStruct)
					}
					continue
				}
				// If it's *TextSprite
				if kind == reflect.Ptr && fVal.Type().AssignableTo(textSpritePtrType) {
					if fVal.IsNil() && fVal.CanSet() {
						SetTextSprite(obj, fVal, v, parent)
					}
					continue
				}
				// If it's *PalFX
				if kind == reflect.Ptr && fVal.Type().AssignableTo(palFxPtrType) {
					if fVal.IsNil() && fVal.CanSet() {
						SetPalFx(obj, fVal, v, parent)
					}
					continue
				}
				// If it's *Rect
				if kind == reflect.Ptr && fVal.Type().AssignableTo(rectPtrType) {
					if fVal.IsNil() && fVal.CanSet() {
						SetRect(obj, fVal, v, parent)
					}
					continue
				}
				// If it's *Fade
				if kind == reflect.Ptr && fVal.Type().AssignableTo(fadePtrType) {
					if fVal.IsNil() && fVal.CanSet() {
						SetFade(obj, fVal, v, parent)
					}
					continue
				}
				// Otherwise it's a basic type
			}

		// If we have a slice or array, recurse into each element
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				populate(v.Index(i), parent, currentLocalCoord, currentSff)
			}
		// If we have a map, recurse into each key-value pair
		case reflect.Map:
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				populate(val, parent, currentLocalCoord, currentSff)
			}
		default:
			// basic types: do nothing
		}
	}

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		fmt.Println("Error: PopulateDataPointers requires a non-nil pointer")
		return
	}
	// Start recursion with the original user-supplied localcoord
	populate(v.Elem(), v.Elem(), rootLocalcoord, rootSff)
}

// Split a [Music] key into (prefix, property) while allowing dots in prefix.
func splitMusicKey(rawKey string) (prefix string, property string) {
	k := strings.TrimSpace(rawKey)
	if k == "" {
		return "", ""
	}
	kl := strings.ToLower(k)

	// Treat bare "bgm.*"/"bgmusic.*"/"music.*" keys as unprefixed so parseMusicSection sees "bgm.loop" etc.
	if strings.HasPrefix(kl, "bgm.") || strings.HasPrefix(kl, "bgmusic.") || strings.HasPrefix(kl, "music.") {
		return "", kl
	}

	// Find the last occurrence of a known music anchor so that prefixes can contain dots.
	anchors := []string{".bgmusic", ".music", ".bgm"}
	best := -1
	for _, a := range anchors {
		if i := strings.LastIndex(kl, a); i > best {
			best = i
		}
	}

	if best >= 0 {
		prefix = strings.TrimSpace(k[:best])
		property = strings.TrimSpace(kl[best+1:]) // without leading dot
		return prefix, property
	}

	// Fallback
	if dotIdx := strings.Index(k, "."); dotIdx >= 0 {
		return strings.TrimSpace(k[:dotIdx]), strings.TrimSpace(strings.ToLower(k[dotIdx+1:]))
	}
	return "", strings.ToLower(k)
}

func parseMusicSection(section *ini.Section) Music {
	// If section is nil, just return empty Music
	if section == nil {
		return make(Music)
	}

	// Local holder that preserves list alignment while allowing "unset" (nil)
	// so defaults from newBgMusic() are not overridden by empty values.
	type musicHolder struct {
		Bgm           []string
		Loop          []*int32
		Volume        []*int32
		LoopStart     []*int32
		LoopEnd       []*int32
		StartPosition []*int32
		FreqMul       []*float32
		LoopCount     []*int32
	}

	propMap := make(map[string]*musicHolder)

	get := func(prefix string) *musicHolder {
		if h, ok := propMap[prefix]; ok {
			return h
		}
		h := &musicHolder{}
		propMap[prefix] = h
		return h
	}

	appendI32PtrList := func(dst *[]*int32, items []string) {
		for _, s := range items {
			s = strings.TrimSpace(s)
			if s == "" {
				*dst = append(*dst, nil)
				continue
			}
			v := Atoi(s)
			*dst = append(*dst, &v)
		}
	}
	appendF32PtrList := func(dst *[]*float32, items []string) {
		for _, s := range items {
			s = strings.TrimSpace(s)
			if s == "" {
				*dst = append(*dst, nil)
				continue
			}
			v := float32(Atof(s))
			*dst = append(*dst, &v)
		}
	}
	maxSize := func(h *musicHolder) int {
		maxLen := len(h.Bgm)
		check := func(n int) {
			if n > maxLen {
				maxLen = n
			}
		}
		check(len(h.Loop))
		check(len(h.Volume))
		check(len(h.LoopStart))
		check(len(h.LoopEnd))
		check(len(h.StartPosition))
		check(len(h.FreqMul))
		check(len(h.LoopCount))
		return maxLen
	}

	// Parse keys
	for _, key := range section.Keys() {
		rawKey := strings.TrimSpace(key.Name())  // e.g. "title.bgm.loop" or "bgmusic"
		rawVal := strings.TrimSpace(key.Value()) // e.g. "1, 2, 3"
		if rawKey == "" {
			continue
		}

		// Split into prefix and property, allowing dots inside prefix.
		prefixRaw, property := splitMusicKey(rawKey)

		// Normalize prefix: lower-case + flatten dots to underscores
		prefix := strings.ToLower(prefixRaw)
		prefix = strings.ReplaceAll(prefix, ".", "_")

		// Split comma-separated values
		values := strings.Split(rawVal, ",")
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}

		// Fill BgmProperties fields based on recognized property names
		switch property {
		case "bgm", "bgmusic":
			h := get(prefix)
			h.Bgm = append(h.Bgm, values...)
		case "bgm.loop", "bgmloop":
			h := get(prefix)
			appendI32PtrList(&h.Loop, values)
		case "bgm.volume", "bgmvolume":
			h := get(prefix)
			appendI32PtrList(&h.Volume, values)
		case "bgm.loopstart", "bgmloopstart":
			h := get(prefix)
			appendI32PtrList(&h.LoopStart, values)
		case "bgm.loopend", "bgmloopend":
			h := get(prefix)
			appendI32PtrList(&h.LoopEnd, values)
		case "bgm.startposition", "bgmstartposition":
			h := get(prefix)
			appendI32PtrList(&h.StartPosition, values)
		case "bgm.freqmul", "bgmfreqmul":
			h := get(prefix)
			appendF32PtrList(&h.FreqMul, values)
		case "bgm.loopcount", "bgmloopcount":
			h := get(prefix)
			appendI32PtrList(&h.LoopCount, values)
		default:
			// unrecognized => skip
			continue
		}
	}

	// Convert propMap => Music
	music := make(Music)
	for prefix, holder := range propMap {
		// Skip prefixes that never had any real bgm-related values assigned.
		if len(holder.Bgm) == 0 &&
			len(holder.Loop) == 0 &&
			len(holder.Volume) == 0 &&
			len(holder.LoopStart) == 0 &&
			len(holder.LoopEnd) == 0 &&
			len(holder.StartPosition) == 0 &&
			len(holder.FreqMul) == 0 &&
			len(holder.LoopCount) == 0 {
			continue
		}

		// Build the final []*bgMusic slice
		count := maxSize(holder)
		for i := 0; i < count; i++ {
			bg := newBgMusic()
			if i < len(holder.Bgm) {
				bg.bgmusic = holder.Bgm[i]
			}
			if i < len(holder.Loop) && holder.Loop[i] != nil {
				bg.bgmloop = *holder.Loop[i]
			}
			if i < len(holder.Volume) && holder.Volume[i] != nil {
				bg.bgmvolume = *holder.Volume[i]
			}
			if i < len(holder.LoopStart) && holder.LoopStart[i] != nil {
				bg.bgmloopstart = *holder.LoopStart[i]
			}
			if i < len(holder.LoopEnd) && holder.LoopEnd[i] != nil {
				bg.bgmloopend = *holder.LoopEnd[i]
			}
			if i < len(holder.StartPosition) && holder.StartPosition[i] != nil {
				bg.bgmstartposition = *holder.StartPosition[i]
			}
			if i < len(holder.FreqMul) && holder.FreqMul[i] != nil {
				bg.bgmfreqmul = *holder.FreqMul[i]
			}
			if i < len(holder.LoopCount) && holder.LoopCount[i] != nil {
				bg.bgmloopcount = *holder.LoopCount[i]
			}
			music[prefix] = append(music[prefix], bg)
		}
	}

	// Debug dump
	music.DebugDump(fmt.Sprintf("parseMusicSection[%s]", section.Name()))

	// Return final music
	return music
}

// -------------------------------------------------------------------
// Inline font helpers
// -------------------------------------------------------------------

// fontKey builds a deduplication key "normalizedPath|height".
func fontKey(path string, height int32) string {
	p := filepath.ToSlash(path)
	return fmt.Sprintf("%s|%d", p, height)
}

// registerFontIndex records that (path,height) is served by font index idx.
func registerFontIndex(indexByKey map[string]int, path string, height int32, idx int) {
	if indexByKey == nil || path == "" {
		return
	}
	indexByKey[fontKey(path, height)] = idx
}

// ensureFontIndex resolves/loads a font and returns its index in the provided map.
// It deduplicates by (resolved filepath, height).
func ensureFontIndex(fnt map[int]*Fnt, indexByKey map[string]int, baseDef string, reqPath string, height int32) int {
	if fnt == nil {
		// Make sure the caller prepared the map (initMaps does it for structs).
		fnt = make(map[int]*Fnt)
	}
	// Search relative to the .def plus typical font dirs.
	resolved := SearchFile(strings.Trim(reqPath, `"`), []string{baseDef, "font/", "", "data/"})
	if resolved == "" {
		resolved = reqPath
	}
	if idx, ok := indexByKey[fontKey(resolved, height)]; ok {
		return idx
	}
	// Find the next free index in fnt map.
	idx := 0
	for {
		if _, taken := fnt[idx]; !taken {
			break
		}
		idx++
	}
	loaded, err := loadFnt(resolved, height)
	if err != nil {
		sys.errLog.Printf("Failed to load %v: %v", resolved, err)
	}
	if loaded == nil {
		loaded = newFnt()
	}
	fnt[idx] = loaded
	registerFontIndex(indexByKey, resolved, height, idx)
	return idx
}

// resolveInlineFonts scans all non-[Files]/non-[Music] sections for keys that end
// with ".font" (or equal "font"), and when the first element isn't an int, treats
// it as a filename, loads/attaches a font at the next free index, and replaces
// that first element with the index (updating both struct and INI via setValueUpdate).
// The 7th element (height) is passed into loadFnt as a height.
func resolveInlineFonts(iniFile *ini.File, baseDef string, fnt map[int]*Fnt, indexByKey map[string]int,
	setValueUpdate func(query string, value interface{}) error) {
	if iniFile == nil {
		return
	}
	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		// Skip defaults, [Files] (fontN.font there is for file declarations), and [Music]
		if secName == ini.DEFAULT_SECTION || strings.EqualFold(secName, "files") || strings.EqualFold(secName, "music") {
			continue
		}
		for _, k := range sec.Keys() {
			kName := k.Name()
			ln := strings.ToLower(kName)
			if !(ln == "font" || strings.HasSuffix(ln, ".font")) {
				continue
			}
			raw := strings.TrimSpace(k.Value())
			if raw == "" {
				continue
			}
			parts := strings.Split(raw, ",")
			if len(parts) == 0 {
				continue
			}
			first := strings.TrimSpace(parts[0])
			if IsInt(first) {
				// already an index
				continue
			}
			// treat as filename; derive height from last element when present
			h := int32(-1)
			if len(parts) > 7 {
				hvStr := strings.TrimSpace(parts[7])
				if IsInt(hvStr) {
					h = Atoi(hvStr)
				}
			}
			idx := ensureFontIndex(fnt, indexByKey, baseDef, first, h)
			parts[0] = Itoa(idx)
			newVal := strings.Join(parts, ",")
			// mirror original loader key normalization (spaces -> underscores)
			secPath := strings.ReplaceAll(secName, " ", "_")
			keyPath := strings.ReplaceAll(kName, " ", "_")
			_ = setValueUpdate(secPath+"."+keyPath, newVal)
		}
	}
}
