package datastore

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"io"
)

// Doctype representation
//
// Used to define object type on the database.
type Doctype struct {
	// ID used on the internals.
	// Because of it we can change the Code and VerboseName whenever
	// wanted.
	ID string `json:"id"`

	// It represents the name of the doctype on the JSON and the public API.
	// We can't have two doctypes with the same code.
	Code string `json:"code"`

	// The human title for the doctype.
	VerboseName string `json:"verbose_name"`

	// Fields' definitions.
	Fields map[string]*Field `json:"fields"`

	// Last revision of the doctype.
	Revision *Revision `json:"revision"`
}

// Decode implements json.Decoder
func (d *Doctype) Decode(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

// Save the doctype definition to the database.
func (d *Doctype) Save() {
	var err error
	pipeline := Conn.Pipeline()

	// Generates an ID if there's no one set
	if len(d.ID) == 0 {
		d.ID = GenerateID(4)
	}

	// create, set and Save a new Revision.
	d.Revision = CreateRevision(d.ID)
	d.Revision.Save(pipeline)

	// add this revision to a sorted set so we can retrieve all
	// the revisions on a chronological order.
	pipeline.ZAdd(joinKey([]string{d.ID, "revisions"}), redis.Z{
		Score:  float64(d.Revision.When.Unix()),
		Member: d.Revision.ID,
	})

	// set the current revision the the field's base
	// hash.
	pipeline.HSet(d.ID, "revision", d.Revision.ID)

	// make doctype be foundable by code
	// and to make codes unique
	pipeline.HSet("doctypes", d.Code, d.ID)

	pipeline.HSet(d.ID, "type", "doctype")

	// Inside this loop there's everything that should be
	// written to the history of changes (or Revision).
	// That's why I loop over the Doctype.ID and Revision.ID
	for _, baseID := range []string{d.ID, d.Revision.ID} {
		pipeline.HSet(baseID, "code", d.Code)
		pipeline.HSet(baseID, "verbose_name", d.VerboseName)
	}

	// Loop over fields to save them the the database.
	for fieldCode, field := range d.Fields {
		// Fillup missing data
		field.Code = fieldCode
		field.Revision = d.Revision

		field.Save(d, pipeline)
	}

	_, err = pipeline.Exec()
	if err != nil {
		panic(err)
	}

	pipeline.Close()
}

// LoadDoctypeByID loads a doctype's definition from the database by ID
func LoadDoctypeByID(id string) (*Doctype, error) {
	var err error

	d := &Doctype{}
	d.ID = id

	// get all basic information from base hash
	get := Conn.HGetAllMap(id).Val()

	if get["type"] != "doctype" {
		return d, fmt.Errorf("%s is type '%s', expecting 'doctype'", id, get["type"])
	}

	d.Code = get["code"]
	d.VerboseName = get["verbose_name"]
	d.Fields = make(map[string]*Field)

	// load fields ids so we can load the fields
	fieldIds := Conn.SMembers(joinKey([]string{id, "fields"})).Val()
	for _, fieldID := range fieldIds {
		LoadFieldByID(d, fieldID)
	}

	d.Revision, err = LoadRevisionByID(get["revision"])
	if err != nil {
		return d, err
	}

	return d, nil
}

// LoadDoctypeByCode loads a doctype's definition from the database by code
func LoadDoctypeByCode(code string) (*Doctype, error) {
	doctypeID := Conn.HGet("doctypes", code).Val()
	if len(doctypeID) == 0 {
		return &Doctype{}, fmt.Errorf("Could not find Doctype by code '%s'", code)
	}
	return LoadDoctypeByID(doctypeID)
}
