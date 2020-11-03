package gomongoschema_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fundamentei/gomongoschema"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		mustUglifyJSON(
			`
{
	"oneOf": [
		{ "type": "string" },
		{
			"additionalProperties": false,
			"properties": { "$oid": { "type": "string" } },
			"required": ["$oid"],
			"type": "object"
		}
	]
}
			`,
		),
		mustUglifyJSON(gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"objectId"}`)),
	)
}

func TestBsonSchemaToJSONSchemaBsonTypeDate(t *testing.T) {
	assert.Equal(
		t,
		mustUglifyJSON(
			`
{
	"oneOf": [
		{ "format": "datetime", "type": "string" },
		{
			"additionalProperties": false,
			"properties": {
				"$date": {
					"additionalProperties": false,
					"properties": { "$numberLong": { "type": ["string", "number"] } },
					"required": ["$numberLong"],
					"type": "object"
				}
			},
			"required": ["$date"],
			"type": "object"
		}
	]
}
			`,
		),
		mustUglifyJSON(gomongoschema.BsonSchemaToJSONSchema(`{"bsonType":"date"}`)),
	)
}

func TestBsonSchemaToJSONSchemaComplexSchema(t *testing.T) {
	assert.Equal(
		t,
		mustUglifyJSON(
			`
{
	"additionalProperties": false,
	"properties": {
		"_id": {
			"oneOf": [
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

func TestValidateSimpleSchema(t *testing.T) {
	assert.Nil(
		t,
		validateAgainstSchema(
			gomongoschema.BsonSchemaToJSONSchema(
				`
{
	"bsonType": "object",
	"required": ["_id", "firstName", "lastName", "createdAt", "updatedAt"],
	"additionalProperties": false,
	"properties": {
		"_id": {
			"bsonType": "objectId"
		},
		"firstName": {
			"bsonType": "string"
		},
		"lastName": {
			"bsonType": "string"
		},
		"createdAt": {
			"bsonType": "date"
		},
		"updatedAt": {
			"bsonType": ["date", "null"]
		}
	}
}
				`,
			),
			&bson.M{
				"_id":       primitive.NewObjectID(),
				"firstName": "John",
				"lastName":  "Connor",
				"createdAt": time.Now(),
				"updatedAt": nil,
			},
		),
	)
}

func TestFailsSchemaValidationMissingFirstName(t *testing.T) {
	assert.NotNil(
		t,
		validateAgainstSchema(
			gomongoschema.BsonSchemaToJSONSchema(
				`
{
	"bsonType": "object",
	"required": ["_id", "firstName", "lastName", "createdAt", "updatedAt"],
	"additionalProperties": false,
	"properties": {
		"_id": {
			"bsonType": "objectId"
		},
		"firstName": {
			"bsonType": "string"
		},
		"lastName": {
			"bsonType": "string"
		},
		"createdAt": {
			"bsonType": "date"
		},
		"updatedAt": {
			"bsonType": ["date", "null"]
		}
	}
}
				`,
			),
			&bson.M{
				"_id":       primitive.NewObjectID(),
				"firstName": nil,
				"lastName":  "Connor",
				"createdAt": time.Now(),
				"updatedAt": nil,
			},
		),
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

func validateAgainstSchema(schema string, doc interface{}) error {
	bsondoc, err := bson.MarshalExtJSON(doc, true, false)

	if err != nil {
		return err
	}

	jsonschema := gomongoschema.BsonSchemaToJSONSchema(schema)

	schemaLoader := gojsonschema.NewStringLoader(jsonschema)
	documentLoader := gojsonschema.NewBytesLoader(bsondoc)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	var merr error

	for _, err := range result.Errors() {
		merr = fmt.Errorf("%s: %w", err.Description(), merr)
	}

	return merr
}
