package datastore

import (
	"fmt"
	"gopkg.in/redis.v3"
	"time"
)

// It should be clear if the change is a Add, Remove or Change (patch)

// Revision meta data
//
// Contains basic information for the revision
type Revision struct {
	// Revision's identification
	ID string `json:"id"`

	// Message summarizing the revision.
	Message string `json:"message,omitempty"`

	// Time and data of the modification
	When time.Time `json:"when"`

	// Kind of changed
	Type string `json:"type"`

	// Parent revision
	Parent string `json:"parent"`

	// Object's ID on the database
	Object string `json:"-"`
}

// Save revision to the database.
func (r *Revision) Save(client *redis.Pipeline) {
	client.ZAdd("revisions", redis.Z{
		Score:  float64(r.When.Unix()),
		Member: r.ID,
	})

	client.HSet(r.ID, "type", "revision")
	client.HSet(r.ID, "object", r.Object)
	client.HSet(r.ID, "when", r.When.Format(time.RFC3339Nano))
	client.HSet(r.ID, "change_type", r.Type)
	client.HSet(r.ID, "parent", r.Parent)
}

// CreateRevision creates a revision meta to the object
func CreateRevision(objectID string) *Revision {
	revision := &Revision{}
	revision.ID = GenerateID(9)
	revision.When = time.Now().UTC()
	revision.Type = "create"
	revision.Object = objectID

	return revision
}

// UpdateRevision creates a update revision
func UpdateRevision(parent *Revision) *Revision {
	revision := &Revision{}
	revision.ID = GenerateID(9)
	revision.When = time.Now().UTC()
	revision.Type = "update"
	revision.Object = parent.Object
	revision.Parent = parent.Parent

	return revision
}

// LoadRevisionByID loads a revision meta data from the database by ID.
func LoadRevisionByID(id string) (*Revision, error) {
	var err error

	r := &Revision{}

	// get all basic information from base hash
	get := Conn.HGetAllMap(id).Val()

	if get["type"] != "revision" {
		return r, fmt.Errorf("%s is type '%s', expecting 'revision'", id, get["type"])
	}

	r.ID = id
	r.Type = get["change_type"]
	r.Object = get["object"]
	r.Message = get["message"]
	r.Parent = get["parent"]

	r.When, err = time.Parse(time.RFC3339Nano, get["when"])
	if err != nil {
		return r, err
	}

	return r, nil
}
