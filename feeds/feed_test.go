package feeds

import (
	"fmt"
	"testing"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

const schemaPath = "../package.schema.json"

type extendPackage struct {
	Package
	NonConformingField string `json:"non_conforming_field"`
}

var (
	schemaLoader = gojsonschema.NewReferenceLoader("file://" + schemaPath)
	dummyPackage = Package{
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
		t.Fatal(err)
	}

	if result.Valid() != true {
		out := "The Package json is not valid against the current schema. see errors :\n"
		for _, desc := range result.Errors() {
			out += fmt.Sprintf("- %s\n", desc)
		}
		t.Fatal(out)
	}
}

func TestInvalidSchema(t *testing.T) {
	t.Parallel()

	// The Schema defines that additional properties are not valid, ensure enforcement
	// against an extra struct field. If an extra field is added, the SchemVer minor should
	// be incremented to advertise an additive change.
	invalidPackageField := extendPackage{dummyPackage, "extrafield"}
	invalidField := gojsonschema.NewGoLoader(invalidPackageField)
	result, err := gojsonschema.Validate(schemaLoader, invalidField)
	if err != nil {
		t.Fatal(err)
	}

	if result.Valid() {
		t.Fatalf("Non-conformant extra field incorrectly validated")
	}

	// The Schema defines a required pattern for the schema_ver, ensure enforcement against
	// empty string.
	invalidPackageFormat := dummyPackage
	invalidPackageFormat.SchemaVer = ""
	invalidFormat := gojsonschema.NewGoLoader(invalidPackageFormat)
	result, err = gojsonschema.Validate(schemaLoader, invalidFormat)
	if err != nil {
		t.Fatal(err)
	}

	if result.Valid() {
		t.Fatalf("Non-conformant field format incorrectly validated")
	}
}
