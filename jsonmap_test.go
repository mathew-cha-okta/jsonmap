package jsonmap

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

type brokenValidator struct{}

func (v brokenValidator) Validate(interface{}) (interface{}, error) {
	return nil, errors.New("this should be a ValidationError")
}

type InnerThing struct {
	Foo   string
	AnInt int
	ABool bool
}

type AnotherInnerThing struct {
	Foo        string
	AnInt      int
	ABool      bool
	HappenedAt time.Time
	ThanksGo   interface{}
}

type AnotherOuterThing struct {
	InnerThing AnotherInnerThing
}

type OuterThing struct {
	InnerThing InnerThing
}

type OuterInnerThingMap struct {
	InnerThingMap map[string]InnerThing
}

type OuterPointerThing struct {
	InnerThing *InnerThing
}

type OuterInterfaceThing struct {
	InnerThing interface{}
}

type OuterSliceThing struct {
	InnerThings []InnerThing
}

type Outer2DSliceThing struct {
	InnerThings [][]InnerThing
}

type OuterMaxSliceThing struct {
	InnerThings []InnerThing
}

type OuterMinSliceThing struct {
	InnerThings []InnerThing
}

type OuterRangeSliceThing struct {
	InnerThings []InnerThing
}

type OuterPointerSliceThing struct {
	InnerThings []*InnerThing
}

type OuterPointerToSliceThing struct {
	InnerThings *[]InnerThing
}

type OtherInnerThing struct {
	Bar string
}

type OuterVariableThing struct {
	InnerType  string
	InnerValue interface{}
}

type OuterVariableThingInnerTypeOneOf struct {
	InnerType  string      `json:"inner_type,omitempty"`
	InnerValue interface{} `json:"inner_thing"`
}

type OuterVariableThingInnerTypeNoJsonTag struct {
	InnerType  string
	InnerValue interface{} `json:"inner_thing"`
}

type OuterVariableThingInnerTypeIgnoredJsonTag struct {
	InnerType  string      `json:"-"`
	InnerValue interface{} `json:"inner_thing"`
}

type OtherOuterVariableThing OuterVariableThing

type ReadOnlyThing struct {
	PrimaryKey string
}

type UnregisteredThing struct {
}

type TypoedThing struct {
	Correct bool
}

type BrokenThing struct {
	Invalid string
}

type TemplatableThing struct {
	SomeField string
}

type NonMarshalableType struct{}

func (t NonMarshalableType) MarshalJSON() ([]byte, error) {
	return nil, errors.New("oops")
}

type InnerNonMarshalableThing struct {
	Oops NonMarshalableType
}

type OuterNonMarshalableThing struct {
	InnerThing InnerNonMarshalableThing
}

type ThingWithSliceOfPrimitives struct {
	Strings []string
}

type ThingWithMapOfInterfaces struct {
	Interfaces map[string]interface{}
}

type ThingWithMapOfStrings struct {
	Strings map[string]string
}

type OuterMapThing struct {
	InnerMap map[string]interface{}
}

type ThingWithTime struct {
	HappenedAt time.Time
}

type ThingWithEnumerableInterface struct {
	ThanksGo interface{}
}

var InnerThingTypeMap = StructMap{
	InnerThing{},
	[]MappedField{
		{
			StructFieldName: "Foo",
			JSONFieldName:   "foo",
			Validator:       String(1, 12),
			Optional:        true,
		},
		{
			StructFieldName: "AnInt",
			JSONFieldName:   "an_int",
			Validator:       Integer(0, 10),
			Optional:        true,
		},
		{
			StructFieldName: "ABool",
			JSONFieldName:   "a_bool",
			Validator:       Boolean(),
			Optional:        true,
		},
	},
}

var AnotherInnerThingTypeMap = StructMap{
	AnotherInnerThing{},
	[]MappedField{
		{
			StructFieldName: "Foo",
			JSONFieldName:   "foo",
			Validator:       String(1, 5),
			Optional:        true,
		},
		{
			StructFieldName: "AnInt",
			JSONFieldName:   "an~int",
			Validator:       Integer(0, 10),
			Optional:        true,
		},
		{
			StructFieldName: "ABool",
			JSONFieldName:   "a_bool",
			Validator:       Boolean(),
			Optional:        true,
		},
		{
			StructFieldName: "HappenedAt",
			JSONFieldName:   "happened_at",
			Contains:        Time(),
			Optional:        true,
		},
		{
			StructFieldName: "ThanksGo",
			JSONFieldName:   "thanks",
			Validator:       OneOf("foo", "bar"),
			Optional:        true,
		},
	},
}

var OuterThingTypeMap = StructMap{
	OuterThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThing",
			JSONFieldName:   "inner_thing",
			Contains:        InnerThingTypeMap,
		},
	},
}

var AnotherOuterThingTypeMap = StructMap{
	AnotherOuterThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThing",
			JSONFieldName:   "another/inner/thing",
			Contains:        AnotherInnerThingTypeMap,
		},
	},
}

var MapOfInnerThingTypeMap = StructMap{
	OuterInnerThingMap{},
	[]MappedField{
		{
			StructFieldName: "InnerThingMap",
			JSONFieldName:   "inner_thing_map",
			Contains:        MapOf(InnerThingTypeMap),
		},
	},
}

var OuterPointerThingTypeMap = StructMap{
	OuterPointerThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThing",
			JSONFieldName:   "inner_thing",
			Contains:        InnerThingTypeMap,
		},
	},
}

var OuterInterfaceThingTypeMap = StructMap{
	OuterInterfaceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThing",
			JSONFieldName:   "inner_thing",
			Contains:        InnerThingTypeMap,
		},
	},
}

var OuterSliceThingTypeMap = StructMap{
	OuterSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOf(InnerThingTypeMap),
		},
	},
}

