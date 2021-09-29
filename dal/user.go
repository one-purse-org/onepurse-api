package dal

import (
	"context"
	"errors"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IUserDAL interface {
	Add(user *model.User) error
	FindByID(userID string) (*model.User, error)
	FindAll() (*[]model.User, error)
	FindByUsername(username string) (*model.User, error)
	UpdateUser(userID string, updateParam bson.D) error
	DeleteUser(userID string) error
}

type UserDAL struct {
	DB         *mongo.Database
	Collection *mongo.Collection
}

func NewUserDAL(db *mongo.Database) *UserDAL {
	return &UserDAL{
		DB:         db,
		Collection: db.Collection("user"),
	}
}

func (u UserDAL) Add(user *model.User) error {
	_, err := u.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user record already exists")
		} else {
			return err
		}
	}
	return nil
}

func (u UserDAL) FindByID(userID string) (*model.User, error) {
	var user *model.User
	err := u.Collection.FindOne(context.TODO(), bson.D{{"_id", userID}}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for user %s not found", userID)
			return nil, errors.New(findErr)
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (u UserDAL) FindAll() (*[]model.User, error) {
	var user *[]model.User

	cursor, err := u.Collection.Find(context.TODO(), bson.D{})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &[]model.User{}, nil // TODO(josiah): confirm that this logic implements what you have in mind
		} else {
			logrus.Fatalf("[Mongo]: error fetching users : %s", err.Error())
			return nil, err
		}
	}

	if err = cursor.All(context.TODO(), &user); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to user model : %s", err.Error())
		return nil, err
	}

	return user, nil
}

func (u UserDAL) FindByUsername(username string) (*model.User, error) {
	var user *model.User

	err := u.Collection.FindOne(context.TODO(), bson.D{{"user_name", username}}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("user %s record does not exist", username)
			return nil, errors.New(findErr)
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (u UserDAL) UpdateUser(userID string, updateParam bson.D) error {
	result, err := u.Collection.UpdateByID(context.TODO(), userID, updateParam)
	if err != nil {
		logrus.Fatalf("[Mongo]: error updating user %s : %s", userID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating user %s : user record not found", userID)
		return errors.New("user record not found")
	}
	return nil
}

func (u UserDAL) DeleteUser(userID string) error {
	result, err := u.Collection.DeleteOne(context.TODO(), bson.D{{"_id", userID}})
	if err != nil {
		logrus.Fatalf("error deleting user %s : %s", userID, err)
		return err
	}

	if result.DeletedCount == 0 {
		logrus.Fatalf("error deleting user %s : user record does not exist", userID)
		return errors.New("user record does not exist")
	}
	return nil
}