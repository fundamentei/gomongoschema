package gomongoschema_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/fundamentei/gomongoschema"
	"github.com/stretchr/testify/assert"
)

func TestBsonSchemaToJSONSchemaBsonTypeToType(t *testing.T) {
	assert.Equal(t, `{"type":"string"}`, gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"string"}`))
	assert.Equal(t, `{"type":"boolean"}`, gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"bool"}`))
}

func TestBsonSchemaToJSONSchemaBsonTypeToTypeRenames(t *testing.T) {
	assert.Equal(t, `{"type":"boolean"}`, gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"bool"}`))
	assert.Equal(t, `{"type":"number"}`, gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"double"}`))
	assert.Equal(t, `{"type":"number"}`, gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"decimal"}`))
}

func TestBsonSchemaToJSONSchemaBsonTypeToTypeRenamesArray(t *testing.T) {
	assert.Equal(
		t,
		`{"type":["boolean","number","number"]}`,
		gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":["bool","double","decimal"]}`),
	)
}

func TestBsonSchemaToJSONSchemaBsonTypeObjectId(t *testing.T) {
	assert.Equal(
		t,
		// string or {"$oid": "string"}
		`{"type":[{"type":"string"},{"additionalProperties":false,"properties":{"$oid":{"type":"string"}},"required":["$oid"],"type":"object"}]}`,
		gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"objectId"}`),
	)
}

func TestBsonSchemaToJSONSchemaBsonTypeDate(t *testing.T) {
	assert.Equal(
		t,
		`{"type":[{"format":"datetime","type":"string"}]}`,
		gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"date"}`),
	)
}

func TestBsonSchemaToJSONSchemaComplexSchema(t *testing.T) {
	assert.Equal(
		t,
		mustUglifyJSON(
			`
{
	"properties": {
		"_id": {
			"type": [
				{ "type": "string" },
				{
					"additionalProperties": false,
					"properties": { "$oid": { "type": "string" } },
					"required": ["$oid"],
					"type": "object"
				}
			]
		}
	},
	"required": ["_id"],
	"additionalProperties": false,
	"type": "object"
}
			`,
		),
		mustUglifyJSON(gomongoschema.BsonSchemaToJSONSchema(
			`
{
	"bsonType": "object",
	"properties": {
		"_id": {
			"bsonType": "objectId"
		}
	},
	"required": ["_id"],
	"additionalProperties": false
}
			`,
		)),
	)
}

func mustUglifyJSON(jstr string) string {
	bb := bytes.NewBuffer([]byte{})

	var jstrv interface{}

	if err := json.NewDecoder(strings.NewReader(jstr)).Decode(&jstrv); err != nil {
		panic(err)
	}

	if err := json.NewEncoder(bb).Encode(&jstrv); err != nil {
		panic(err)
	}

	return strings.TrimSpace(bb.String())
}