var Outer2DSliceThingTypeMap = StructMap{
	Outer2DSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOf(SliceOf(InnerThingTypeMap)),
		},
	},
}

var ContainsMaxSliceSizeTypeMap = StructMap{
	OuterMaxSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOfMax(InnerThingTypeMap, 2),
		},
	},
}

var ContainsMinSliceSizeTypeMap = StructMap{
	OuterMinSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOfMin(InnerThingTypeMap, 2),
		},
	},
}

var ContainsRangeSliceSizeTypeMap = StructMap{
	OuterRangeSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOfRange(InnerThingTypeMap, 1, 2),
		},
	},
}

var OuterPointerSliceThingTypeMap = StructMap{
	OuterPointerSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOf(InnerThingTypeMap),
		},
	},
}

var OuterPointerToSliceThingTypeMap = StructMap{
	OuterPointerToSliceThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThings",
			JSONFieldName:   "inner_things",
			Contains:        SliceOf(InnerThingTypeMap),
		},
	},
}

var OtherInnerThingTypeMap = StructMap{
	OtherInnerThing{},
	[]MappedField{
		{
			StructFieldName: "Bar",
			JSONFieldName:   "bar",
			Validator:       String(1, 155),
			Optional:        true,
		},
	},
}

var OuterVariableThingTypeMap = StructMap{
	OuterVariableThing{},
	[]MappedField{
		{
			StructFieldName: "InnerType",
			JSONFieldName:   "inner_type",
			Validator:       String(1, 255),
		},
		{
			StructFieldName: "InnerValue",
			JSONFieldName:   "inner_thing",
			Contains: VariableType("InnerType", map[string]TypeMap{
				"foo": InnerThingTypeMap,
				"bar": OtherInnerThingTypeMap,
			}),
		},
	},
}

var OuterVariableThingWithOneOfInnerTypeMap = StructMap{
	OuterVariableThingInnerTypeOneOf{},
	[]MappedField{
		{
			StructFieldName: "InnerType",
			JSONFieldName:   "inner_type",
			Validator:       OneOf("these", "are", "allowed"),
		},
		{
			StructFieldName: "InnerValue",
			JSONFieldName:   "inner_thing",
			Contains: VariableType("InnerType", map[string]TypeMap{
				"foo":     InnerThingTypeMap,
				"bar":     OtherInnerThingTypeMap,
				"these":   NewPrimitiveMap(Integer(-5, 10)),
				"are":     NewPrimitiveMap(String(1, 5)),
				"allowed": InnerThingTypeMap,
			}),
		},
	},
}

var OuterVariableThingWithInnerTypeNoJsonTagTypeMap = StructMap{
	OuterVariableThingInnerTypeNoJsonTag{},
	[]MappedField{
		{
			StructFieldName: "InnerType",
			JSONFieldName:   "inner_type",
			Validator:       OneOf("these", "are", "allowed"),
		},
		{
			StructFieldName: "InnerValue",
			JSONFieldName:   "inner_thing",
			Contains: VariableType("InnerType", map[string]TypeMap{
				"foo": InnerThingTypeMap,
				"bar": OtherInnerThingTypeMap,
			}),
		},
	},
}

var OuterVariableThingWithInnerTypeIgnoredJsonTagTypeMap = StructMap{
	OuterVariableThingInnerTypeIgnoredJsonTag{},
	[]MappedField{
		{
			StructFieldName: "InnerType",
			JSONFieldName:   "inner_type",
			Validator:       OneOf("these", "are", "allowed"),
		},
		{
			StructFieldName: "InnerValue",
			JSONFieldName:   "inner_thing",
			Contains: VariableType("InnerType", map[string]TypeMap{
				"foo": InnerThingTypeMap,
				"bar": OtherInnerThingTypeMap,
			}),
		},
	},
}

var BrokenOuterVariableThingTypeMap = StructMap{
	OtherOuterVariableThing{},
	[]MappedField{
		{
			StructFieldName: "InnerType",
			JSONFieldName:   "inner_type",
			Validator:       String(1, 255),
		},
		{
			StructFieldName: "InnerValue",
			JSONFieldName:   "inner_thing",
			Contains: VariableType("InnerTypeo", map[string]TypeMap{
				"foo": InnerThingTypeMap,
				"bar": OtherInnerThingTypeMap,
			}),
		},
	},
}

var ReadOnlyThingTypeMap = StructMap{
	ReadOnlyThing{},
	[]MappedField{
		{
			StructFieldName: "PrimaryKey",
			JSONFieldName:   "primary_key",
			ReadOnly:        true,
		},
	},
}

var TypoedThingTypeMap = StructMap{
	TypoedThing{},
	[]MappedField{
		{
			StructFieldName: "Incorrect",
			JSONFieldName:   "correct",
			Validator:       Boolean(),
		},
	},
}

var BrokenThingTypeMap = StructMap{
	BrokenThing{},
	[]MappedField{
		{
			StructFieldName: "Invalid",
			JSONFieldName:   "invalid",
			Validator:       brokenValidator{},
		},
	},
}

var TemplatableThingTypeMap = StructMap{
	TemplatableThing{},
	[]MappedField{
		{
			StructFieldName: "SomeField",
			JSONFieldName:   "some_field",
			Contains:        StringRenderer("{{.Context.Foo}}:{{.Value}}"),
		},
	},
}

var InnerNonMarshalableThingTypeMap = StructMap{
	InnerNonMarshalableThing{},
	[]MappedField{
		{
			StructFieldName: "Oops",
			JSONFieldName:   "oops",
		},
	},
}

var OuterNonMarshalableThingTypeMap = StructMap{
	OuterNonMarshalableThing{},
	[]MappedField{
		{
			StructFieldName: "InnerThing",
			JSONFieldName:   "inner_thing",
			Contains:        InnerNonMarshalableThingTypeMap,
		},
	},
}

