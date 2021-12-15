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

type ITransactionDAL interface {
	CreateTransfer(transfer *model.Transfer) error
	CreateWithdrawal(withdrawal *model.Withdrawal) error
	GetTransferByID(transferID string) (*model.Transfer, error)
	GetWithdrawalByID(withdrawalID string) (*model.Withdrawal, error)
	UpdateTransfer(transferID string, updateParam bson.D) error
	UpdateWithdrawal(withdrawalID string, updateParam bson.D) error
}

type TransactionDAL struct {
	DB                   *mongo.Database
	TransferCollection   *mongo.Collection
	WithdrawalCollection *mongo.Collection
}

func NewTransactionDAL(db *mongo.Database) *TransactionDAL {
	return &TransactionDAL{
		DB:                   db,
		TransferCollection:   db.Collection("transfer"),
		WithdrawalCollection: db.Collection("withdraw"),
	}
}

func (t TransactionDAL) CreateTransfer(transfer *model.Transfer) error {
	_, err := t.TransferCollection.InsertOne(context.TODO(), transfer)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("Transfer record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateWithdrawal(withdrawal *model.Withdrawal) error {
	_, err := t.WithdrawalCollection.InsertOne(context.TODO(), withdrawal)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("withdrawal record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) GetTransferByID(transferID string) (*model.Transfer, error) {
	var transfer *model.Transfer
	err := t.TransferCollection.FindOne(context.TODO(), bson.D{{"_id", transferID}}).Decode(&transfer)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for trasfer %s not found", transferID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return transfer, nil
}

func (t TransactionDAL) GetWithdrawalByID(withdrawalID string) (*model.Withdrawal, error) {
	var withdrawal *model.Withdrawal
	err := t.WithdrawalCollection.FindOne(context.TODO(), bson.D{{"_id", withdrawalID}}).Decode(&withdrawal)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for withdrawal %s not found", withdrawalID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return withdrawal, nil
}

func (t TransactionDAL) UpdateTransfer(transferID string, updateParam bson.D) error {
	result, err := t.TransferCollection.UpdateByID(context.TODO(), transferID, updateParam)
	if err != nil {
		logrus.Fatalf("[Mongo]: error updating transfer %s: %s", transferID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating transfer %s : transfer record not found", transferID)
		return errors.New("transfer record not found")
	}
	return nil
}

func (t TransactionDAL) UpdateWithdrawal(withdrawalID string, updateParam bson.D) error {
	result, err := t.WithdrawalCollection.UpdateByID(context.TODO(), withdrawalID, updateParam)
	if err != nil {
		logrus.Fatalf("[Mongo]: error updating withdrawal %s: %s", withdrawalID, err.Error())
		return err
	}

	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating withdrawal %s : user record not found", withdrawalID)
		return errors.New("withdraw record not found")
	}

	return nil
}
