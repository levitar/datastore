package datastore

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestDoctype(t *testing.T) {
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

		doctypeCreated.Save()

		Convey("Load doctype from database", func() {
			doctypeLoaded, docErr := LoadDoctypeByID(doctypeCreated.ID)
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
}