var ThingWithSliceOfPrimitivesTypeMap = StructMap{
	ThingWithSliceOfPrimitives{},
	[]MappedField{
		{
			StructFieldName: "Strings",
			JSONFieldName:   "strings",
			Contains:        SliceOf(NewPrimitiveMap(String(1, 16))),
		},
	},
}
var ThingWithInnerMapTypeMap = StructMap{
	OuterMapThing{},
	[]MappedField{
		{
			StructFieldName: "InnerMap",
			JSONFieldName:   "inner_map",
			Contains:        MapOf(NewPrimitiveMap(Interface())),
		},
	},
}

var ThingWithMapOfInterfacesTypeMap = StructMap{
	ThingWithMapOfInterfaces{},
	[]MappedField{
		{
			StructFieldName: "Interfaces",
			JSONFieldName:   "interfaces",
			Contains:        MapOf(NewPrimitiveMap(Interface())),
		},
	},
}

var ThingWithMapOfStringsTypeMap = StructMap{
	ThingWithMapOfStrings{},
	[]MappedField{
		{
			StructFieldName: "Strings",
			JSONFieldName:   "strings",
			Contains:        MapOf(NewPrimitiveMap(String(0, 5))),
		},
	},
}

var ThingWithTimeSchema = StructMap{
	ThingWithTime{},
	[]MappedField{
		{
			StructFieldName: "HappenedAt",
			JSONFieldName:   "happened_at",
			Contains:        Time(),
		},
	},
}

var ThingWithEnumerableInterfaceSchema = StructMap{
	ThingWithEnumerableInterface{},
	[]MappedField{
		{
			StructFieldName: "ThanksGo",
			JSONFieldName:   "thanks",
			Validator:       OneOf("foo", "bar"),
		},
	},
}

var TestTypeMapper = NewTypeMapper(
	InnerThingTypeMap,
	AnotherInnerThingTypeMap,
	OuterThingTypeMap,
	AnotherOuterThingTypeMap,
	OuterPointerThingTypeMap,
	OuterInterfaceThingTypeMap,
	OuterSliceThingTypeMap,
	ContainsMaxSliceSizeTypeMap,
	ContainsMinSliceSizeTypeMap,
	ContainsRangeSliceSizeTypeMap,
	OuterPointerSliceThingTypeMap,
	OuterPointerToSliceThingTypeMap,
	OuterVariableThingTypeMap,
	OuterVariableThingWithOneOfInnerTypeMap,
	OuterVariableThingWithInnerTypeNoJsonTagTypeMap,
	OuterVariableThingWithInnerTypeIgnoredJsonTagTypeMap,
	BrokenOuterVariableThingTypeMap,
	ReadOnlyThingTypeMap,
	TypoedThingTypeMap,
	BrokenThingTypeMap,
	TemplatableThingTypeMap,
	InnerNonMarshalableThingTypeMap,
	OuterNonMarshalableThingTypeMap,
	ThingWithSliceOfPrimitivesTypeMap,
	ThingWithInnerMapTypeMap,
	ThingWithMapOfInterfacesTypeMap,
	ThingWithMapOfStringsTypeMap,
	ThingWithTimeSchema,
	ThingWithEnumerableInterfaceSchema,
	MapOfInnerThingTypeMap,
	Outer2DSliceThingTypeMap,
)

func TestValidateInnerThing(t *testing.T) {
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"foo": "fooz", "an_int": 10, "a_bool": true}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if v.Foo != "fooz" {
		t.Fatal("Field Foo does not have expected value 'fooz':", v.Foo)
	}
}

func TestValidateAnotherInnerThing(t *testing.T) {
	expected := `Validation Errors: 
/foo: too long, may not be more than 5 characters
/an~0int: too large, may not be larger than 10
/happened_at: not a valid RFC 3339 time value
/thanks: Value must be one of: ["foo","bar"]
`
	v := &AnotherInnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"foo": "foozzzy", "an~int": 11, "happened_at": "hi", "thanks": "baz"}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterThing(t *testing.T) {
	v := &OuterThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_thing": {"foo": "fooz"}}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if v.InnerThing.Foo != "fooz" {
		t.Fatal("Inner field Foo does not have expected value 'fooz':", v.InnerThing.Foo)
	}
}

func TestValidateAnotherOuterThing(t *testing.T) {
	expected := `Validation Errors: 
/another~1inner~1thing/foo: too long, may not be more than 5 characters
/another~1inner~1thing/an~0int: too large, may not be larger than 10
/another~1inner~1thing/happened_at: not a valid RFC 3339 time value
/another~1inner~1thing/thanks: Value must be one of: ["foo","bar"]
`

	v := &AnotherOuterThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"another/inner/thing": {"foo": "foozzzy", "an~int": 11, "happened_at": "hi", "thanks": "baz"}}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterSliceThing(t *testing.T) {
	v := &OuterSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}]}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.InnerThings) != 1 {
		t.Fatal("InnerThings should contain 1 element, instead contains", len(v.InnerThings))
	}
	if v.InnerThings[0].Foo != "fooz" {
		t.Fatal("InnerThing field Foo does not have expected value 'fooz':", v.InnerThings[0].Foo)
	}
}

