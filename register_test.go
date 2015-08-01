package datastore

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type User struct {
	Username    string `json:"username" field:"username,unique"`
	Name        string `json:"name" field:"name"`
	WithoutName string `field:",unique"`
}

func (u User) Slug() string {
	return fmt.Sprint(u.DoctypeCode(), "/", u.Username)
}

func (u User) DoctypeCode() string {
	return "user"
}

func TestRegisterDocumenter(t *testing.T) {
	Convey("Registering Doctype", t, func() {
		RegisterDoctype(User{})

		Convey("Save a Struct to the Database", func() {
			u := User{
				Username:    "alisson",
				Name:        "Alisson Patricio",
				WithoutName: "Any fucking shit goes here",
			}

			Id := SaveDocument(u)

			documentLoaded, documentLoadedErr := LoadDocumentByID(Id)
			if documentLoadedErr != nil {
				panic(documentLoadedErr)
			}

			fmt.Println(documentLoaded)
		})
	})
}
