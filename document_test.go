package datastore

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestDocument(t *testing.T) {
	Convey("Create a doctype", t, func() {
		createDoctypeJSON := strings.NewReader(`{
			"code": "page",
			"verbose_name": "Pagina",
			"fields": {
				"title": {
					"verbose_name": "Titulo",
					"expected_types": ["string"]
				},
				"body": {
					"verbose_name": "Texto",
					"expected_types": ["string"]
				}
			}
		}`)

		doctypeCreated := Doctype{}
		err := doctypeCreated.Decode(createDoctypeJSON)
		if err != nil {
			panic(err)
		}

		doctypeCreated.Save()

		Convey("Save a document", func() {
			createDocumentJSON := strings.NewReader(`{
				"slug": "my-first-page",
				"doctype": "page",
				"fields": {
					"title": "My First Page",
					"body": "Here comes by body"
				}
			}`)

			documentCreated := Document{}
			err := documentCreated.Decode(createDocumentJSON)
			if err != nil {
				panic(err)
			}

			documentCreated.Save()

			Convey("Load document from database", func() {
				documentLoaded, documentLoadedErr := LoadDocumentByID(documentCreated.ID)
				if documentLoadedErr != nil {
					panic(documentLoadedErr)
				}

				Convey("Compare document created with loaded", func() {
					docCreatedJSON, docCreatedJSONErr := json.MarshalIndent(documentCreated, "", "\t")
					if docCreatedJSONErr != nil {
						panic(docCreatedJSONErr)
					}

					docLoadedJSON, docLoadedJSONErr := json.MarshalIndent(documentLoaded, "", "\t")
					if docLoadedJSONErr != nil {
						panic(docLoadedJSONErr)
					}

					So(docCreatedJSON, ShouldResemble, docLoadedJSON)
				})
			})
		})
	})

	Convey("Document not found", t, func() {
		So(func() {
			_, err := LoadDocumentByID("RandomID1231")
			if err != nil {
				panic(err)
			}
		}, ShouldPanic)
	})
}
