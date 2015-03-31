package datastore

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v2"
	"io"
)

// Doctype representation
//
// Used to define object type on the database.
type Doctype struct {
	// ID used on the internals.
	// Because of it we can change the Code and VerboseName whenever
	// wanted.
	Id string `json:"id"`

	// It represents the name of the field on the JSON and the public API.
	// We can't have two doctypes with the same code.
	Code string `json:"code"`

	// The human title for the doctype.
	VerboseName string `json:"verbose_name"`

	// Fields' definition.
	Fields map[string]*Field `json:"fields"`

	// Last revision of the doctype.
	Revision *Revision `json:"revision"`
}

// Implmements json.Decoder
func (d *Doctype) Decode(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

// Sabe the doctype definition to the database.
func (d *Doctype) Save(client *redis.Pipeline) {
	// Generates an ID if there's no one set
	if len(d.Id) == 0 {
		d.Id = GenerateID(6)
	}

	// create, set and Save a new Revision.
	d.Revision = CreateRevision(d.Id)
	d.Revision.Save(client)

	// add this revision to a sorted set so we can retrieve all
	// the revisions on a chronological order.
	client.ZAdd(joinKey([]string{d.Id, "revisions"}), redis.Z{
		Score:  float64(d.Revision.When.Unix()),
		Member: d.Revision.Id,
	})

	// set the current revision the the field's base
	// hash.
	client.HSet(d.Id, "revision", d.Revision.Id)

	// make doctype be foundable by code
	// and to make codes unique
	client.HSet("doctypes", d.Code, d.Id)

	client.HSet(d.Id, "type", "doctype")

	// Inside this loop there's everything that should be
	// written to the history of changes (or Revision).
	// That's why I loop over the Doctype.Id and Revision.Id
	for _, base_id := range []string{d.Id, d.Revision.Id} {
		client.HSet(base_id, "code", d.Code)
		client.HSet(base_id, "verbose_name", d.VerboseName)
	}

	// Loop over fields to save them the the database.
	for field_code, field := range d.Fields {
		// Fillup missing data
		field.Code = field_code
		field.Revision = d.Revision

		field.Save(d, client)
	}
}

// Load a doctype's definition from the database by ID
func LoadDoctypeByID(id string, client *redis.Client) (*Doctype, error) {
	var err error

	d := &Doctype{}
	d.Id = id

	// get all basic information from base hash
	get := client.HGetAllMap(id).Val()

	if get["type"] != "doctype" {
		return d, fmt.Errorf("%s is type '%s', expecting 'doctype'", id, get["type"])
	}

	d.Code = get["code"]
	d.VerboseName = get["verbose_name"]
	d.Fields = make(map[string]*Field)

	// load fields ids so we can load the fields
	field_ids := client.SMembers(joinKey([]string{id, "fields"})).Val()
	for _, field_id := range field_ids {
		LoadFieldByID(d, field_id, client)
	}

	d.Revision, err = LoadRevisionByID(get["revision"], client)
	if err != nil {
		return d, err
	}

	return d, nil
}