func TestValidateOuterSliceThingInvalidElement(t *testing.T) {
	expected := `Validation Errors: 
/inner_things/0/foo: too long, may not be more than 12 characters
`
	v := &OuterSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooziswaytoolooong"}]}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterSliceThingMultipleInvalidElements(t *testing.T) {
	expected := `Validation Errors: 
/inner_things/0/foo: too long, may not be more than 12 characters
/inner_things/1/foo: too long, may not be more than 12 characters
`
	v := &OuterSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooziswaytoolooong"}, {"foo": "fooziswaytoolooong2"}]}`), v)
	require.EqualError(t, err, expected)
}
func TestValidateOuter2DSliceThing(t *testing.T) {
	expected := `Validation Errors: 
/inner_things/0/1/foo: too long, may not be more than 12 characters
/inner_things/1/0/foo: too long, may not be more than 12 characters
/inner_things/1/1/foo: too long, may not be more than 12 characters
`
	v := &Outer2DSliceThing{}
	original := `{"inner_things": [[{"foo": "fooz"}, {"foo": "fooziswaytoolooong2"}], [{"foo": "fooziswaytoolooong"},{"foo": "fooziswaytoolooongagain"}]]}`
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(original), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterSliceThingNotAList(t *testing.T) {
	expected := `Validation Errors: 
/inner_things: expected a list
`
	v := &OuterSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": "foo"}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterSliceThingOverMax(t *testing.T) {
	expected := `Validation Errors: 
/inner_things: must have at most 2 elements
`
	v := &OuterMaxSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}, {"foo": "fooz2"}, {"foo": "fooz3"}]}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterSliceThingUnderMax(t *testing.T) {
	v := &OuterMaxSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}, {"foo": "fooz2"}]}`), v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateOuterSliceThingUnderMin(t *testing.T) {
	expected := `Validation Errors: 
/inner_things: must have at least 2 elements
`
	v := &OuterMinSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}]}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterSliceThingOverMin(t *testing.T) {
	v := &OuterMinSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}, {"foo": "fooz2"}]}`), v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateOuterRangeSliceThingUnderMin(t *testing.T) {
	expected := `Validation Errors: 
/inner_things: must have between 1 and 2 elements
`
	v := &OuterRangeSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": []}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterRangeSliceThingOverMax(t *testing.T) {
	expected := `Validation Errors: 
/inner_things: must have between 1 and 2 elements
`
	v := &OuterRangeSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}, {"foo": "fooz2"}, {"foo": "fooz3"}]}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateOuterRangeSliceThingInRange(t *testing.T) {
	v := &OuterRangeSliceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_things": [{"foo": "fooz"}, {"foo": "fooz2"}]}`), v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateReadOnlyThing(t *testing.T) {
	v := &ReadOnlyThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"primary_key": "foo"}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if v.PrimaryKey != "" {
		t.Fatal("ReadOnly field unexpectedly set")
	}
}

func TestValidateReadOnlyThingValueNotProvided(t *testing.T) {
	v := &ReadOnlyThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if v.PrimaryKey != "" {
		t.Fatal("ReadOnly field unexpectedly set")
	}
}

func TestValidateUnregisteredThing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("No panic")
		}
	}()
	v := &UnregisteredThing{}
	TestTypeMapper.Unmarshal(EmptyContext, []byte(`{}`), v)
	t.Fatal("Unexpected success")
}

func TestValidateStringTypeMismatch(t *testing.T) {
	expected := `Validation Errors: 
/foo: not a string
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"foo": 12.0}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateStringTooShort(t *testing.T) {
	expected := `Validation Errors: 
/foo: too short, must be at least 1 characters
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"foo": ""}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateStringTooLong(t *testing.T) {
	expected := `Validation Errors: 
/foo: too long, may not be more than 12 characters
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"foo": "thisvalueistoolong"}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateBooleanTypeMismatch(t *testing.T) {
	expected := `Validation Errors: 
/a_bool: not a boolean
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"a_bool": 12.0}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateIntegerTypeMismatch(t *testing.T) {
	expected := `Validation Errors: 
/an_int: not an integer
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"an_int": false}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateIntegerNumericTypeMismatch(t *testing.T) {
	expected := `Validation Errors: 
/an_int: not an integer
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"an_int": 12.1}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateIntegerTooSmall(t *testing.T) {
	expected := `Validation Errors: 
/an_int: too small, must be at least 0
`

	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"an_int": -1}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateIntegerTooLarge(t *testing.T) {
	expected := `Validation Errors: 
/an_int: too large, may not be larger than 10
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"an_int": 2048}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateMultipleTypeMismatch(t *testing.T) {
	expected := `Validation Errors: 
/an_int: too large, may not be larger than 10
/a_bool: not a boolean
`
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"an_int": 2048, "a_bool": 12.0}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateMapOfInnerThing(t *testing.T) {
	expected1 := `Validation Errors: 
/inner_thing_map/key1/an_int: too large, may not be larger than 10
/inner_thing_map/key1/a_bool: not a boolean
/inner_thing_map/key2/an_int: too large, may not be larger than 10
/inner_thing_map/key2/a_bool: not a boolean
`
	expected2 := `Validation Errors: 
/inner_thing_map/key2/an_int: too large, may not be larger than 10
/inner_thing_map/key2/a_bool: not a boolean
/inner_thing_map/key1/an_int: too large, may not be larger than 10
/inner_thing_map/key1/a_bool: not a boolean
`
	v := &OuterInnerThingMap{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_thing_map":{"key1":{"an_int": 2048, "a_bool": 12.0}, "key2":{"an_int": 2048, "a_bool": 12.0}}}`), v)

	if err.Error() != expected1 && err.Error() != expected2 {
		t.Fatal("Unexpected error message:", err.Error())
	}
}

