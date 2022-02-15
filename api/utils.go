package api

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"time"
)

//CreateNotification creates a user notification entry in the database and fires an AWS SNS message to the device endpoint
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

// GetNumberMetrics fetches all the information required for the NumberMetrics struct
func (a *API) GetNumberMetrics(ctx context.Context) (*model.NumberMetrics, error) {
	numUser, err := a.Deps.DAL.UserDAL.Count(ctx)
	if err != nil {
		return nil, err
	}
	numAgent, err := a.Deps.DAL.AgentDAL.Count(ctx)
	if err != nil {
		return nil, err
	}
	numTransaction, err := a.Deps.DAL.TransactionDAL.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	metrics := model.NumberMetrics{
		NumberOfUser:        numUser,
		NumberOfAgent:       numAgent,
		NumberOfTransaction: numTransaction,
	}
	return &metrics, nil
}

// GetCurrencyMetrics fetches all the information required for the CurrencyMetrics struct
func (a *API) GetCurrencyMetrics(ctx context.Context) (*model.CurrencyMetrics, error) {
	return nil, nil
}

// TransactionVolumeMetrics fetches all the information required for the transactionMetrics struct
func (a *API) TransactionVolumeMetrics(ctx context.Context, start, end time.Time) (*model.TransactionVolumeMetrics, error) {
	return nil, nil
}
