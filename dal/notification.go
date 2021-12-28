package dal

import (
	"context"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type INotificationDAL interface {
	CreateUserNotification(ctx context.Context, notification *model.UserNotification) error
	FetchUserNotification(ctx context.Context, query bson.D) (*model.UserNotification, error)
	FetchUserNotifications(ctx context.Context, query bson.D) (*[]model.UserNotification, error)
	UpdateUserNotification(ctx context.Context, ID string, query bson.D) error
	DeleteUserNotification(ctx context.Context, ID string) error
}

type NotificationDAL struct {
	DB                         *mongo.Database
	UserNotificationCollection *mongo.Collection
}

func NewNotificationDAL(db *mongo.Database) *NotificationDAL {
	return &NotificationDAL{
		DB:                         db,
		UserNotificationCollection: db.Collection("user-notification"),
	}
}

func (n NotificationDAL) CreateUserNotification(ctx context.Context, notification *model.UserNotification) error {
	_, err := n.UserNotificationCollection.InsertOne(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

func (n NotificationDAL) FetchUserNotification(ctx context.Context, query bson.D) (*model.UserNotification, error) {
	var notification model.UserNotification
	err := n.UserNotificationCollection.FindOne(ctx, query).Decode(&notification)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for notification not found")
			return nil, errors.New(findErr)
		}
		return nil, err
	}

	return &notification, nil
}

func (n NotificationDAL) FetchUserNotifications(ctx context.Context, query bson.D) (*[]model.UserNotification, error) {
	var notifications []model.UserNotification
	cursor, err := n.UserNotificationCollection.Find(ctx, query)
	if err != nil {
		logrus.Errorf("[Mongo]: error fetching notifications: %s", err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &notifications); err != nil {
		logrus.Errorf("[Mongo]: error decoding notification results: %s", err.Error())
		return nil, err
	}
	return &notifications, nil
}

func (n NotificationDAL) UpdateUserNotification(ctx context.Context, ID string, updateParam bson.D) error {
	result, err := n.UserNotificationCollection.UpdateByID(ctx, ID, updateParam)
	if err != nil {
		logrus.Fatalf("[Mongo]: error updating notification %s: %s", ID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating notification %s: record not found", ID)
		return errors.New("notification record not found")
	}
	return nil
}

func (n NotificationDAL) DeleteUserNotification(ctx context.Context, ID string) error {
	result, err := n.UserNotificationCollection.DeleteOne(ctx, bson.D{{"_id", ID}})
	if err != nil {
		logrus.Fatalf("[Mongo]: error deleting user notification %s: %s", ID, err.Error())
		return err
	}

	if result.DeletedCount == 0 {
		logrus.Fatalf("[Mongo]: error deleting notification %s: notification not found", ID)
		return errors.New("notification record not found")
	}
	return nil
}
