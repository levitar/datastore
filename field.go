package datastore

import (
	"gopkg.in/redis.v3"
	"strconv"
)

// Field represents of a field of a Doctype
type Field struct {
	// ID used on the internals of the database.
	// Because of it we can change the Code and VerboseName whenever
	// wanted.
	ID string `json:"id"`

	// It represents the name of the field on the JSON and the public API.
	// No Doctype can have two fields with the same code.
	Code string `json:"code"`

	// The human title for the field.
	VerboseName string `json:"verbose_name"`

	// Types of values expected on the field.
	// It accepts multiple types so we can have multiple doctypes
	// referenced on the values.
	ExpectedTypes []string `json:"expected_types"` // needs to be a list of types

	// Flag to set the field to store multiple values instead of just one.
	// Usefull for things like tags or categories.
	MultipleValues bool `json:"multiple_values"`

	// Last revision of the field.
	Revision *Revision `json:"revision"`
}

// Save the field definition to the database.
func (f *Field) Save(doctype *Doctype, pipeline *redis.Pipeline) {
	// Generates an ID if there's no one set
	if len(f.ID) == 0 {
		f.ID = GenerateID(2)
	}

	// base key used on the database, it should prepend anything
	// from this field on the database.
	baseKey := joinKey([]string{doctype.ID, "field", f.ID})

	// add this revision to a sorted set so we can retrieve all
	// the revisions on a chronological order.
	pipeline.ZAdd(joinKey([]string{baseKey, "revisions"}), redis.Z{
		Score:  float64(f.Revision.When.Unix()),
		Member: f.Revision.ID,
	})

	// set the current revision the the field's base
	// hash.
	pipeline.HSet(baseKey, "revision", f.Revision.ID)

	// Inside this loop there's everything that should be
	// written to the history of changes (or Revision).
	// That's why I loop over the Doctype.ID and Revision.ID
	for _, baseID := range []string{doctype.ID, f.Revision.ID} {
		// change base key to use the Revision.ID when necessary.
		baseKey = joinKey([]string{baseID, "field", f.ID})

		// Add fields to doctype's (and revision's) fields set
		// it's necessary so the doctype (and revision) can retrieve
		// all the fields in it's definition.
		pipeline.SAdd(joinKey([]string{baseID, "fields"}), f.ID)

		pipeline.HSet(baseKey, "verbose_name", f.VerboseName)
		pipeline.HSet(baseKey, "code", f.Code)
		pipeline.HSet(baseKey, "multiple_values", strconv.FormatBool(f.MultipleValues))

		for _, expectedType := range f.ExpectedTypes {
			pipeline.SAdd(joinKey([]string{baseKey, "expected_types"}), expectedType)
		}
	}
}

// LoadFieldByID loads a doctype's field's definition from the database by ID
func LoadFieldByID(d *Doctype, id string) {
	var err error

	f := &Field{}
	f.ID = id

	// make base field's key
	baseKey := joinKey([]string{d.ID, "field", f.ID})

	// get all basic information from base hash
	get := Conn.HGetAllMap(baseKey).Val()

	f.Code = get["code"]
	f.VerboseName = get["verbose_name"]

	f.MultipleValues, err = strconv.ParseBool(get["multiple_values"])
	if err != nil {
		panic(err)
	}

	f.Revision, err = LoadRevisionByID(get["revision"])
	if err != nil {
		panic(err)
	}

	f.ExpectedTypes = Conn.SMembers(joinKey([]string{baseKey, "expected_types"})).Val()

	// add field to doctype's instance fields definitions
	d.Fields[f.Code] = f
}
