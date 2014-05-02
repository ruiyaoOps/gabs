package gabs

import (
	"encoding/json"
	"testing"
)

func TestBasic(t *testing.T) {
	sample := []byte(`{"test":{"value":10},"test2":20}`)

	val, err := ParseJson(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if result, ok := val.Search([]string{"test", "value"}...).Data().(float64); ok {
		if result != 10 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find test.value")
	}

	if _, ok := val.Search("test2", "value").Data().(string); ok {
		t.Errorf("Somehow found a field that shouldn't exist")
	}

	if result, ok := val.Search("test2").Data().(float64); ok {
		if result != 20 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find test2")
	}
}

func TestModify(t *testing.T) {
	sample := []byte(`{"test":{"value":10},"test2":20}`)

	val, err := ParseJson(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if err := val.Set(45.0, "test", "value"); err != nil {
		t.Errorf("Failed to set field")
	}

	if result, ok := val.Search([]string{"test", "value"}...).Data().(float64); ok {
		if result != 45 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find test.value")
	}

	if out := val.String(); `{"test":{"value":45},"test2":20}` != out {
		t.Errorf("Incorrectly serialized: %v", out)
	}

	if out := val.Search("test").String(); `{"value":45}` != out {
		t.Errorf("Incorrectly serialized: %v", out)
	}
}

func TestArrays(t *testing.T) {
	json1, _ := ParseJson([]byte(`{
		"languages":{
			"english":{
				"places":0
			},
			"french": {
				"places": [
					"france",
					"belgium"
				]
			}
		}
	}`))

	json2, _ := ParseJson([]byte(`{
		"places":[
			"great_britain",
			"united_states_of_america",
			"the_world"
		]
	}`))

	if english_places := json2.Search("places").Data(); english_places != nil {
		json1.Set(english_places, "languages", "english", "places")
	} else {
		t.Errorf("Didn't find places in json2")
	}

	if english_places := json1.Search("languages", "english", "places").Data(); english_places != nil {

		english_array, ok := english_places.([]interface{})
		if !ok {
			t.Errorf("places in json1 (%v) was not an array", english_places)
		}

		if len(english_array) != 3 {
			t.Errorf("wrong length of array: %v", len(english_array))
		}

	} else {
		t.Errorf("Didn't find places in json1")
	}
}

func TestLargeSample(t *testing.T) {
	sample := []byte(`{
		"test":{
			"innerTest":{
				"value":10,
				"value2":22,
				"value3":{
					"moreValue":45
				}
			}
		},
		"test2":20
	}`)

	val, err := ParseJson(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if result, ok := val.Search("test", "innerTest", "value3", "moreValue").Data().(float64); ok {
		if result != 45 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find value")
	}
}

func TestShorthand(t *testing.T) {
	json, _ := ParseJson([]byte(`{
		"outter":{
			"inner":{
				"value":5,
				"value2":10,
				"value3":11
			},
			"inner2":{
			}
		},
		"outter2":{
			"inner":0
		}
	}`))

	missingValue := json.S("outter").S("doesntexist").S("alsodoesntexist").S("inner").S("value").Data()
	if missingValue != nil {
		t.Errorf("missing value was actually found: %v\n", missingValue)
	}

	realValue := json.S("outter").S("inner").S("value2").Data().(float64)
	if realValue != 10 {
		t.Errorf("real value was incorrect: %v\n", realValue)
	}

	err := json.S("outter2").Set(json.S("outter").S("inner").Data(), "inner")
	if err != nil {
		t.Errorf("error setting outter2: %v\n", err)
	}

	compare := `{"outter":{"inner":{"value":5,"value2":10,"value3":11},"inner2":{}}` +
		`,"outter2":{"inner":{"value":5,"value2":10,"value3":11}}}`
	out := json.String()
	if out != compare {
		t.Errorf("wrong serialized structure: %v\n", out)
	}

	compare2 := `{"outter":{"inner":{"value":6,"value2":10,"value3":11},"inner2":{}}` +
		`,"outter2":{"inner":{"value":6,"value2":10,"value3":11}}}`

	json.S("outter").S("inner").Set(6, "value")
	out = json.String()
	if out != compare2 {
		t.Errorf("wrong serialized structure: %v\n", out)
	}
}

func TestCreation(t *testing.T) {
	json, _ := ParseJson([]byte(`{}`))
	inner := json.C("test").C("inner")

	inner.Set(10, "first")
	inner.Set(20, "second")

	inner.A("array")
	inner.Push("first element of the array", "array")
	inner.Push(2, "array")
	inner.Push("three", "array")

	expected := `{"test":{"inner":{"array":["first element of the array",2,"three"],"first":10,"second":20}}}`
	actual := json.String()
	if actual != expected {
		t.Errorf("received incorrect output from json object: %v\n", actual)
	}
}

type outterJson struct {
	FirstInner  innerJson
	SecondInner innerJson
	ThirdInner  innerJson
}

type innerJson struct {
	NumberType float64
	StringType string
}

type jsonStructure struct {
	FirstOutter  outterJson
	SecondOutter outterJson
}

var jsonContent = []byte(`{
	"firstOutter":{
		"firstInner":{
			"numberType":11,
			"stringType":"hello world, first first"
		},
		"secondInner":{
			"numberType":12,
			"stringType":"hello world, first second"
		},
		"thirdInner":{
			"numberType":13,
			"stringType":"hello world, first third"
		}
	},
	"secondOutter":{
		"firstInner":{
			"numberType":21,
			"stringType":"hello world, second first"
		},
		"secondInner":{
			"numberType":22,
			"stringType":"hello world, second second"
		},
		"thirdInner":{
			"numberType":23,
			"stringType":"hello world, second third"
		}
	}
}`)

/*
Simple use case, compares unmarshalling declared structs vs dynamically searching for
the equivalent hierarchy. Hopefully we won't see too great a performance drop from the
dynamic approach.
*/

func BenchmarkStatic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var jsonObj jsonStructure
		json.Unmarshal(jsonContent, &jsonObj)

		if val := jsonObj.FirstOutter.SecondInner.NumberType; val != 12 {
			b.Errorf("Wrong value of FirstOutter.SecondInner.NumberType: %v\n", val)
		}
		expected := "hello world, first second"
		if val := jsonObj.FirstOutter.SecondInner.StringType; val != expected {
			b.Errorf("Wrong value of FirstOutter.SecondInner.StringType: %v\n", val)
		}
		if val := jsonObj.SecondOutter.ThirdInner.NumberType; val != 23 {
			b.Errorf("Wrong value of SecondOutter.ThirdInner.NumberType: %v\n", val)
		}
		expected = "hello world, second second"
		if val := jsonObj.SecondOutter.SecondInner.StringType; val != expected {
			b.Errorf("Wrong value of SecondOutter.SecondInner.StringType: %v\n", val)
		}
	}
}

func BenchmarkDynamic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jsonObj, err := ParseJson(jsonContent)
		if err != nil {
			b.Errorf("Error parsing json: %v\n", err)
		}

		FOSI := jsonObj.S("firstOutter", "secondInner")
		SOSI := jsonObj.S("secondOutter", "secondInner")
		SOTI := jsonObj.S("secondOutter", "thirdInner")

		if val := FOSI.S("numberType").Data().(float64); val != 12 {
			b.Errorf("Wrong value of FirstOutter.SecondInner.NumberType: %v\n", val)
		}
		expected := "hello world, first second"
		if val := FOSI.S("stringType").Data().(string); val != expected {
			b.Errorf("Wrong value of FirstOutter.SecondInner.StringType: %v\n", val)
		}
		if val := SOTI.S("numberType").Data().(float64); val != 23 {
			b.Errorf("Wrong value of SecondOutter.ThirdInner.NumberType: %v\n", val)
		}
		expected = "hello world, second second"
		if val := SOSI.S("stringType").Data().(string); val != expected {
			b.Errorf("Wrong value of SecondOutter.SecondInner.StringType: %v\n", val)
		}
	}
}