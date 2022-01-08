package api

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func (a *API) CreateNotification(ctx context.Context, userID, title, message, infoType, deviceToken string, infoData interface{}) error {
	// Create in database
	notification := &model.UserNotification{
		ID:        cuid.New(),
		UserID:    userID,
		Title:     title,
		Message:   message,
		InfoType:  infoType,
		InfoData:  infoData,
		CreatedAt: time.Now(),
		Read:      false,
	}
	err := a.Deps.DAL.NotificationDAL.CreateUserNotification(ctx, notification)
	if err != nil {
		return errors.Wrap(err, "unable to create notification")
	}

	// Create sns push notification
	_, err = a.Deps.AWS.SNS.SendPushNotification(deviceToken, message, title)
	if err != nil {
		return errors.Wrapf(err, "unable to send push notification")
	}
	return nil
}

func (a *API) startTransaction() (mongo.Session, error) {
	ses, err := a.Deps.DAL.Client.StartSession()
	if err != nil {
		logrus.Fatalf("[Mongo]: unable to create a session: %s", err.Error())
		return ses, errors.New("unable to start transaction")
	}
	return ses, nil
}
