package dal

import (
	"context"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type ITransactionDAL interface {
	CreateAccount(ctx context.Context, account *model.Account) error
	CreateTransfer(ctx context.Context, transfer *model.Transfer) error
	CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error
	CreateDeposit(ctx context.Context, deposit *model.Deposit) error
	CreateExchange(ctx context.Context, exchange *model.Exchange) error
	CreateOnePurseTransaction(ctx context.Context, transaction *model.OnePurseTransaction) error
	CreateRate(ctx context.Context, transaction *model.Rate) error
	CreateAdminPayment(ctx context.Context, payments *model.AdminPayment) error

	GetAccount(ctx context.Context, query bson.D) (*model.Account, error)
	GetTransferByID(ctx context.Context, transferID string) (*model.Transfer, error)
	GetWithdrawalByID(ctx context.Context, withdrawalID string) (*model.Withdrawal, error)
	GetDepositByID(ctx context.Context, depositID string) (*model.Deposit, error)
	GetExchangeByID(ctx context.Context, exchangeID string) (*model.Exchange, error)
	GetOnePurseTransactionByID(ctx context.Context, transactionID string) (*model.OnePurseTransaction, error)
	GetRate(ctx context.Context) (*model.Rate, error)
	GetAdminPayment(ctx context.Context, query bson.D) (*model.AdminPayment, error)

	UpdateAccount(ctx context.Context, accountID string, updateParam bson.D) error
	UpdateTransfer(ctx context.Context, transferID string, updateParam bson.D) error
	UpdateWithdrawal(ctx context.Context, withdrawalID string, updateParam bson.D) error
	UpdateDeposit(ctx context.Context, depositID string, updateParam bson.D) error
	UpdateExchange(ctx context.Context, exchangeID string, updateParam bson.D) error
	UpdateOnePurseTransaction(ctx context.Context, transactionID string, updateParam bson.D) error
	UpdateRate(ctx context.Context, updateParam bson.D) error
	UpdateAdminPayment(ctx context.Context, ID string, updateParam bson.D) error

	FetchAccounts(ctx context.Context, query bson.D) (*[]model.Account, error)
	FetchTransfers(ctx context.Context, query bson.D) (*[]model.Transfer, error)
	FetchWithdrawals(ctx context.Context, query bson.D) (*[]model.Withdrawal, error)
	FetchDeposits(ctx context.Context, query bson.D) (*[]model.Deposit, error)
	FetchExchanges(ctx context.Context, query bson.D) (*[]model.Exchange, error)
	FetchOnePurseTransactions(ctx context.Context, query bson.D) (*[]model.OnePurseTransaction, error)
	FetchAdminPayments(ctx context.Context, query bson.D) (*[]model.AdminPayment, error)

	CountAll(ctx context.Context) (int32, error)
	CheckTimeLimit() error
}
type TransactionDAL struct {
	DB                            *mongo.Database
	TransferCollection            *mongo.Collection
	WithdrawalCollection          *mongo.Collection
	DepositCollection             *mongo.Collection
	ExchangeCollection            *mongo.Collection
	OnePurseTransactionCollection *mongo.Collection
	AccountCollection             *mongo.Collection
	RateCollection                *mongo.Collection
	AdminPaymentCollection        *mongo.Collection
}

func NewTransactionDAL(db *mongo.Database) *TransactionDAL {
	return &TransactionDAL{
		DB:                            db,
		TransferCollection:            db.Collection("transfer"),
		WithdrawalCollection:          db.Collection("withdraw"),
		DepositCollection:             db.Collection("deposit"),
		ExchangeCollection:            db.Collection("exchange"),
		OnePurseTransactionCollection: db.Collection("one-purse-transaction"),
		AccountCollection:             db.Collection("account"),
		RateCollection:                db.Collection("rate"),
		AdminPaymentCollection:        db.Collection("admin-payments"),
	}
}

// CreateAccount creates a database record of a user or agent bank account
func (t TransactionDAL) CreateAccount(ctx context.Context, account *model.Account) error {
	_, err := t.AccountCollection.InsertOne(ctx, account)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("Account record already exists. You might be duplicating an account")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateTransfer(ctx context.Context, transfer *model.Transfer) error {
	_, err := t.TransferCollection.InsertOne(ctx, transfer)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("Transfer record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error {
	_, err := t.WithdrawalCollection.InsertOne(ctx, withdrawal)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("withdrawal record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateDeposit(ctx context.Context, deposit *model.Deposit) error {
	_, err := t.DepositCollection.InsertOne(ctx, deposit)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("deposit record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateExchange(ctx context.Context, exchange *model.Exchange) error {
	_, err := t.ExchangeCollection.InsertOne(ctx, exchange)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("exchange record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateOnePurseTransaction(ctx context.Context, transaction *model.OnePurseTransaction) error {
	_, err := t.OnePurseTransactionCollection.InsertOne(ctx, transaction)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("transaction record already exists. You might be repeating a transaction")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateAdminPayment(ctx context.Context, payment *model.AdminPayment) error {
	_, err := t.AdminPaymentCollection.InsertOne(ctx, payment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("payment record already exists. You might be repeating a payment")
		}
		return err
	}
	return nil
}

func (t TransactionDAL) CreateRate(ctx context.Context, rate *model.Rate) error {
	_, err := t.RateCollection.InsertOne(ctx, rate)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("rate record already exist. You might be duplicating a rate")
		}
		return err
	}
	return nil
}

// GetAccount fetches bank account information based on the specified query parameters ...
func (t TransactionDAL) GetAccount(ctx context.Context, query bson.D) (*model.Account, error) {
	var account *model.Account
	err := t.AccountCollection.FindOne(ctx, query).Decode(account)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("no account matches specified query: %s", query)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return account, nil
}

func (t TransactionDAL) GetTransferByID(ctx context.Context, transferID string) (*model.Transfer, error) {
	var transfer *model.Transfer
	err := t.TransferCollection.FindOne(ctx, bson.D{{"_id", transferID}}).Decode(transfer)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for trasfer %s not found", transferID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return transfer, nil
}

func (t TransactionDAL) GetWithdrawalByID(ctx context.Context, withdrawalID string) (*model.Withdrawal, error) {
	var withdrawal *model.Withdrawal
	err := t.WithdrawalCollection.FindOne(ctx, bson.D{{"_id", withdrawalID}}).Decode(withdrawal)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for withdrawal %s not found", withdrawalID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return withdrawal, nil
}

func (t TransactionDAL) GetDepositByID(ctx context.Context, depositID string) (*model.Deposit, error) {
	var deposit *model.Deposit
	err := t.DepositCollection.FindOne(ctx, bson.D{{"_id", depositID}}).Decode(deposit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for deposit %s not found", depositID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return deposit, nil
}

func (t TransactionDAL) GetExchangeByID(ctx context.Context, exchangeID string) (*model.Exchange, error) {
	var exchange *model.Exchange
	err := t.ExchangeCollection.FindOne(ctx, bson.D{{"_id", exchangeID}}).Decode(exchange)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for exchange %s not found", exchangeID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return exchange, nil
}

func (t TransactionDAL) GetOnePurseTransactionByID(ctx context.Context, transactionID string) (*model.OnePurseTransaction, error) {
	var transaction *model.OnePurseTransaction
	err := t.OnePurseTransactionCollection.FindOne(ctx, bson.D{{"_id", transactionID}}).Decode(transaction)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for transaction %s not found", transactionID)
			return nil, errors.New(findErr)
		}
		return nil, err
	}

	return transaction, nil
}

func (t TransactionDAL) GetAdminPayment(ctx context.Context, query bson.D) (*model.AdminPayment, error) {
	var payment *model.AdminPayment
	err := t.AdminPaymentCollection.FindOne(ctx, query).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("record for admin payment not found")
		}
		return nil, err
	}
	return payment, nil
}

func (t TransactionDAL) GetRate(ctx context.Context) (*model.Rate, error) {
	var rate *model.Rate
	err := t.RateCollection.FindOne(ctx, bson.D{}).Decode(&rate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("record for rate not found")
		}
		return nil, err
	}
	return rate, nil
}

