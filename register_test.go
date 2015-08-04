package datastore

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type User struct {
	Username    string `json:"username" field:"username,unique"`
	Name        string `json:"name" field:"name"`
	WithoutName string `field:",unique"`
}

func (u *User) Slug() string {
	return fmt.Sprint(u.DoctypeCode(), "/", u.Username)
}

func (u *User) DoctypeCode() string {
	return "user"
}

func TestRegisterDocumenter(t *testing.T) {
	Convey("Registering Doctype", t, func() {
		user := &User{}
		RegisterDoctype(user)

		_, has_user_doctype := Doctypes[user.DoctypeCode()]
		So(has_user_doctype, ShouldBeTrue)

		Convey("Save a document instance to the Database", func() {
			user.Username = "alisson"
			user.Name = "Alisson Patricio"
			user.WithoutName = "Alisson Patricio"

			documentCreated := SaveDocument(user)

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
}
