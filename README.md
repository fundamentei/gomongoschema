## `gomongoschema`

This package contains helpers (or whatever you call it) to deal with MongoDB's `$jsonSchema` validation messages. As of
today (**2020-03-10**) it lacks support for a proper error message for when the document fails an **insertion** or **update**.

This issue will be resolved by https://jira.mongodb.org/browse/SERVER-20547. However—it's only landing Mid-2021 if everything goes well.

## How does this work?

MongoDB has its own `bsonType` aside from the standard JSON Schema `type`, along with other specialties. The way this "helper" works is by converting the given Mongo `$jsonSchema` to a standard schema that can be used with any schema validation library. Such as [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema).

_**Disclaimer**_: that's the most effective way I found to do it at the moment since we lack official support.

## Installation

```SH
go get -u github.com/fundamentei/gomongoschema
```

## How to use?

This package won't abstract the database usage for you. You'll need to find a way to pull the `$jsonSchema` out of the collection you're validating the document for. And also choose
the library that you'll use to perform the schema validation.

### The easy way

```Go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fundamentei/gomongoschema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		panic(err)
	}

	defer client.Disconnect(ctx)

	if err != nil {
		panic(err)
	}

	validator := gomongoschema.NewValidator(
		gomongoschema.NewMongoDriverSchemaFetcher(func() ([]*mongo.CollectionSpecification, error) {
			db := client.Database("gomongoschema")

			return db.ListCollectionSpecifications(
				ctx,
				bson.M{},
				&options.ListCollectionsOptions{},
			)
		}),
	)

	verr := validator.Validate("users", &bson.M{
		"_id":       primitive.NewObjectID(),
		"firstName": "John",
		"createdAt": nil,
	})

	fmt.Println(verr)
}
```

### The hard way

```Go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fundamentei/gomongoschema"
	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		panic(err)
	}

	defer client.Disconnect(ctx)

	usersJSONSchema, err := getCollectionJSONSchema(client.Database("gomongoschema"), "users")

	if err != nil {
		panic(err)
	}

	validate(usersJSONSchema, &bson.M{
		"_id":       primitive.NewObjectID(),
		"firstName": "John",
		"lastName":  "Connor",
	})
}

func validate(bsonSchema string, doc interface{}) {
	docb, err := bson.MarshalExtJSON(doc, true, false)

	if err != nil {
		panic(err)
	}

	schemaLoader := gojsonschema.NewStringLoader(gomongoschema.BsonSchemaToJSONSchema(bsonSchema))
	documentLoader := gojsonschema.NewBytesLoader(docb)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Println("Document is valid")
	} else {
		fmt.Println("Invalid document")

		for _, err := range result.Errors() {
			fmt.Println(err)
		}
	}
}

func getCollectionJSONSchema(db *mongo.Database, collection string) (string, error) {
	specs, err := db.ListCollectionSpecifications(
		context.Background(),
		bson.M{"name": collection},
		&options.ListCollectionsOptions{},
	)

	if err != nil {
		return "", err
	}

	var match *mongo.CollectionSpecification

	for _, spec := range specs {
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
		return "", fmt.Errorf("%q option not found", "validator")
	}

	jsonSchema := validator.Document().Lookup("$jsonSchema")
	return jsonSchema.Document().String(), nil
}
```

```JS
db.runCommand({
  collMod: "users",
  validator: {
    $jsonSchema: {
      bsonType: "object",
      properties: {
        firstName: {
          bsonType: "string",
        },
        lastName: {
          bsonType: "string",
        },
      },
      required: ["firstName", "lastName"],
      additionalProperties: true,
    },
  },
  validationLevel: "strict",
  validationAction: "error",
});
```
