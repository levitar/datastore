package datastore

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"gopkg.in/redis.v2"
	"encoding/json"
)

func TestDoctype(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network:	"tcp",
		Addr:	"127.0.0.1:6379",
	})
	pipeline := client.Pipeline()

	Convey("Create a test doctype", t, func() {
		create_doctype_json := strings.NewReader(`{
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
				},
				"tags": {
					"verbose_name": "Tags",
					"expected_types": ["string"],
					"multiple_values": true
				}
			}
		}`)

		doctype_created := Doctype{}
		err := doctype_created.Decode(create_doctype_json)
		if err != nil {
			panic(err)
		}

		doctype_created.Save(pipeline)
		_, err = pipeline.Exec()
		if err != nil {
			panic(err)
		}
		
		Convey("Load doctype from database", func() {
			doctype_loaded, doc_err := LoadDoctypeByID(doctype_created.Id, client)
			if doc_err != nil {
				panic(doc_err)
			}

			Convey("Compare doctype created with loaded", func() {
				dt_created_json, dt_created_json_err := json.MarshalIndent(doctype_created, "", "\t")
				if dt_created_json_err != nil {
					panic(dt_created_json_err)
				}

				dt_loaded_json, dt_loaded_json_err := json.MarshalIndent(doctype_loaded, "", "\t")
				if dt_loaded_json_err != nil {
					panic(dt_loaded_json_err)
				}
				
				So(dt_created_json, ShouldResemble, dt_loaded_json)
			})
		})
	})

	client.Close()
}
