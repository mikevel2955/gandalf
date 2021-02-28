package utils

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

func ReadConfig(config interface{}) error {
	configValue := reflect.ValueOf(config).Elem()
	if !configValue.CanSet() {
		return errors.New("passed config can't be read")
	}

	configType := configValue.Type()
	numField := configType.NumField()
	for i := 0; i < numField; i++ {
		fieldValue := configValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		fieldType := configType.Field(i)
		fieldTag := fieldType.Tag

		envName, ok := fieldTag.Lookup("env")
		if !ok {
			continue
		}
		envValue, ok := os.LookupEnv(envName)
		if !ok {
			envValue, ok = fieldTag.Lookup("def")
			if !ok {
				continue
			}
		}

		switch fieldType.Type.String() {
		case "string":
			fieldValue.SetString(envValue)
		case "int", "int8", "int32", "int64":
			n, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("can't read int value for '%s': %v", fieldType.Name, err))
			}
			fieldValue.SetInt(n)
		case "uint", "uint8", "uint32", "uint64":
			u, err := strconv.ParseUint(envValue, 10, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("can't read uint value for '%s': %v", fieldType.Name, err))
			}
			fieldValue.SetUint(u)
		case "float", "float32", "float64":
			x, err := strconv.ParseFloat(envValue, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("can't read float value for '%s': %v", fieldType.Name, err))
			}
			fieldValue.SetFloat(x)
		case "time.Duration":
			d, err := time.ParseDuration(envValue)
			if err != nil {
				return errors.New(fmt.Sprintf("can't read time.Duration value for '%s': %v", fieldType.Name, err))
			}
			fieldValue.SetInt(d.Nanoseconds())
		case "bool":
			b, err := strconv.ParseBool(envValue)
			if err != nil {
				return errors.New(fmt.Sprintf("can't read bool value for '%s': %v", fieldType.Name, err))
			}
			fieldValue.SetBool(b)
		default:
			return errors.New(fmt.Sprintf("unsupported type for '%s'", fieldType.Name))
		}
	}
	return nil
}
