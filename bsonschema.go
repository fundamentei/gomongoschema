package gomongoschema

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// BsonSchemaToJSONSchema this method is responsible for converting the Mongo JSON schema to be compatible with the std
// JSON schema. MongoDB's implementation of JSON Schema includes the addition of the bsonType keyword, which allows you
// to use all BSON types in the $jsonSchema operator - this func will make everything compatible
func BsonSchemaToJSONSchema(bsonSchema string) string {
	modifiedSchema := "{}"
	gjs := gjson.Parse(bsonSchema)

	// TODO(andreyvital): refactor this later
	objectIDType := []interface{}{
		map[string]string{
			"type": "string",
		},
		map[string]interface{}{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"$oid"},
			"properties": map[string]interface{}{
				"$oid": map[string]string{
					"type": "string",
				},
			},
		},
	}

	numberLongType := []interface{}{
		map[string]string{
			"type": "number",
		},
		map[string]interface{}{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"$numberLong"},
			"properties": map[string]interface{}{
				"$numberLong": map[string][]string{
					"type": {"string", "number"},
				},
			},
		},
	}

	dateType := []interface{}{
		map[string]string{
			"type":   "string",
			"format": "datetime",
		},
		map[string]interface{}{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"$date"},
			"properties": map[string]interface{}{
				"$date": map[string]interface{}{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"$numberLong"},
					"properties": map[string]interface{}{
						"$numberLong": map[string][]string{
							"type": {"string", "number"},
						},
					},
				},
			},
		},
	}

	timestampType := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"$timestamp"},
		"properties": map[string]interface{}{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"t", "i"},
			"properties": map[string]interface{}{
				"t": map[string]string{"type": "number"},
				"i": map[string]string{"type": "number"},
			},
		},
	}

	regexType := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"$regex", "$options"},
		"properties": map[string]interface{}{
			"$regex":   map[string]string{"type": "string"},
			"$options": map[string]string{"type": "string"},
		},
	}

	for k, v := range gjs.Map() {
		// https://docs.mongodb.com/manual/reference/operator/query/type/#document-type-available-types
		if k == "bsonType" {
			// handling for special BSON types (objectId, date, long, timestamp, regex)
			if v.Type == gjson.String && v.String() == "objectId" {
				modifiedSchema = mustSetJSON(modifiedSchema, "oneOf", objectIDType)
				continue
			}

			if v.Type == gjson.String && v.String() == "date" {
				modifiedSchema = mustSetJSON(modifiedSchema, "oneOf", dateType)
				continue
			}

			if v.Type == gjson.String && v.String() == "long" {
				modifiedSchema = mustSetJSON(modifiedSchema, "oneOf", numberLongType)
				continue
			}

			if v.Type == gjson.String && v.String() == "timestamp" {
				modifiedSchema = mustSetJSON(modifiedSchema, "oneOf", timestampType)
				continue
			}

			if v.Type == gjson.String && v.String() == "regex" {
				modifiedSchema = mustSetJSON(modifiedSchema, "oneOf", regexType)
				continue
			}

			// BSON type renames (bool, double, decimal)
			if v.Type == gjson.String && v.String() == "bool" {
				modifiedSchema = mustSetJSON(modifiedSchema, "type", "boolean")
				continue
			}

			if v.Type == gjson.String && (v.String() == "double" || v.String() == "decimal") {
				modifiedSchema = mustSetJSON(modifiedSchema, "type", "number")
				continue
			}

			if v.IsArray() {
				oneOfSimpleTypes := []string{}
				oneOfComplexTypes := []interface{}{}

				for _, tv := range v.Array() {
					if tv.Type != gjson.String {
						// don't know exactly what to do here
						continue
					}

					if tv.String() == "objectId" {
						oneOfComplexTypes = append(oneOfComplexTypes, objectIDType...)
						continue
					}

					if tv.String() == "date" {
						oneOfComplexTypes = append(oneOfComplexTypes, dateType...)
						continue
					}

					if tv.String() == "long" {
						oneOfComplexTypes = append(oneOfComplexTypes, numberLongType...)
						continue
					}

					if tv.String() == "timestamp" {
						oneOfComplexTypes = append(oneOfComplexTypes, timestampType)
						continue
					}

					if tv.String() == "regex" {
						oneOfComplexTypes = append(oneOfComplexTypes, regexType)
						continue
					}

					// BSON type renames (bool, double, decimal)
					simpleType := tv.String()

					if simpleType == "bool" {
						simpleType = "boolean"
					} else if simpleType == "double" || simpleType == "decimal" {
						simpleType = "number"
					}

					oneOfSimpleTypes = append(oneOfSimpleTypes, simpleType)
				}

				if len(oneOfComplexTypes) > 0 {
					for _, simpleType := range oneOfSimpleTypes {
						oneOfComplexTypes = append(oneOfComplexTypes, map[string]string{
							"type": simpleType,
						})
					}

					modifiedSchema = mustSetJSON(modifiedSchema, "oneOf", oneOfComplexTypes)
				} else {
					modifiedSchema = mustSetJSON(modifiedSchema, "type", oneOfSimpleTypes)
				}
				continue
			}

			modifiedSchema = mustSetJSON(modifiedSchema, "type", v.String())
			continue
		}

		if v.IsObject() {
			gjs := gjson.Parse(BsonSchemaToJSONSchema(v.Raw))
			modifiedSchema = mustSetJSON(modifiedSchema, k, gjs.Value())
			continue
		}

		modifiedSchema = mustSetJSON(modifiedSchema, k, v.Value())
	}

	return modifiedSchema
}

// TODO(andreyvital): refactor this - don't panic
func mustSetJSON(input string, key string, value interface{}) string {
	json, err := sjson.Set(input, key, value)

	if err != nil {
		panic(err)
	}

	return json
}