// FetchAccounts fetches a list of account based on the specified query ...
func (t TransactionDAL) FetchAccounts(ctx context.Context, query bson.D) (*[]model.Account, error) {
	var account []model.Account

	cursor, err := t.AccountCollection.Find(ctx, query)
	if err != nil {
		logrus.Errorf("[Mongo]: error fetching accounts: %s", err.Error())
		return nil, err
	}
	if err = cursor.All(ctx, &account); err != nil {
		logrus.Errorf("[Mongo]: error decoding account results: %s", err.Error())
		return nil, err
	}
	return &account, nil
}

func (t TransactionDAL) FetchTransfers(ctx context.Context, query bson.D) (*[]model.Transfer, error) {
	var transfers []model.Transfer

	cursor, err := t.TransferCollection.Find(ctx, query)
	if err != nil {
		logrus.Errorf("[Mongo]: error fetching transfers: %s", err.Error())
		return nil, err
	}
	if err = cursor.All(ctx, &transfers); err != nil {
		logrus.Errorf("[Mongo]: error decoding transfer results: %s", err.Error())
		return nil, err
	}
	return &transfers, nil
}

func (t TransactionDAL) FetchDeposits(ctx context.Context, query bson.D) (*[]model.Deposit, error) {
	var deposits []model.Deposit

	cursor, err := t.DepositCollection.Find(ctx, query)
	if err != nil {
		log.Fatalf("[Mongo]: error decoding deposit results: %s", err.Error())
		return nil, err
	}
	if err = cursor.All(ctx, &deposits); err != nil {
		log.Fatalf("[Mongo]: error decoding deposit results: %s", err.Error())
		return nil, err
	}
	return &deposits, nil
}

func (t TransactionDAL) FetchWithdrawals(ctx context.Context, query bson.D) (*[]model.Withdrawal, error) {
	var withdrawals []model.Withdrawal

	cursor, err := t.WithdrawalCollection.Find(ctx, query)
	if err != nil {
		log.Fatalf("[Mongo]: error fetching withdrawals: %s", err.Error())
		return nil, err
	}
	if err = cursor.All(ctx, &withdrawals); err != nil {
		log.Fatalf("[Mongo]: error decoding withdrawal results: %s", err.Error())
		return nil, err
	}
	return &withdrawals, err
}

