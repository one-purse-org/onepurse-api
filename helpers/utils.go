package helpers

import (
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//DoSufficientFundsCheck checks that the user has sufficient funds in wallet to carry out the transaction
func DoSufficientFundsCheck(user *model.User, amount float32, currency string) bool {
	if user.Wallet[currency].AvailableBalance < amount {
		return false
	}
	return true
}

func MarshalStructToBSONDoc(structure interface{}) (bson.D, error) {
	var doc bson.D

	val, err := bson.Marshal(structure)
	if err != nil {
		return nil, err
	}

	err = bson.Unmarshal(val, &doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
