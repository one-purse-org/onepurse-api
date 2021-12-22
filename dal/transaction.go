package dal

import (
	"context"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ITransactionDAL interface {
	CreateTransfer(transfer *model.Transfer) error
	CreateWithdrawal(withdrawal *model.Withdrawal) error
	CreateDeposit(deposit *model.Deposit) error
	GetTransferByID(transferID string) (*model.Transfer, error)
	GetWithdrawalByID(withdrawalID string) (*model.Withdrawal, error)
	GetDepositByID(depositID string) (*model.Deposit, error)
	UpdateTransfer(transferID string, updateParam bson.D) error
	UpdateWithdrawal(withdrawalID string, updateParam bson.D) error
	UpdateDeposit(depositID string, updateParam bson.D) error
	CheckTimeLimit() error
}

type TransactionDAL struct {
	DB                   *mongo.Database
	TransferCollection   *mongo.Collection
	WithdrawalCollection *mongo.Collection
	DepositCollection    *mongo.Collection
	ExchangeCollection   *mongo.Collection
}

func NewTransactionDAL(db *mongo.Database) *TransactionDAL {
	return &TransactionDAL{
		DB:                   db,
		TransferCollection:   db.Collection("transfer"),
		WithdrawalCollection: db.Collection("withdraw"),
		DepositCollection:    db.Collection("deposit"),
		ExchangeCollection:   db.Collection("exchange"),
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

func (t TransactionDAL) CreateDeposit(deposit *model.Deposit) error {
	_, err := t.DepositCollection.InsertOne(context.TODO(), deposit)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("deposit record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) GetTransferByID(transferID string) (*model.Transfer, error) {
	var transfer *model.Transfer
	err := t.TransferCollection.FindOne(context.TODO(), bson.D{{"_id", transferID}}).Decode(transfer)

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
	err := t.WithdrawalCollection.FindOne(context.TODO(), bson.D{{"_id", withdrawalID}}).Decode(withdrawal)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for withdrawal %s not found", withdrawalID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return withdrawal, nil
}

func (t TransactionDAL) GetDepositByID(depositID string) (*model.Deposit, error) {
	var deposit *model.Deposit
	err := t.DepositCollection.FindOne(context.TODO(), bson.D{{"_id", depositID}}).Decode(deposit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for deposit %s not found", depositID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return deposit, nil
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

func (t TransactionDAL) UpdateDeposit(depositID string, updatePram bson.D) error {
	result, err := t.DepositCollection.UpdateByID(context.TODO(), depositID, updatePram)
	if err != nil {
		logrus.Fatalf("[Mongo]: error updating deposit %s: %s", depositID, err.Error())
		return err
	}

	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating deposit %s : user record not found", depositID)
		return errors.New("deposit record not found")
	}

	return nil
}

func (t TransactionDAL) CheckTimeLimit() error {
	fmt.Println("here one")
	logrus.Info("Running Transaction Time Limit Cron")
	// first fetch all transactions created today
	var transfers *[]model.Transfer
	fmt.Println("here again")
	cursor, err := t.TransferCollection.Find(context.TODO(), bson.D{{"created_at", bson.D{{"$eq", time.Now().Add(-30 * time.Minute)}}}})
	if err != nil {
		logrus.Fatalf("[Mongo]: error fetching transactions: %s", err.Error())
		return err
	}
	if err = cursor.All(context.TODO(), &transfers); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to transfer model: %s", err.Error())
		return err
	}

	// update fetched transactions
	for _, t := range *transfers {
		if t.Status == "pending" {
		}
	}
	logrus.Info("Completed Transaction Time Limit Cron")
	fmt.Println("done")
	return nil
}