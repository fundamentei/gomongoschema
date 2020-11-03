## `gomongoschema`

This package contains helpers (or whatever you call it) to deal with MongoDB's `$jsonSchema` validation messages. As of
today (**2020-03-10**) it lacks support for a proper error message for when the document fails an **insertion** or **update**.

This issue will be resolved by https://jira.mongodb.org/browse/SERVER-20547. However—it's only landing Mid-2021 if everything goes well.

## How does this work?

MongoDB has its own `bsonType` aside from the standard JSON Schema `type`, along with other specialties. The way this "helper" works is by converting the given Mongo `$jsonSchema` to a standard schema that can be used with any schema validation library. Such as [xeipuuv/gojsonschema](github.com/xeipuuv/gojsonschema).

_**Disclaimer**_: that's the most effective way I found to do it at the moment since we lack official support.
