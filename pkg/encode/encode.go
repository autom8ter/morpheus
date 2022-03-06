package encode

import (
	"go.mongodb.org/mongo-driver/bson"
)

func Marshal(this interface{}) ([]byte, error) {
	return bson.Marshal(this)
}

func Unmarshal(data []byte, this interface{}) error {
	return bson.Unmarshal(data, this)
}
