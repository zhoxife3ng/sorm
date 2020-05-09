package util

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"fmt"
	"runtime/debug"
)

// refer https://github.com/didi/gendry/tree/master/scanner
// ByteUnmarshaler is the interface implemented by types
// that can unmarshal a JSON description of themselves.
// The input can be assumed to be a valid encoding of
// a JSON value. UnmarshalByte must copy the JSON data
// if it wishes to retain the data after returning.
//
// By convention, to approximate the behavior of Unmarshal itself,
// ByteUnmarshaler implement UnmarshalByte([]byte("null")) as a no-op.
type ByteUnmarshaler interface {
	UnmarshalByte(data []byte) error
}

//Rows defines methods that scanner needs, which database/sql.Rows already implements
const (
	cTimeFormat = "2006-01-02 15:04:05"
)

var (
	//ErrTargetNotSettable means the second param of Bind is not settable
	ErrTargetNotSettable = errors.New("[scanner]: target is not settable! a pointer is required")
	//ErrSliceToString means only []uint8 can be transmuted into string
	ErrSliceToString = errors.New("[scanner]: can't transmute a non-uint8 slice to string")
	//ErrEmptyResult occurs when target of Scan isn't slice and the result of the query is empty
	ErrEmptyResult = errors.New(`[scanner]: empty result`)
)

//ScanErr will be returned if an underlying type couldn't be AssignableTo type of target field
type ScanErr struct {
	structName, fieldName string
	from, to              reflect.Type
}

func (s ScanErr) Error() string {
	return fmt.Sprintf("[scanner]: %s.%s is %s which is not AssignableBy %s", s.structName, s.fieldName, s.to, s.from)
}

func newScanErr(structName, fieldName string, from, to reflect.Type) ScanErr {
	return ScanErr{structName, fieldName, from, to}
}

// refer https://github.com/didi/gendry/tree/master/scanner
func ScanMap(data []map[string]interface{}, target interface{}, tagName string) error {
	if nil == target || reflect.ValueOf(target).IsNil() || reflect.TypeOf(target).Kind() != reflect.Ptr {
		return ErrTargetNotSettable
	}
	if nil == data {
		return nil
	}
	return bindSlice(data, target, tagName)
}

// refer https://github.com/didi/gendry/tree/master/scanner
func ScanStruct(data map[string]interface{}, target interface{}, tagName string) error {
	if nil == target || reflect.ValueOf(target).IsNil() || reflect.TypeOf(target).Kind() != reflect.Ptr {
		return ErrTargetNotSettable
	}
	if nil == data {
		return ErrEmptyResult
	}
	return bind(data, target, tagName)
}

//caller must guarantee to pass a &slice as the second param
func bindSlice(arr []map[string]interface{}, target interface{}, tagName string) error {
	targetObj := reflect.ValueOf(target)
	if !targetObj.Elem().CanSet() {
		return ErrTargetNotSettable
	}
	length := len(arr)
	valueArrObj := reflect.MakeSlice(targetObj.Elem().Type(), 0, length)
	typeObj := valueArrObj.Type().Elem()
	var err error
	for i := 0; i < length; i++ {
		newObj := reflect.New(typeObj)
		newObjInterface := newObj.Interface()
		err = bind(arr[i], newObjInterface, tagName)
		if nil != err {
			return err
		}
		valueArrObj = reflect.Append(valueArrObj, newObj.Elem())
	}
	targetObj.Elem().Set(valueArrObj)
	return nil
}

func bind(result map[string]interface{}, target interface{}, tagName string) (resp error) {
	defer func() {
		if r := recover(); nil != r {
			resp = fmt.Errorf("error:[%v], stack:[%s]", r, string(debug.Stack()))
		}
	}()
	valueObj := reflect.ValueOf(target).Elem()
	if !valueObj.CanSet() {
		return ErrTargetNotSettable
	}
	typeObj := valueObj.Type()
	if typeObj.Kind() == reflect.Ptr {
		ptrType := typeObj.Elem()
		newObj := reflect.New(ptrType)
		newObjInterface := newObj.Interface()
		err := bind(result, newObjInterface, tagName)
		if nil == err {
			valueObj.Set(newObj)
		}
		return err
	}
	typeObjName := typeObj.Name()

	for i := 0; i < valueObj.NumField(); i++ {
		fieldTypeI := typeObj.Field(i)
		fieldName := fieldTypeI.Name

		//for convenience
		wrapErr := func(from, to reflect.Type) ScanErr {
			return newScanErr(typeObjName, fieldName, from, to)
		}

		valuei := valueObj.Field(i)
		if !valuei.CanSet() {
			continue
		}
		tagName, ok := lookUpTagName(fieldTypeI, tagName)
		if !ok || "" == tagName {
			continue
		}
		mapValue, ok := result[tagName]
		if !ok || mapValue == nil {
			continue
		}
		// if one field is a pointer type, we must allocate memory for it first
		// except for that the pointer type implements the interface ByteUnmarshaler
		if fieldTypeI.Type.Kind() == reflect.Ptr && !fieldTypeI.Type.Implements(_byteUnmarshalerType) {
			valuei.Set(reflect.New(fieldTypeI.Type.Elem()))
			valuei = valuei.Elem()
		}
		err := convert(mapValue, valuei, wrapErr)
		if nil != err {
			return err
		}
	}
	return nil
}

var _byteUnmarshalerType = reflect.TypeOf(new(ByteUnmarshaler)).Elem()

type convertErrWrapper func(from, to reflect.Type) ScanErr

