package mongo

import mdriver "go.mongodb.org/mongo-driver/mongo"


func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	if mdriver.IsDuplicateKeyError(err) {
		return true
	}

	return false
}