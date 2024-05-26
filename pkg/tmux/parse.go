package tmux

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// getFormat returns the "format" tags for the given object.
func getFormat(obj any) []string {
	v := reflect.ValueOf(obj)
	values := make([]string, 0)
	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("format")
		tag = fmt.Sprintf("#{%s}", tag)
		values = append(values, tag)
	}
	return values
}

// parseOutput parses the output of a tmux command into the given object.
func parseOutput(output string, obj any) error {
	v := reflect.ValueOf(obj).Elem()
	numFields := v.NumField()
	columns := strings.Split(output, ColumnSep)
	if len(columns) != numFields {
		return ErrInvalidFormat
	}

	for i, col := range columns {
		if i >= numFields {
			break
		}
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		fieldType := field.Type()
		switch fieldType.Kind() {
		case reflect.String:
			field.SetString(col)
		case reflect.Bool:
			field.SetBool(col == "1")
		case reflect.Int:
			num, err := strconv.ParseInt(col, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(num)

		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				date, err := parseUnixTime(col)
				if err != nil {
					return err
				}
				field.Set(reflect.ValueOf(date))
			}
		}
	}
	return nil
}

// parseUnixTime converts an epoch string to time.Time.
func parseUnixTime(epoch string) (time.Time, error) {
	createdEpoch, err := strconv.ParseInt(epoch, 10, 0)
	if err != nil {
		return time.Time{}, err
	}
	created := time.Unix(createdEpoch, 0)
	return created, nil
}
