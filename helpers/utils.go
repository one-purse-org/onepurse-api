package helpers

import (
	"encoding/base32"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"time"
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

//MarshalStructToBSONDoc marshals a struct to a mongo document
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

//CreateOTP creates an otp code using the userID as secret
func CreateOTP(user *model.User) (*otp.Key, error) {
	token, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "onepurse",
		AccountName: "hello@onepurse.co",
		Period:      300,
		SecretSize:  0,
		Secret:      []byte(user.ID),
		Digits:      0,
		Algorithm:   0,
		Rand:        nil,
	})
	if err != nil {
		logrus.Errorf("[OTP]: error generating otp: %s", err.Error())
		return nil, err
	}
	return token, err
}

//CreateOTPCode creates an otp code using the userID as secret
func CreateOTPCode(userID string) (string, error) {
	s := generateSecret(userID)

	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA512,
	}
	passcode, err := totp.GenerateCodeCustom(s, time.Now(), opts)
	if err != nil {
		logrus.Errorf("[OTP]: error generating otp code: %s", err.Error())
		return "", err
	}
	return passcode, err
}

//ValidateOTPCode validates a passcode using the userID as secret
func ValidateOTPCode(userID string, passcode string) bool {
	s := generateSecret(userID)

	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA512,
	}
	valid, err := totp.ValidateCustom(passcode, s, time.Now(), opts)
	if err != nil {
		logrus.Errorf("[OTP]: unable to validate otp: %s", err.Error())
		return false
	}

	return valid
}

//generateSecret generates secret for generating and validating OTP
func generateSecret(key string) string {
	secret := []byte(key)
	s := base32.StdEncoding.EncodeToString(secret)
	return s
}

//DoUserWalletCheck checks to see if a user has activated the specified wallet
func DoUserWalletCheck(user *model.User, walletType string) bool {
	if _, found := user.Wallet[walletType]; found {
		return true
	}
	return false
}
