package scan

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func ScanRow(v interface{}, r RowScanner, f string) error {
	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		return fmt.Errorf("%q must be a pointer", k.String())
	}

	vType = vType.Elem()
	vValue := reflect.ValueOf(v).Elem()

	if vType.Kind() != reflect.Struct {
		return fmt.Errorf("cannot scan Row struct")
	}

	fslice := strings.Split(f, ",")
	if len(fslice) == 0 {
		return nil
	}

	cols := make([]string, 0)
	for _, c := range fslice {
		cols = append(cols, strings.TrimSpace(c))
	}

	pointers := structPointers(vValue, cols)
	if len(pointers) == 0 {
		return nil
	}

	err := r.Scan(pointers...)
	if err != nil {
		return err
	}

	return r.Err()
}

func ScanRows(v interface{}, r RowsScanner, f string) error {
	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		return fmt.Errorf("%q must be a pointer", k.String())
	}

	sliceType := vType.Elem()
	if reflect.Slice != sliceType.Kind() {
		return fmt.Errorf("%q must be a slice", sliceType.String())
	}

	fslice := strings.Split(f, ",")
	if len(fslice) == 0 {
		return nil
	}

	cols := make([]string, 0)
	for _, c := range fslice {
		cols = append(cols, strings.TrimSpace(c))
	}

	sliceVal := reflect.Indirect(reflect.ValueOf(v))
	itemType := sliceType.Elem()

	for r.Next() {
		sliceItem := reflect.New(itemType).Elem()
		pointers := structPointers(sliceItem, cols)

		if len(pointers) == 0 {
			return nil
		}

		err := r.Scan(pointers...)
		if err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, sliceItem))
	}

	return r.Err()
}

func fieldByName(v reflect.Value, name string) (reflect.Value, error) {
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		tag, ok := typ.Field(i).Tag.Lookup("db")
		if ok && tag == name {
			return v.Field(i), nil
		}
	}

	return reflect.Value{}, errors.New("cannot find field")
}

func structPointers(stct reflect.Value, cols []string) []interface{} {
	pointers := make([]interface{}, 0, len(cols))
	for _, colName := range cols {
		fieldVal, err := fieldByName(stct, colName)
		if err != nil || !fieldVal.IsValid() || !fieldVal.CanSet() {
			var nothing interface{}
			pointers = append(pointers, &nothing)
			continue
		}
		pointers = append(pointers, fieldVal.Addr().Interface())
	}

	return pointers
}
