package datastore

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type User struct {
	ID          string `json"id"`
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

func (u *User) GetID() string {
	return u.ID
}

func (u *User) SetID(id string) {
	u.ID = id
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

			SaveDocument(user)

			documentLoaded, documentLoadedErr := LoadDocumentByID(user.GetID())
			if documentLoadedErr != nil {
				panic(documentLoadedErr)
			}

			fmt.Println(documentLoaded)
		})
	})
}
