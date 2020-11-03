package gomongoschema_test

import (
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
