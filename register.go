package datastore

import (
	"reflect"
	"strings"
)

// global function so we can easily access a doctype without needing to retrieve it from the database
var Doctypes = make(map[string]*Doctype)

// Register a Go Struct as a Doctype
func RegisterDoctype(doctype Documenter) {
	doctypeType := reflect.TypeOf(doctype)
	code := doctype.DoctypeCode()

	newDoctype := &Doctype{
		Code:        code, // could also use sstrings.ToLower(doctypeType.String()) as default value
		VerboseName: doctypeType.Name(),
		Fields:      make(map[string]*Field),
	}

	// is it a pointer? if so, get the struct
	if doctypeType.Kind() == reflect.Ptr {
		doctypeType = doctypeType.Elem()
	}

	for i := 0; i < doctypeType.NumField(); i++ {
		typeField := doctypeType.Field(i)
		tags := typeField.Tag

		jsonTags := strings.Split(tags.Get("json"), ",")
		jsonName := jsonTags[0]

		// if name is - then ignore it
		if jsonName == "-" {
			continue
		}

		// if there's no struct tag defining it's name use the struct's field name as default
		if jsonName == "" {
			jsonName = typeField.Name
		}

		// set field to the new doctype we're building
		newDoctype.Fields[jsonName] = &Field{
			Code:          jsonName,
			VerboseName:   typeField.Name,
			ExpectedTypes: []string{typeField.Type.String()},
		}
	}

	Doctypes[code] = newDoctype

	// save new doctype to the database
	newDoctype.Save()
}
