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
	Add(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, userID string) (*model.User, error)
	FindAll(ctx context.Context) (*[]model.User, error)
	FindOne(ctx context.Context, query bson.D) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, userID string, updateParam bson.D) error
	DeleteUser(ctx context.Context, userID string) error
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

func (u UserDAL) Add(ctx context.Context, user *model.User) error {
	_, err := u.Collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user record already exists")
		}
		return err
	}
	return nil
}

func (u UserDAL) FindByID(ctx context.Context, userID string) (*model.User, error) {
	var user *model.User
	err := u.Collection.FindOne(ctx, bson.D{{"_id", userID}}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for user %s not found", userID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return user, nil
}

func (u UserDAL) FindAll(ctx context.Context) (*[]model.User, error) {
	var users []model.User

	cursor, err := u.Collection.Find(ctx, bson.D{})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &[]model.User{}, nil // TODO(josiah): confirm that this logic implements what you have in mind
		}
		logrus.Fatalf("[Mongo]: error fetching users : %s", err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &users); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to users model : %s", err.Error())
		return nil, err
	}

	return &users, nil
}

func (u UserDAL) FindOne(ctx context.Context, query bson.D) (*model.User, error) {
	var user *model.User
	if ctx == nil {
		ctx = context.TODO()
	}
	err := u.Collection.FindOne(ctx, query).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("no user matches the query")
			return nil, errors.New(findErr)
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (u UserDAL) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user *model.User

	err := u.Collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": username},
		},
	}).Decode(&user)

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

func (u UserDAL) UpdateUser(ctx context.Context, userID string, updateParam bson.D) error {
	result, err := u.Collection.UpdateByID(ctx, userID, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error updating user %s : %s", userID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Errorf("[Mongo]: error updating user %s : user record not found", userID)
		return errors.New("user record not found")
	}
	return nil
}

func (u UserDAL) DeleteUser(ctx context.Context, userID string) error {
	result, err := u.Collection.DeleteOne(ctx, bson.D{{"_id", userID}})
	if err != nil {
		logrus.Errorf("error deleting user %s : %s", userID, err)
		return err
	}

	if result.DeletedCount == 0 {
		logrus.Errorf("error deleting user %s : user record does not exist", userID)
		return errors.New("user record does not exist")
	}
	return nil
}