func isIntSeriesType(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

func isUintSeriesType(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uint64
}

func isFloatSeriesType(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}

func lookUpTagName(typeObj reflect.StructField, tagName string) (string, bool) {
	name, ok := typeObj.Tag.Lookup(tagName)
	if !ok {
		return "", false
	}
	name = resolveTagName(name)
	return name, ok
}

func convert(mapValue interface{}, valuei reflect.Value, wrapErr convertErrWrapper) error {
	//vit: ValueI Type
	vit := valuei.Type()
	//mvt: MapValue Type
	mvt := reflect.TypeOf(mapValue)
	if nil == mvt {
		return nil
	}
	//[]byte tp []byte && time.Time to time.Time
	if mvt.AssignableTo(vit) {
		valuei.Set(reflect.ValueOf(mapValue))
		return nil
	}
	//time.Time to string
	switch assertT := mapValue.(type) {
	case time.Time:
		return handleConvertTime(assertT, mvt, vit, &valuei, wrapErr)
	}

	if scanner, ok := valuei.Addr().Interface().(sql.Scanner); ok {
		return scanner.Scan(mapValue)
	}

	//according to go-mysql-driver/mysql, driver.Value type can only be:
	//int64 or []byte(> maxInt64)
	//float32/float64
	//[]byte
	//time.Time if parseTime=true or DATE type will be converted into []byte
	switch mvt.Kind() {
	case reflect.Int64:
		if isIntSeriesType(vit.Kind()) {
			valuei.SetInt(mapValue.(int64))
		} else if isUintSeriesType(vit.Kind()) {
			valuei.SetUint(uint64(mapValue.(int64)))
		} else if vit.Kind() == reflect.Bool {
			v := mapValue.(int64)
			if v > 0 {
				valuei.SetBool(true)
			} else {
				valuei.SetBool(false)
			}
		} else if vit.Kind() == reflect.String {
			valuei.SetString(strconv.FormatInt(mapValue.(int64), 10))
		} else {
			return wrapErr(mvt, vit)
		}
	case reflect.Float32:
		if isFloatSeriesType(vit.Kind()) {
			valuei.SetFloat(float64(mapValue.(float32)))
		} else {
			return wrapErr(mvt, vit)
		}
	case reflect.Float64:
		if isFloatSeriesType(vit.Kind()) {
			valuei.SetFloat(mapValue.(float64))
		} else {
			return wrapErr(mvt, vit)
		}
	case reflect.Slice:
		return handleConvertSlice(mapValue, mvt, vit, &valuei, wrapErr)
	default:
		return wrapErr(mvt, vit)
	}
	return nil
}

func handleConvertSlice(mapValue interface{}, mvt, vit reflect.Type, valuei *reflect.Value, wrapErr convertErrWrapper) error {
	mapValueSlice, ok := mapValue.([]byte)
	if !ok {
		return ErrSliceToString
	}
	mapValueStr := string(mapValueSlice)
	vitKind := vit.Kind()
	switch {
	case vitKind == reflect.String:
		valuei.SetString(mapValueStr)
	case isIntSeriesType(vitKind):
		intVal, err := strconv.ParseInt(mapValueStr, 10, 64)
		if nil != err {
			return wrapErr(mvt, vit)
		}
		valuei.SetInt(intVal)
	case isUintSeriesType(vitKind):
		uintVal, err := strconv.ParseUint(mapValueStr, 10, 64)
		if nil != err {
			return wrapErr(mvt, vit)
		}
		valuei.SetUint(uintVal)
	case isFloatSeriesType(vitKind):
		floatVal, err := strconv.ParseFloat(mapValueStr, 64)
		if nil != err {
			return wrapErr(mvt, vit)
		}
		valuei.SetFloat(floatVal)
	case vitKind == reflect.Bool:
		intVal, err := strconv.ParseInt(mapValueStr, 10, 64)
		if nil != err {
			return wrapErr(mvt, vit)
		}
		if intVal > 0 {
			valuei.SetBool(true)
		} else {
			valuei.SetBool(false)
		}
	default:
		if _, ok := valuei.Interface().(ByteUnmarshaler); ok {
			return byteUnmarshal(mapValueSlice, valuei, wrapErr)
		}
		return wrapErr(mvt, vit)
	}
	return nil
}

// valuei Here is the type of ByteUnmarshaler
func byteUnmarshal(mapValueSlice []byte, valuei *reflect.Value, wrapErr convertErrWrapper) error {
	var pt reflect.Value
	initFlag := false
	// init pointer
	if valuei.IsNil() {
		pt = reflect.New(valuei.Type().Elem())
		initFlag = true
	} else {
		pt = *valuei
	}
	err := pt.Interface().(ByteUnmarshaler).UnmarshalByte(mapValueSlice)
	if nil != err {
		structName := pt.Elem().Type().Name()
		return fmt.Errorf("[scanner]: %s.UnmarshalByte fail to unmarshal the bytes, err: %s", structName, err)
	}
	if initFlag {
		valuei.Set(pt)
	}
	return nil
}

func handleConvertTime(assertT time.Time, mvt, vit reflect.Type, valuei *reflect.Value, wrapErr convertErrWrapper) error {
	if vit.Kind() == reflect.String {
		sTime := assertT.Format(cTimeFormat)
		valuei.SetString(sTime)
		return nil
	}
	return wrapErr(mvt, vit)
}

func resolveTagName(tag string) string {
	idx := strings.IndexByte(tag, ',')
	if -1 == idx {
		return tag
	}
	return tag[:idx]
}
