package datastore

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/redis.v2"
	"strings"
	"testing"
)

func TestDoctype(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	pipeline := client.Pipeline()

	Convey("Create a test doctype", t, func() {
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
				},
				"tags": {
					"verbose_name": "Tags",
					"expected_types": ["string"],
					"multiple_values": true
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

		Convey("Load doctype from database", func() {
			doctypeLoaded, docErr := LoadDoctypeByID(doctypeCreated.ID, client)
			if docErr != nil {
				panic(docErr)
			}

			Convey("Compare doctype created with loaded", func() {
				dtCreatedJSON, dtCreatedJSONErr := json.MarshalIndent(doctypeCreated, "", "\t")
				if dtCreatedJSONErr != nil {
					panic(dtCreatedJSONErr)
				}

				dtLoadedJSON, dtLoadedJSONErr := json.MarshalIndent(doctypeLoaded, "", "\t")
				if dtLoadedJSONErr != nil {
					panic(dtLoadedJSONErr)
				}

				So(dtCreatedJSON, ShouldResemble, dtLoadedJSON)
			})
		})
	})

	client.Close()
}
