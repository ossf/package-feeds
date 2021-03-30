package feeds

import (
	"fmt"
	"testing"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

const schema_path = "../package.schema.json"

type extendPackage struct {
	*Package
	NonConformingField string `json:"non_conforming_field"`
}

var (
	schemaLoader = gojsonschema.NewReferenceLoader("file://" + schema_path)
	dummyPackage = &Package{
		Name:        "foobarpackage",
		Version:     "1.0.0",
		CreatedDate: time.Now().UTC(),
		Type:        "npm",
		SchemaVer:   schemaVer,
	}
)

func TestValidSchema(t *testing.T) {
	t.Parallel()

	validPackage := gojsonschema.NewGoLoader(dummyPackage)
	result, err := gojsonschema.Validate(schemaLoader, validPackage)

	if err != nil {
		panic(err.Error())
	}

	if result.Valid() != true {
		out := "The Package json is not valid against the current schema. see errors :\n"
		for _, desc := range result.Errors() {
			out += fmt.Sprintf("- %s\n", desc)
		}
		t.Fatalf(out)
	}
}

func TestInvalidSchema(t *testing.T) {
	t.Parallel()

	// The Schema defines that additional properties are not valid, ensure enforcement
	// against an extra struct field. If an extra field is added, the SchemVer minor should
	// be incremented to advertise an additive change.
	invalid_package := &extendPackage{dummyPackage, "extrafield"}
	invalidField := gojsonschema.NewGoLoader(invalid_package)
	result, err := gojsonschema.Validate(schemaLoader, invalidField)

	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		t.Fatalf("Non-conformant extra field incorrectly validated")
	}

	// The Schema defines a required pattern for the schema_ver, ensure enforcement against
	// empty string.
	dummyPackage.SchemaVer = ""
	invalidFormat := gojsonschema.NewGoLoader(dummyPackage)
	result, err = gojsonschema.Validate(schemaLoader, invalidFormat)

	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		t.Fatalf("Non-conformant field format incorrectly validated")
	}

}
