package datastore

import (
	"gopkg.in/redis.v2"
	"strconv"
)

// Field represents of a field of a Doctype
type Field struct {
	// ID used on the internals of the database.
	// Because of it we can change the Code and VerboseName whenever
	// wanted.
	Id string `json:"id"`

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

// Saves the field definition to the database.
func (f *Field) Save(doctype *Doctype, client *redis.Pipeline) {
	// Generates an ID if there's no one set
	if len(f.Id) == 0 {
		f.Id = GenerateID(2)
	}

	// base key used on the database, it should prepend anything
	// from this field on the database.
	base_key := joinKey([]string{doctype.Id, "field", f.Id})

	// add this revision to a sorted set so we can retrieve all
	// the revisions on a chronological order.
	client.ZAdd(joinKey([]string{base_key, "revisions"}), redis.Z{
		Score:  float64(f.Revision.When.Unix()),
		Member: f.Revision.Id,
	})

	// set the current revision the the field's base
	// hash.
	client.HSet(base_key, "revision", f.Revision.Id)

	// Inside this loop there's everything that should be
	// written to the history of changes (or Revision).
	// That's why I loop over the Doctype.Id and Revision.Id
	for _, base_id := range []string{doctype.Id, f.Revision.Id} {
		// change base key to use the Revision.Id when necessary.
		base_key = joinKey([]string{base_id, "field", f.Id})

		// Add fields to doctype's (and revision's) fields set
		// it's necessary so the doctype (and revision) can retrieve
		// all the fields in it's definition.
		client.SAdd(joinKey([]string{base_id, "fields"}), f.Id)

		client.HSet(base_key, "verbose_name", f.VerboseName)
		client.HSet(base_key, "code", f.Code)
		client.HSet(base_key, "multiple_values", strconv.FormatBool(f.MultipleValues))

		for _, expected_type := range f.ExpectedTypes {
			client.SAdd(joinKey([]string{base_key, "expected_types"}), expected_type)
		}
	}
}

// Load a doctype's field's definition from the database by ID
func LoadFieldByID(d *Doctype, id string, client *redis.Client) {
	var err error

	f := &Field{}
	f.Id = id

	// make base field's key
	base_key := joinKey([]string{d.Id, "field", f.Id})

	// get all basic information from base hash
	get := client.HGetAllMap(base_key).Val()

	f.Code = get["code"]
	f.VerboseName = get["verbose_name"]

	f.MultipleValues, err = strconv.ParseBool(get["multiple_values"])
	if err != nil {
		panic(err)
	}

	f.Revision, err = LoadRevisionByID(get["revision"], client)
	if err != nil {
		panic(err)
	}

	f.ExpectedTypes = client.SMembers(joinKey([]string{base_key, "expected_types"})).Val()

	// add field to doctype's instance fields definitions
	d.Fields[f.Code] = f
}