func TestValidateMapOfInnerThingFirstEntryValid(t *testing.T) {
	expected := `Validation Errors: 
/inner_thing_map/key2/an_int: too large, may not be larger than 10
/inner_thing_map/key2/a_bool: not a boolean
`
	v := &OuterInnerThingMap{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_thing_map":{"key1":{"an_int": 5, "a_bool": true}, "key2":{"an_int": 2048, "a_bool": 12.0}}}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateWithUnexpectedError(t *testing.T) {
	expected := `Validation Errors: 
/invalid: this should be a ValidationError
`
	v := &BrokenThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"invalid": "definitely"}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateThingWithMapOfStrings(t *testing.T) {
	expected := `Validation Errors: 
/strings/key1: too long, may not be more than 5 characters
`
	original := `{"strings":{"key1":"tooooooolongomg"}}`
	v := &ThingWithMapOfStrings{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(original), v)
	require.EqualError(t, err, expected)
}

func TestUnmarshalVariableTypeThing(t *testing.T) {
	{
		v := &OuterVariableThing{}
		err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"foo","inner_thing":{"foo":"bar"}}`), v)
		if err != nil {
			t.Fatal(err)
		}
		if v.InnerType != "foo" {
			t.Fatal("Unexpected value of InnerType:", v.InnerType)
		}
		it, ok := v.InnerValue.(*InnerThing)
		if !ok {
			t.Fatal("InnerValue has the wrong type:", reflect.TypeOf(v.InnerValue).String())
		}
		if it.Foo != "bar" {
			t.Fatal("Unexpected value of InnerThing.Foo:", it.Foo)
		}
	}
	{
		v := &OuterVariableThing{}
		err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"bar","inner_thing":{"bar":"foo"}}`), v)
		if err != nil {
			t.Fatal(err)
		}
		if v.InnerType != "bar" {
			t.Fatal("Unexpected value of InnerType:", v.InnerType)
		}
		it, ok := v.InnerValue.(*OtherInnerThing)
		if !ok {
			t.Fatal("InnerValue has the wrong type:", reflect.TypeOf(v.InnerValue).String())
		}
		if it.Bar != "foo" {
			t.Fatal("Unexpected value of InnerThing.Foo:", it.Bar)
		}
	}
}

func TestValidateVariableTypeThing(t *testing.T) {
	expected := `Validation Errors: 
/inner_thing: invalid type identifier: 'unknown'
`
	v := &OuterVariableThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"unknown","inner_thing":{"foo":"bar"}}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateVariableTypeWithSwitchFieldValidationError(t *testing.T) {
	expected := `Validation Errors: 
/inner_type: Value must be one of: ["these","are","allowed"]
/inner_thing: cannot validate, invalid input for 'inner_type'
`
	v := &OuterVariableThingInnerTypeOneOf{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"unknown","inner_thing":{"foo":"bar"}}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateVariableTypeSwitchFieldNoJsonTag(t *testing.T) {
	expected := `Validation Errors: 
/inner_type: Value must be one of: ["these","are","allowed"]
/inner_thing: invalid type identifier
`
	v := &OuterVariableThingInnerTypeNoJsonTag{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"unknown","inner_thing":{"foo":"bar"}}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateVariableTypeSwitchFieldIgnoredJsonTag(t *testing.T) {
	expected := `Validation Errors: 
/inner_type: Value must be one of: ["these","are","allowed"]
/inner_thing: invalid type identifier
`
	v := &OuterVariableThingInnerTypeIgnoredJsonTag{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"unknown","inner_thing":{"foo":"bar"}}`), v)
	require.EqualError(t, err, expected)
}

func TestValidateNotAnObject(t *testing.T) {
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`[1, 2, 3]`), v)
	require.EqualError(t, err, "json: cannot unmarshal, not an object")
}

func TestUnmarshalList(t *testing.T) {
	v := &InnerThing{}
	err := InnerThingTypeMap.Unmarshal(EmptyContext, nil, []interface{}{}, reflect.ValueOf(v))
	if err == nil {
		t.Fatal("Unexpected success")
	}
	if err.Error() != "expected an object" {
		t.Fatal("Unexpected error message:", err.Error())
	}
}

func TestUnmarshalMissingRequiredField(t *testing.T) {
	expected := `Validation Errors: 
/inner_thing: missing required field
`
	v := &OuterThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{}`), v)
	require.EqualError(t, err, expected)
}

func TestUnmarshalNonPointer(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("No panic")
		}
		if r != "cannot unmarshal to non-pointer" {
			t.Fatal("Incorrect panic message", r)
		}
	}()
	v := InnerThing{}
	TestTypeMapper.Unmarshal(EmptyContext, []byte(`{}`), v)
}

