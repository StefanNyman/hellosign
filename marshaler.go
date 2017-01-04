// Copyright 2016 Precisely AB.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
)

func fieldTagName(tag string) string {
	sArr := strings.Split(tag, ",")
	if len(sArr) > 0 {
		return sArr[0]
	}
	return ""
}

func omitEmpty(tag string) bool {
	sArr := strings.Split(tag, ",")
	if len(sArr) == 2 {
		return sArr[1] == "omitempty"
	}
	return false
}

func (c *hellosign) marshalMultipart(obj interface{}) (*bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if err := marshalObj(w, "", obj); err != nil {
		return nil, nil, err
	}
	return &b, w, nil
}

func marshalObj(w *multipart.Writer, prefix string, obj interface{}) error {
	structType := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("cannot marshal nil ptr")
		}
		val = val.Elem()
		structType = reflect.TypeOf(val.Interface())
	}

	for i := 0; i < val.NumField(); i++ {
		valField := val.Field(i)
		f := valField.Interface()
		val := reflect.ValueOf(f)

		field := structType.Field(i)
		tag := field.Tag.Get("form")
		tagName := fieldTagName(tag)
		oe := omitEmpty(tag)

		if tagName == "" || tagName == "-" {
			continue
		}

		switch val.Kind() {
		case reflect.Ptr:
			if val.IsNil() {
				continue
			}
			vInter := val.Interface()
			if err := marshalObj(w, fmt.Sprintf("%s%s", prefix, tagName), vInter); err != nil {
				return err
			}
		case reflect.Map:
			if oe && len(val.MapKeys()) == 0 {
				continue
			}
			for _, k := range val.MapKeys() {
				if err := writeString(w, fmt.Sprintf("%s[%s]", prefix+tagName, k.String()), val.MapIndex(k).String()); err != nil {
					return err
				}
			}
		case reflect.Slice:
			if oe && val.Len() == 0 {
				continue
			}
			fIndexVal := val.Index(0)
			switch fIndexVal.Kind() {
			case reflect.Slice:
				if tagName == "file" {
					for i := 0; i < val.Len(); i++ {
						key := fmt.Sprintf("%s%s[%d]", prefix, tagName, i)
						inter := val.Index(i).Interface()
						bArr, ok := inter.([]byte)
						if !ok {
							return fmt.Errorf("%s is not a byte slice", key)
						}
						ff, err := w.CreateFormFile(key, fmt.Sprintf("Document %d", i))
						if err != nil {
							return err
						}
						ff.Write(bArr)
					}
				}
				// No else case as we don't really have any other kinds of slices in slices.
			case reflect.String:
				for i := 0; i < val.Len(); i++ {
					key := fmt.Sprintf("%s%s[%d]", prefix, tagName, i)
					if err := writeString(w, key, val.Index(i).String()); err != nil {
						return err
					}
				}
			case reflect.Struct:
				for i := 0; i < val.Len(); i++ {
					sObj := val.Index(i).Interface()
					key := fmt.Sprintf("%s%s[%d]", prefix, tagName, i)
					if err := marshalObj(w, key, sObj); err != nil {
						return err
					}
				}

			}

		default:
			key := fmt.Sprintf("%s%s", prefix, tagName)
			if prefix != "" {
				key = fmt.Sprintf("%s[%s]", prefix, tagName)
			}
			if err := marshalPrimitive(w, oe, key, val.Interface()); err != nil {
				return err
			}
		}

	}
	return nil
}

func marshalPrimitive(w *multipart.Writer, oe bool, tagName string, v interface{}) error {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		if err := marshalBool(w, tagName, oe, val.Bool()); err != nil {
			return err
		}
	case reflect.String:
		if oe && val.String() == "" {
			return nil
		}
		if err := writeString(w, tagName, val.String()); err != nil {
			return err
		}
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		if err := marshalInt(w, tagName, oe, val.Int()); err != nil {
			return err
		}
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uint:
		if err := marshalUint(w, tagName, oe, val.Uint()); err != nil {
			return err
		}
	default:
		if oe && val.String() == "" {
			return nil
		}
		if err := writeString(w, tagName, val.String()); err != nil {
			return err
		}
	}
	return nil
}

func writeString(w *multipart.Writer, name, val string) error {
	ff, err := w.CreateFormField(name)
	if err != nil {
		return nil
	}
	_, err = ff.Write([]byte(val))
	return err
}

func marshalUint(w *multipart.Writer, tagName string, oe bool, val uint64) error {
	if oe && val == 0 {
		return nil
	}
	strUint := strconv.FormatUint(val, 10)
	return writeString(w, tagName, strUint)
}

func marshalInt(w *multipart.Writer, tagName string, oe bool, val int64) error {
	if oe && val == 0 {
		return nil
	}
	strInt := strconv.FormatInt(val, 10)
	return writeString(w, tagName, strInt)
}

func marshalBool(w *multipart.Writer, tagName string, oe bool, val bool) error {
	if oe && val == false {
		return nil
	}
	strBool := strconv.FormatInt(int64(BoolToInt(val)), 10)
	return writeString(w, tagName, strBool)
}
