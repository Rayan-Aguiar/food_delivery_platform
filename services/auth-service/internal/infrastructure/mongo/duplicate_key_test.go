package mongo

import (
	"errors"
	"testing"

	mdriver "go.mongodb.org/mongo-driver/mongo"
)

func TestIsDuplicateKeyError_NilReturnsFalse(t *testing.T) {
	if isDuplicateKeyError(nil) {
		t.Error("isDuplicateKeyError(nil) deve retornar false")
	}
}

func TestIsDuplicateKeyError_GenericErrorReturnsFalse(t *testing.T) {
	if isDuplicateKeyError(errors.New("algum outro erro")) {
		t.Error("erro genérico não deve ser identificado como duplicate key")
	}
}

func TestIsDuplicateKeyError_DuplicateKeyReturnsTrue(t *testing.T) {
	// mongo.WriteException com código 11000 = E11000 duplicate key
	err := mdriver.WriteException{
		WriteErrors: mdriver.WriteErrors{
			{Code: 11000, Message: "E11000 duplicate key error collection: auth.credentials index: email_1"},
		},
	}
	if !isDuplicateKeyError(err) {
		t.Error("WriteException com code 11000 deve retornar true")
	}
}

func TestIsDuplicateKeyError_OtherWriteExceptionReturnsFalse(t *testing.T) {
	// WriteException com código diferente de 11000
	err := mdriver.WriteException{
		WriteErrors: mdriver.WriteErrors{
			{Code: 121, Message: "document failed validation"},
		},
	}
	if isDuplicateKeyError(err) {
		t.Error("WriteException com code != 11000 não deve retornar true")
	}
}