func (t TransactionDAL) FetchExchanges(ctx context.Context, query bson.D) (*[]model.Exchange, error) {
	var exchanges []model.Exchange

	cursor, err := t.ExchangeCollection.Find(ctx, query)
	if err != nil {
		log.Fatalf("[Mongo]: error fetching exchanges: %s", err.Error())
		return nil, err
	}
	if err = cursor.All(ctx, &exchanges); err != nil {
		log.Fatalf("[Mongo]: error decoding exchange results: %s", err.Error())
		return nil, err
	}
	return &exchanges, nil
}

func (t TransactionDAL) FetchOnePurseTransactions(ctx context.Context, query bson.D) (*[]model.OnePurseTransaction, error) {
	var transactions []model.OnePurseTransaction

	cursor, err := t.OnePurseTransactionCollection.Find(ctx, query)
	if err != nil {
		logrus.Errorf("[Mongo]: error fetching transactions: %s", err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &transactions); err != nil {
		log.Fatalf("[Mongo]: error decoding transaction results: %s", err.Error())
		return nil, err
	}
	return &transactions, nil
}

func (t TransactionDAL) FetchAdminPayments(ctx context.Context, query bson.D) (*[]model.AdminPayment, error) {
	var payments []model.AdminPayment

	cursor, err := t.AdminPaymentCollection.Find(ctx, query)
	if err != nil {
		logrus.Errorf("[Mongo]: error fetching admin payments: %s", err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &payments); err != nil {
		logrus.Errorf("[Mongo]: error decoding admin payment results: %s", err.Error())
		return nil, err
	}
	return &payments, nil
}

//UpdateAccount updates an account information ...
func (t TransactionDAL) UpdateAccount(ctx context.Context, accountID string, updateParam bson.D) error {
	result, err := t.AccountCollection.UpdateByID(ctx, accountID, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error updating account %s information: %s", accountID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Errorf("[Mongo]: error updating account %s information: account record not found", accountID)
		return errors.New("account record not found")
	}
	return nil
}

func (t TransactionDAL) UpdateTransfer(ctx context.Context, transferID string, updateParam bson.D) error {
	result, err := t.TransferCollection.UpdateByID(ctx, transferID, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error updating transfer %s: %s", transferID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating transfer %s : transfer record not found", transferID)
		return errors.New("transfer record not found")
	}
	return nil
}

func (t TransactionDAL) UpdateWithdrawal(ctx context.Context, withdrawalID string, updateParam bson.D) error {
	result, err := t.WithdrawalCollection.UpdateByID(ctx, withdrawalID, updateParam)
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

func (t TransactionDAL) UpdateDeposit(ctx context.Context, depositID string, updateParam bson.D) error {
	result, err := t.DepositCollection.UpdateByID(ctx, depositID, updateParam)
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

func (t TransactionDAL) UpdateExchange(ctx context.Context, exchangeID string, updateParam bson.D) error {
	result, err := t.ExchangeCollection.UpdateByID(ctx, exchangeID, updateParam)
	if err != nil {
		logrus.Fatalf("[Mongo]: error updating exchange %s: %s", exchangeID, err.Error())
		return err
	}

	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating exchange %s :  record not found", exchangeID)
		return errors.New("exchange record not found")
	}

	return nil
}

func (t TransactionDAL) UpdateOnePurseTransaction(ctx context.Context, transactionID string, updateParam bson.D) error {
	result, err := t.OnePurseTransactionCollection.UpdateByID(ctx, transactionID, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error updating transaction %s: %s", transactionID, err.Error())
		return err
	}

	if result.MatchedCount == 0 {
		logrus.Errorf("[Mongo] error updating transaction %s: record not found", transactionID)
		return errors.New("transaction record not found")
	}
	return nil
}

func (t TransactionDAL) UpdateAdminPayment(ctx context.Context, ID string, updateParam bson.D) error {
	result, err := t.AdminPaymentCollection.UpdateByID(ctx, ID, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error updating payment %s: %s", ID, err.Error())
		return err
	}

	if result.MatchedCount == 0 {
		logrus.Errorf("[Mongo]: error updating payment %s: record not found", ID)
		return errors.New("payment record not found")
	}
	return nil
}

func (t TransactionDAL) UpdateRate(ctx context.Context, updateParam bson.D) error {
	result, err := t.RateCollection.UpdateOne(ctx, bson.D{}, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error updating rate: %s", err.Error())
		return err
	}

	if result.MatchedCount == 0 {
		logrus.Errorf("[Mongo]: error updating rate: record not found")
		return errors.New("rate record not found")
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
	return nil
}

func (t TransactionDAL) CountAll(ctx context.Context) (int32, error) {
	nT, err := t.TransferCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	nOT, err := t.OnePurseTransactionCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	nE, err := t.ExchangeCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	nW, err := t.WithdrawalCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	nD, err := t.DepositCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	total := nOT + nE + nT + nW + nD
	return int32(total), nil
}
