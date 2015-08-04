package datastore

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"io"
)

type Documenter interface {
	Slug() string
	DoctypeCode() string
}

// Document definition
//
// Holds the document's data
type Document struct {
	// Document's identification
	ID string `json:"id"`

	// It represents the slug of the document on the JSON and the public API.
	// We can't have two documents with the same slug.
	Slug string `json:"slug"`

	// The doctype of this document, so we can know it's fields
	// and everything to build, validate and store the document
	DoctypeCode string   `json:"doctype"`
	Doctype     *Doctype `json:"-"`

	Fields map[string]interface{} `json:"fields"`

	// Last revision of the document.
	Revision *Revision `json:"revision"`
}

// Decode implements json.Decoder
func (d *Document) Decode(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

// Save the doctype definition to the database.
func (d *Document) Save() {
	var err error
	pipeline := Conn.Pipeline()

	// Generates an ID if there's no one set
	if len(d.ID) == 0 {
		d.ID = GenerateID(8)
	}

	// load doctype so we can build and validate the document
	if d.Doctype == nil {
		d.Doctype, err = LoadDoctypeByCode(d.DoctypeCode)
		if err != nil {
			panic(err)
		}
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

	// make document be foundable by slug
	// and to make slugs unique
	pipeline.HSet("documents", d.Slug, d.ID)

	Conn.HSet(d.ID, "type", "document")

	// Inside this loop there's everything that should be
	// written to the history of changes (or Revision).
	// That's why I loop over the Document.ID and Revision.ID
	for _, baseID := range []string{d.ID, d.Revision.ID} {
		pipeline.HSet(baseID, "slug", d.Slug)
		pipeline.HSet(baseID, "doctype", d.Doctype.ID)
	}

	// Loop over fields to save the values to the database.
	for _, field := range d.Doctype.Fields {
		d.StoreValue(field, pipeline)
	}

	_, err = pipeline.Exec()
	if err != nil {
		panic(err)
	}

	pipeline.Close()
}

// StoreValue of the field to the database.
func (d *Document) StoreValue(f *Field, pipeline *redis.Pipeline) {
	value := d.Fields[f.Code]

	// Inside this loop there's everything that should be
	// written to the history of changes (or Revision).
	// That's why I loop over the Document.ID and Revision.ID
	for _, baseID := range []string{d.ID, d.Revision.ID} {
		baseKeyHSet := joinKey([]string{baseID, "values"})
		//baseKey := joinKey([]string{baseID, "value", f.ID})

		// choose the right Redi's type to save the value
		// and also save space on memory.
		if f.MultipleValues == false {
			fieldType := f.ExpectedTypes[0]

			if fieldType == "string" {
				pipeline.HSet(baseKeyHSet, f.ID, value.(string))
			}
		}
	}
}

// LoadValue of the field to the database.
func (d *Document) LoadValue(f *Field) {
	// get all basic information from base hash
	baseHSet := Conn.HGetAllMap(joinKey([]string{d.ID, "values"})).Val()
	//baseKey := joinKey([]string{baseID, "value", f.ID})

	// load from redis based on field's type
	if f.MultipleValues == false {
		fieldType := f.ExpectedTypes[0]

		if fieldType == "string" {
			d.Fields[f.Code] = baseHSet[f.ID]
		}
	}
}

// LoadDocumentByID loads a document from the database by ID
func LoadDocumentByID(id string) (*Document, error) {
	var err error

	d := &Document{}
	d.ID = id

	// get all basic information from base hash
	get := Conn.HGetAllMap(id).Val()

	if len(get) == 0 {
		return d, fmt.Errorf("Document NotFound for ID: %s", id)
	}

	if get["type"] != "document" {
		return d, fmt.Errorf("%s is type '%s', expecting 'document'", id, get["type"])
	}

	d.Slug = get["slug"]

	d.Doctype, err = LoadDoctypeByID(get["doctype"])
	if err != nil {
		return d, err
	}
	d.DoctypeCode = d.Doctype.Code

	d.Revision, err = LoadRevisionByID(get["revision"])
	if err != nil {
		return d, err
	}

	// load fields ids so we can load the fields
	d.Fields = make(map[string]interface{})
	for _, field := range d.Doctype.Fields {
		d.LoadValue(field)
	}

	return d, err
}

// Save a Documenter to the database
func SaveDocument(stru_doc Documenter) *Document {
	db_doc := &Document{
		Slug:        stru_doc.Slug(),
		DoctypeCode: stru_doc.DoctypeCode(),
		Fields:      FromStructToMap(stru_doc),
	}

	// save documenter to the database
	db_doc.Save()

	return db_doc
}
