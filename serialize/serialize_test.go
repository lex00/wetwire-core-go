package serialize

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type TestStruct struct {
	FirstName string
	LastName  string
	Age       int
	IsActive  bool
	Tags      []string
}

type NestedStruct struct {
	ID     string
	Config ConfigStruct
}

type ConfigStruct struct {
	MaxRetries  int
	EnableCache bool
}

func TestToMapSnakeCase(t *testing.T) {
	s := TestStruct{FirstName: "John", LastName: "Doe", Age: 30}
	m := ToMap(s, SnakeCase)
	assert.Equal(t, "John", m["first_name"])
	assert.Equal(t, "Doe", m["last_name"])
	assert.Equal(t, 30, m["age"])
}

func TestToMapCamelCase(t *testing.T) {
	s := TestStruct{FirstName: "John", LastName: "Doe", Age: 30}
	m := ToMap(s, CamelCase)
	assert.Equal(t, "John", m["firstName"])
	assert.Equal(t, "Doe", m["lastName"])
	assert.Equal(t, 30, m["age"])
}

func TestToMapPascalCase(t *testing.T) {
	s := TestStruct{FirstName: "John", LastName: "Doe", Age: 30}
	m := ToMap(s, PascalCase)
	assert.Equal(t, "John", m["FirstName"])
	assert.Equal(t, "Doe", m["LastName"])
	assert.Equal(t, 30, m["Age"])
}

func TestToMapOmitEmpty(t *testing.T) {
	s := TestStruct{FirstName: "John"} // other fields are zero values
	m := ToMap(s, SnakeCase, OmitEmpty)
	assert.Equal(t, "John", m["first_name"])
	_, hasLastName := m["last_name"]
	assert.False(t, hasLastName, "empty string should be omitted")
	_, hasAge := m["age"]
	assert.False(t, hasAge, "zero int should be omitted")
}

func TestToMapNested(t *testing.T) {
	n := NestedStruct{
		ID: "123",
		Config: ConfigStruct{
			MaxRetries:  3,
			EnableCache: true,
		},
	}
	m := ToMap(n, SnakeCase)
	assert.Equal(t, "123", m["id"])
	config, ok := m["config"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 3, config["max_retries"])
	assert.Equal(t, true, config["enable_cache"])
}

func TestToYAML(t *testing.T) {
	s := TestStruct{FirstName: "John", LastName: "Doe", Age: 30}
	data, err := ToYAML(s, SnakeCase)
	require.NoError(t, err)

	var parsed map[string]any
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "John", parsed["first_name"])
	assert.Equal(t, "Doe", parsed["last_name"])
}

func TestToJSON(t *testing.T) {
	s := TestStruct{FirstName: "John", LastName: "Doe", Age: 30}
	data, err := ToJSON(s, CamelCase)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "John", parsed["firstName"])
	assert.Equal(t, "Doe", parsed["lastName"])
}

func TestToJSONOmitEmpty(t *testing.T) {
	s := TestStruct{FirstName: "John"}
	data, err := ToJSON(s, SnakeCase, OmitEmpty)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	_, hasLastName := parsed["last_name"]
	assert.False(t, hasLastName)
}

func TestToMapWithSlice(t *testing.T) {
	s := TestStruct{FirstName: "John", Tags: []string{"admin", "user"}}
	m := ToMap(s, SnakeCase, OmitEmpty)
	assert.Equal(t, "John", m["first_name"])
	assert.Equal(t, []string{"admin", "user"}, m["tags"])
}