func TestMarshalInnerThing(t *testing.T) {
	v := &InnerThing{
		Foo:   "bar",
		AnInt: 7,
		ABool: true,
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"foo":"bar","an_int":7,"a_bool":true}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalOuterThing(t *testing.T) {
	v := &OuterThing{
		InnerThing: InnerThing{
			Foo:   "bar",
			AnInt: 3,
			ABool: false,
		},
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_thing":{"foo":"bar","an_int":3,"a_bool":false}}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalOuterPointerThing(t *testing.T) {
	v := &OuterPointerThing{
		InnerThing: &InnerThing{
			Foo:   "bar",
			AnInt: 3,
			ABool: false,
		},
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_thing":{"foo":"bar","an_int":3,"a_bool":false}}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestUnmarshalOuterPointerThingWithNull(t *testing.T) {
	v := &OuterPointerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_thing": null}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if v.InnerThing != nil {
		t.Fatal("Expected InnerThing to be nil")
	}
}

func TestMarshalOuterInterfaceThing(t *testing.T) {
	v := &OuterInterfaceThing{
		InnerThing: &InnerThing{
			Foo:   "bar",
			AnInt: 3,
			ABool: false,
		},
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_thing":{"foo":"bar","an_int":3,"a_bool":false}}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestUnmarshalOuterInterfaceThing(t *testing.T) {
	v := &OuterInterfaceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_thing": {"foo":"bar","an_int":3,"a_bool":false}}`), v)
	if err != nil {
		t.Fatal(err)
	}

	innerThing, ok := v.InnerThing.(*InnerThing)
	if !ok {
		t.Fatal("InnerThing has an unexpected type")
	}

	if innerThing.Foo != "bar" {
		t.Fatal("InnerThing.Bar has an unexpected value")
	}

	if innerThing.AnInt != 3 {
		t.Fatal("InnerThing.AnInt has an unexpected value")
	}

	if innerThing.ABool != false {
		t.Fatal("InnerThing.ABool has an unexpected value")
	}
}

func TestUnmarshalOuterInterfaceThingWithNull(t *testing.T) {
	v := &OuterInterfaceThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_thing": null}`), v)
	if err != nil {
		t.Fatal(err)
	}
	if v.InnerThing != nil {
		t.Fatal("Expected InnerThing to be nil")
	}
}

func TestMarshalOuterSliceThing(t *testing.T) {
	v := &OuterSliceThing{
		InnerThings: []InnerThing{
			{
				Foo:   "bar",
				AnInt: 3,
				ABool: false,
			},
		},
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_things":[{"foo":"bar","an_int":3,"a_bool":false}]}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}

}

func TestMarshalOuterPointerSliceThing(t *testing.T) {
	v := &OuterPointerSliceThing{
		InnerThings: []*InnerThing{
			{
				Foo:   "bar",
				AnInt: 3,
				ABool: false,
			},
		},
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_things":[{"foo":"bar","an_int":3,"a_bool":false}]}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalOuterPointerToSliceThing(t *testing.T) {
	v := &OuterPointerToSliceThing{
		InnerThings: &[]InnerThing{
			{
				Foo:   "bar",
				AnInt: 3,
				ABool: false,
			},
		},
	}
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_things":[{"foo":"bar","an_int":3,"a_bool":false}]}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalVariableTypeThing(t *testing.T) {
	{
		v := &OuterVariableThing{
			InnerType: "foo",
			InnerValue: &InnerThing{
				Foo: "test",
			},
		}

		data, err := TestTypeMapper.Marshal(EmptyContext, v)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `{"inner_type":"foo","inner_thing":{"foo":"test","an_int":0,"a_bool":false}}` {
			t.Fatal("Unexpected Marshal output:", string(data))
		}
	}
	{
		v := &OuterVariableThing{
			InnerType: "bar",
			InnerValue: &OtherInnerThing{
				Bar: "test",
			},
		}

		data, err := TestTypeMapper.Marshal(EmptyContext, v)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `{"inner_type":"bar","inner_thing":{"bar":"test"}}` {
			t.Fatal("Unexpected Marshal output:", string(data))
		}
	}
}

func TestMarshalVariableTypeThingIntegerInvalid(t *testing.T) {
	v := &OuterVariableThingInnerTypeOneOf{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"these","inner_thing":15}`), v)

	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_type":"these","inner_thing":null}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalVariableTypeThingIntegerValid(t *testing.T) {
	v := &OuterVariableThingInnerTypeOneOf{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"these","inner_thing":5}`), v)

	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_type":"these","inner_thing":5}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalVariableTypeThingIntegerValidZeroCase(t *testing.T) {
	v := &OuterVariableThingInnerTypeOneOf{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_type":"these","inner_thing":0}`), v)

	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inner_type":"these","inner_thing":0}` {
		t.Fatal("Unexpected Marshal output:", string(data))
	}
}

func TestMarshalBrokenVariableTypeThing(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("No panic")
		}
		if r != "no such underlying field: InnerTypeo" {
			t.Fatal("Incorrect panic message", r)
		}
	}()

	v := &OtherOuterVariableThing{
		InnerType: "foo",
		InnerValue: &InnerThing{
			Foo: "test",
		},
	}

	TestTypeMapper.Marshal(EmptyContext, v)
}

func TestMarshalVariableTypeThingInvalidTypeIdentifier(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("No panic")
		}
		if r != "variable type serialization error: invalid type identifier: 'wrong'" {
			t.Fatal("Incorrect panic message", r)
		}
	}()

	v := &OuterVariableThing{
		InnerType: "wrong",
		InnerValue: &InnerThing{
			Foo: "test",
		},
	}

	TestTypeMapper.Marshal(EmptyContext, v)
}

func TestMarshalNoSuchStructField(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("No panic")
		}
		if r != "no such underlying field: Incorrect" {
			t.Fatal("Incorrect panic message", r)
		}
	}()
	v := &TypoedThing{
		Correct: false,
	}
	TestTypeMapper.Marshal(EmptyContext, v)
}

func TestUnmarshalNoSuchStructField(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("No panic")
		}
		if r != "no such underlying field: Incorrect" {
			t.Fatal("Incorrect panic message", r)
		}
	}()
	v := &TypoedThing{}
	TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"correct": false}`), v)
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	v := &InnerThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"this is": "definitely invalid JSON]`), v)
	if err == nil {
		t.Fatal("Unexpected success")
	}
	if err.Error() != "unexpected end of JSON input" {
		t.Fatal("Unexpected error message:", err.Error())
	}
}

func TestMarshalNonMarshalableThing(t *testing.T) {
	v := &OuterNonMarshalableThing{}
	_, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err == nil {
		t.Fatal("Unexpected success")
	}
	if err.Error() != "json: error calling MarshalJSON for type jsonmap.NonMarshalableType: oops" {
		t.Fatal(err.Error())
	}
}

func TestMarshalSliceOfNonMarshalableThing(t *testing.T) {
	v := []OuterNonMarshalableThing{
		{},
	}
	_, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err == nil {
		t.Fatal("Unexpected success")
	}
	if err.Error() != "json: error calling MarshalJSON for type jsonmap.NonMarshalableType: oops" {
		t.Fatal(err.Error())
	}
}

