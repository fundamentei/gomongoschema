package gomongoschema

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// ErrNoSchemaAvailable indicates that there's no schema available for a given collection
	ErrNoSchemaAvailable = errors.New("schema is not available")
)

// IsNoSchemaAvailable tells if an error is related to no schema available
func IsNoSchemaAvailable(err error) bool {
	return err == ErrNoSchemaAvailable
}

// SchemaFetcherFn fetches the $jsonSchema of the specified collection. Fails in case it doesn't exist
type SchemaFetcherFn func(collection string) (string, error)

// NewMongoDriverSchemaFetcher ...
func NewMongoDriverSchemaFetcher(
	pullSpecifications func() ([]*mongo.CollectionSpecification, error),
) SchemaFetcherFn {
	return func(collection string) (string, error) {
		specifications, err := pullSpecifications()

		if err != nil {
			return "", err
		}

		var match *mongo.CollectionSpecification

		for _, spec := range specifications {
			if spec.Name == collection {
				match = spec
				break
			}
		}

		if match == nil {
			return "", fmt.Errorf("collection not found: %q", collection)
		}

		validator := match.Options.Lookup("validator")

		if _, ok := validator.DocumentOK(); !ok {
			return "", ErrNoSchemaAvailable
		}

		jsonSchema := validator.Document().Lookup("$jsonSchema")
		return jsonSchema.Document().String(), nil
	}
}

// Validator is a $jsonSchema validator
type Validator interface {
	Validate(collection string, document interface{}) error
}

type validator struct {
	fetchSchema SchemaFetcherFn
}

func (v *validator) Validate(collection string, document interface{}) error {
	schema, err := v.fetchSchema(collection)

	if err != nil {
		return err
	}

	bsondocument, err := bson.MarshalExtJSON(document, true, false)

	validationResult, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(BsonSchemaToJSONSchema(schema)),
		gojsonschema.NewBytesLoader(bsondocument),
	)

	if err != nil {
		return err
	}

	if validationResult.Valid() {
		return nil
	}

	var merr error

	for _, err := range validationResult.Errors() {
		merr = fmt.Errorf("%s: %w", err.Description(), merr)
	}

	return merr
}

// NewValidator is for creating a new validator
func NewValidator(fetchSchema SchemaFetcherFn) Validator {
	return &validator{
		fetchSchema: fetchSchema,
	}
}
