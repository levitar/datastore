package datastore

import (
	"reflect"
	"strings"
)

// global function so we can easily access a doctype without needing to retrieve it from the database
var Doctypes map[string]*Doctype

// Register a Go Struct as a Doctype
func RegisterDoctype(doctype Documenter) {
	doctypeType := reflect.TypeOf(doctype)
	code := doctype.DoctypeCode()

	newDoctype := &Doctype{
		Code:        code, // could also use sstrings.ToLower(doctypeType.String()) as default value
		VerboseName: doctypeType.Name(),
		Fields:      make(map[string]*Field),
	}

	for i := 0; i < doctypeType.NumField(); i++ {
		typeField := doctypeType.Field(i)
		tags := typeField.Tag

		fieldTags := strings.Split(tags.Get("field"), ",")
		fieldName := fieldTags[0]
		//fieldFlags := fieldTags[1:]

		// if there's no struct tag defining it's name use the struct's field name as default
		if fieldName == "" {
			fieldName = typeField.Name
		}

		// set field to the new doctype we're building
		newDoctype.Fields[fieldName] = &Field{
			Code:          fieldName,
			VerboseName:   typeField.Name,
			ExpectedTypes: []string{typeField.Type.String()},
		}
	}

	Doctypes[code] = newDoctype

	// save new doctype to the database
	newDoctype.Save()
}

func init() {
	Doctypes = make(map[string]*Doctype)
}
