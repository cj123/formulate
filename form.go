package formulate

import (
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
)

type StructField struct {
	reflect.StructField
}

func (sf StructField) GetName() string {
	tagName := sf.Tag.Get("name")

	if tagName != "" {
		return tagName
	}

	return camelCase(sf.Name)
}

func camelCase(s string) string {
	return strings.Join(camelcase.Split(s), " ")
}

func (sf StructField) Type(original string) string {
	t := sf.Tag.Get("type")

	if t != "" {
		return t
	}

	return original
}

func (sf StructField) Elem() string {
	return sf.Tag.Get("elem")
}

func (sf StructField) HasMin() bool {
	return sf.Tag.Get("min") != ""
}

func (sf StructField) Min() string {
	return sf.Tag.Get("min")
}

func (sf StructField) HasMax() bool {
	return sf.Tag.Get("max") != ""
}

func (sf StructField) Max() string {
	return sf.Tag.Get("max")
}

func (sf StructField) HasStep() bool {
	return sf.Tag.Get("step") != ""
}

func (sf StructField) Step() string {
	return sf.Tag.Get("step")
}
