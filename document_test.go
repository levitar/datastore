package datastore

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/redis.v2"
	"strings"
	"testing"
)

func TestDocument(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	pipeline := client.Pipeline()

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

		doctypeCreated.Save(pipeline)
		_, err = pipeline.Exec()
		if err != nil {
			panic(err)
		}

		pipeline.Close()

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

			documentCreated.Save(client)

			Convey("Load document from database", func() {
				documentLoaded, documentLoadedErr := LoadDocumentByID(documentCreated.ID, client)
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

	client.Close()
}