func TestMarshalIndent(t *testing.T) {
	v := &OuterThing{
		InnerThing: InnerThing{
			Foo:   "bar",
			AnInt: 3,
			ABool: false,
		},
	}
	expected := "{\n" +
		"    \"inner_thing\": {\n" +
		"        \"foo\": \"bar\",\n" +
		"        \"an_int\": 3,\n" +
		"        \"a_bool\": false\n" +
		"    }\n" +
		"}"
	data, err := TestTypeMapper.MarshalIndent(EmptyContext, v, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestMarshalSlice(t *testing.T) {
	v := []InnerThing{
		{
			Foo:   "bar",
			AnInt: 3,
			ABool: false,
		},
		{
			Foo:   "bam",
			AnInt: 4,
			ABool: true,
		},
	}
	expected := `[{"foo":"bar","an_int":3,"a_bool":false},{"foo":"bam","an_int":4,"a_bool":true}]`
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestMarshalSliceOfPointers(t *testing.T) {
	v := []*InnerThing{
		&InnerThing{
			Foo:   "bar",
			AnInt: 3,
			ABool: false,
		},
		&InnerThing{
			Foo:   "bam",
			AnInt: 4,
			ABool: true,
		},
	}
	expected := `[{"foo":"bar","an_int":3,"a_bool":false},{"foo":"bam","an_int":4,"a_bool":true}]`
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestMarshalTemplatableThing(t *testing.T) {
	ctx := struct {
		Foo string
	}{
		Foo: "foo",
	}

	v := &TemplatableThing{
		SomeField: "bar",
	}

	expected := `{"some_field":"foo:bar"}`
	data, err := TestTypeMapper.Marshal(ctx, v)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestMarshalThingWithSliceOfPrimitives(t *testing.T) {
	v := ThingWithSliceOfPrimitives{
		Strings: []string{"foo", "bar"},
	}

	expected := `{"strings":["foo","bar"]}`
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestMarshalThingWithNilSliceOfPrimitives(t *testing.T) {
	v := ThingWithSliceOfPrimitives{}

	expected := `{"strings":null}`
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestValidateThingWithSliceOfPrimitives(t *testing.T) {
	original := `{"strings":["foo","bar"]}`
	v := &ThingWithSliceOfPrimitives{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(original), v)
	if err != nil {
		t.Fatal(err)
	}

	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatal("Unoriginal Marshal output:", string(data), original)
	}
}

func TestValidateOuterMapThingNotAMap(t *testing.T) {
	expected := `Validation Errors: 
/inner_map: expected a map
`

	v := &OuterMapThing{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"inner_map": 3}`), v)
	require.EqualError(t, err, expected)
}

func TestMarshalThingWithMapOfInterfaces(t *testing.T) {
	interfaces := map[string]interface{}{
		"foo": "bar",
		"baz": 10,
		"qux": []string{"dang"},
	}

	v := ThingWithMapOfInterfaces{
		Interfaces: interfaces,
	}

	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(map[string]interface{}{"interfaces": interfaces})
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(expected) {
		t.Fatal("unexpected Marshal output", string(data), string(expected))
	}
}

func TestValidateThingWithMapOfInterfaces(t *testing.T) {
	original := `{"interfaces":{"baz":10,"dux":null,"foo":"bar","qux":["dang"]}}`
	v := &ThingWithMapOfInterfaces{}
	err := TestTypeMapper.Unmarshal(EmptyContext, []byte(original), v)
	if err != nil {
		t.Fatal(err)
	}

	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatal("Unoriginal Marshal output:", string(data), original)
	}
}

func TestMarshalThingWithTime(t *testing.T) {
	ts, err := time.Parse(time.RFC822, time.RFC822)
	if err != nil {
		panic(err)
	}

	v := ThingWithTime{
		HappenedAt: ts,
	}

	expected := `{"happened_at":"2006-01-02T15:04:00Z"}`
	data, err := TestTypeMapper.Marshal(EmptyContext, v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatal("Unexpected Marshal output:", string(data), expected)
	}
}

func TestUnmarshalThingWithTime(t *testing.T) {
	ts, err := time.Parse(time.RFC822, time.RFC822)
	if err != nil {
		panic(err)
	}

	v := &ThingWithTime{}

	err = TestTypeMapper.Unmarshal(EmptyContext, []byte(`{"happened_at":"2006-01-02T15:04:00Z"}`), v)
	if err != nil {
		t.Fatal(err)
	}

	if !ts.Equal(v.HappenedAt) {
		t.Fatal("Timestamp mismatch:", v.HappenedAt, ts)
	}
}

func TestGenericUnmarshalInvalidInput(t *testing.T) {
	invalidCases := []struct {
		Input        string
		Into         ThingWithEnumerableInterface
		ErrorMessage string
	}{
		{
			Input: `{"thanks": "baz"}`,
			Into:  ThingWithEnumerableInterface{},
			ErrorMessage: `Validation Errors: 
/thanks: Value must be one of: ["foo","bar"]
`,
		},
		{
			Input: `{"thanks": 12}`,
			Into:  ThingWithEnumerableInterface{},
			ErrorMessage: `Validation Errors: 
/thanks: not a string
`,
		},
	}

	for _, invalidCase := range invalidCases {
		dest := invalidCase.Into
		err := TestTypeMapper.Unmarshal(EmptyContext, []byte(invalidCase.Input), &dest)
		require.Error(t, err)
		require.Equal(t, invalidCase.ErrorMessage, err.Error())
	}
}

func TestValidThingWithEnumerableInterface(t *testing.T) {
	validCases := []struct {
		Input    string
		Expected ThingWithEnumerableInterface
	}{
		{
			Input: `{"thanks": "foo"}`,
			Expected: ThingWithEnumerableInterface{
				ThanksGo: "foo",
			},
		},
		{
			Input: `{"thanks": "bar"}`,
			Expected: ThingWithEnumerableInterface{
				ThanksGo: "bar",
			},
		},
	}

	for _, validCase := range validCases {
		dest := validCase.Expected
		err := TestTypeMapper.Unmarshal(EmptyContext, []byte(validCase.Input), &dest)
		require.Nil(t, err)
		require.EqualValues(t, validCase.Expected, dest)
	}
}

type dogStruct struct {
	Age      int
	Name     string
	Owners   []string
	IsDead   bool
	Birthday time.Time
	Location *string
}

// Ostensibly non-testing versions of this would have error checking and such

func intRangeFactory(min, max int64) func(int64) bool {
	return func(n int64) bool {
		return min <= n && n <= max
	}
}

func sliceRangeFactory(min, max int) func([]string) bool {
	return func(sli []string) bool {
		return min <= len(sli) && len(sli) <= max
	}
}

var dogParamMap = QueryMap{
	UnderlyingType: dogStruct{},
	Parameters: []MappedParameter{
		{
			StructFieldName: "Age",
			ParameterName:   "age",
			Mapper: IntQueryParameterMapper{
				Validators: []func(int64) bool{
					intRangeFactory(0, 100),
				},
			},
		},
		{
			StructFieldName: "Name",
			ParameterName:   "name",
			Mapper: StringQueryParameterMapper{
				[]func(string) bool{
					StringRangeValidator(1, 10),
					StringRegexValidator(regexp.MustCompile(".*")),
				},
			},
		},
		{
			StructFieldName: "Owners",
			ParameterName:   "owners",
			Mapper: StrSliceQueryParameterMapper{
				[]func([]string) bool{
					sliceRangeFactory(0, 3),
				},
				StringQueryParameterMapper{
					[]func(string) bool{
						StringRangeValidator(1, 10),
						StringRegexValidator(regexp.MustCompile("[a-z]")),
					},
				},
			},
		},
		{
			StructFieldName: "IsDead",
			ParameterName:   "is_dead",
			Mapper:          BoolQueryParameterMapper{},
		},
		{
			StructFieldName: "Birthday",
			ParameterName:   "birthday",
			Mapper:          TimeQueryParameterMapper{},
		},
		{
			StructFieldName: "Location",
			ParameterName:   "location",
			Mapper: StrPointerQueryParameterMapper{
				UnderlyingQueryParameterMapper: StringQueryParameterMapper{},
			},
		},
	},
}

type requestFilter struct {
	UUID   string
	Count  int
	States []string
	Search string
}

var requestFilterMapping = QueryMap{
	UnderlyingType: requestFilter{},
	Parameters: []MappedParameter{
		{
			StructFieldName: "UUID",
			ParameterName:   "uuid",
			Mapper: StringQueryParameterMapper{
				[]func(string) bool{
					StringRegexValidator(uuidRegex),
					utf8.ValidString,
				},
			},
		},
		{
			StructFieldName: "Count",
			ParameterName:   "count",
			Mapper: IntQueryParameterMapper{
				Validators: []func(int64) bool{
					intRangeFactory(0, 500),
				},
			},
		},

		{
			StructFieldName: "Search",
			ParameterName:   "search",
			Mapper: StringQueryParameterMapper{
				[]func(string) bool{
					utf8.ValidString,
				},
			},
		},
	},
}

func TestParamMapping(t *testing.T) {
	tt := time.Now()
	tb, _ := tt.MarshalText()
	urlQuery, _ := url.ParseQuery(`location=barcelona&owners=Alice&name=Spot&owners=Bob&age=10&is_dead=false&birthday=` + string(tb))
	dog := dogStruct{}

	err := dogParamMap.Decode(urlQuery, &dog)
	require.NoError(t, err)
	require.Equal(t, dog.Age, 10)
	require.Equal(t, dog.Name, "Spot")
	require.Equal(t, dog.IsDead, false)
	require.Equal(t, dog.Birthday.Format(time.RFC3339), tt.Format(time.RFC3339))
	require.EqualValues(t, dog.Owners, []string{"Alice", "Bob"})
	require.Equal(t, *dog.Location, "barcelona")

	newMap := make(map[string][]string)
	err = dogParamMap.Encode(dog, newMap)
	require.NoError(t, err)
	require.EqualValues(t, urlQuery, newMap)

	urlQuery, _ = url.ParseQuery(`count=38&uuid=00000000-0000-1000-9000-000000000000&search=foobar`)
	filter := requestFilter{}
	err = requestFilterMapping.Decode(urlQuery, &filter)
	require.NoError(t, err)
	require.Equal(t, 38, filter.Count)
	require.Equal(t, "foobar", filter.Search)
	require.Equal(t, "00000000-0000-1000-9000-000000000000", filter.UUID)

	urlQuery, _ = url.ParseQuery("count=-1&uuid=00000000-0000-1000-9000-000000000000&search=bar")
	err = requestFilterMapping.Decode(urlQuery, &filter)
	require.Error(t, err, "a validation test failed")
	urlQuery, _ = url.ParseQuery("count=1&uuid=00000000-0000-1000-9000-000000000000&search=\xDAbar")
	err = requestFilterMapping.Decode(urlQuery, &filter)
	require.Error(t, err, "a validation test failed")
}

func TestHeaderMap(t *testing.T) {
	header := http.Header{}
	header.Add("name", "spot")
	header.Add("owners", "alice")
	header.Add("owners", "bob")
	header.Add("is_dead", "false")
	header.Add("age", "10")

	dog := dogStruct{}
	err := dogParamMap.DecodeHeader(header, &dog)
	require.NoError(t, err)
	require.Equal(t, dog.Age, 10)
	require.Equal(t, dog.Name, "spot")
	require.Equal(t, dog.IsDead, false)
	require.EqualValues(t, dog.Owners, []string{"alice", "bob"})

	var newHeader http.Header
	newHeader = make(map[string][]string)
	err = dogParamMap.EncodeHeader(dog, newHeader)
	require.NoError(t, err)
}
