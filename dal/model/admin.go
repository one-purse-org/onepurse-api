package model

// Admin is the struct that defines an admin
type Admin struct {
	ID       string `bson:"_id"`
	FullName string `bson:"full_name" json:"full_name"`
	Username string `bson:"username" json:"username"`
	Email    string `bson:"email" json:"email"`
	Phone    string `bson:"phone" json:"phone"`
	Avatar   string `bson:"avatar" json:"avatar"`
	Role     *Role  `bson:"role" json:"role"`
}

type NewPasswordChallengeInput struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Session  string `json:"session"`
}

// Currency is the struct that defines currencies
type Currency struct {
	Label string `bson:"label" json:"label"`
	Slug  string `bson:"slug" json:"slug"`
	Icon  string `bson:"icon" json:"icon"`
}

// Role defines the admin roles and access
type Role struct {
	Name   string   `bson:"name" json:"name"`
	Slug   string   `bson:"slug" json:"slug"`
	Access []Access `bson:"access" json:"access"`
}

type Access struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// NumberMetrics defines the model for number of user and transaction metrics
type NumberMetrics struct {
	NumberOfUser        int32 `json:"number_of_user"`
	NumberOfAgent       int32 `json:"number_of_agent"`
	NumberOfTransaction int32 `json:"number_of_transaction"`
}

// TransactionVolumeMetrics defines the model for transaction metrics
type TransactionVolumeMetrics struct {
	Values map[string]float32 `json:"values"`
}

// CurrencyVolume defines the model for the cu
type CurrencyVolume struct {
	Fiat   map[string]float32 `json:"fiat"`
	Crypto map[string]float32 `json:"crypto"`
}

type CurrencyMetrics struct {
	InApp      CurrencyVolume `json:"in_app" bson:"in_app"`
	Deposit    CurrencyVolume `json:"deposit" bson:"deposit"`
	Withdrawal CurrencyVolume `json:"withdrawal" bson:"withdrawal"`
	Canceled   CurrencyVolume `json:"canceled" bson:"canceled"`
}

type Metric struct {
	NumberMetric      *NumberMetrics            `json:"number_metric"`
	TransactionMetric *TransactionVolumeMetrics `json:"transaction_metric"`
	CurrencyMetric    *CurrencyMetrics          `json:"currency_metric"`
}

const DASHBOARD = "dashboard"
const VERIFICATION = "verification"
const DEACTIVATE = "deactivate"
const TRANSACTION = "transaction"
const RATES = "rates"
const REVENUE = "revenue"
const ADMIN_PAYMENT = "admin-payment"
const ACTIVITY_LOGS = "activity-logs"
const MANAGE_PERSONEL = "manage-admins"
const TWO_FACTOR_AUTH = "2fa"

// Defined Role
var dashboardRole = &Access{
	Name:        DASHBOARD,
	Slug:        DASHBOARD + "-role",
	Description: "can view all insights on the dashboard",
}

var verificationRole = &Access{
	Name:        VERIFICATION,
	Slug:        VERIFICATION + "-role",
	Description: "can approve and reject an agent and user verification",
}

var transactionRole = &Access{
	Name:        TRANSACTION,
	Slug:        TRANSACTION + "-role",
	Description: "can complete, cancel and dispute transactions",
}

var rateRole = &Access{
	Name:        RATES,
	Slug:        RATES + "-role",
	Description: "can update exchange rates",
}

var revenueRole = &Access{
	Name:        REVENUE,
	Slug:        REVENUE + "-role",
	Description: "can view revenues",
}

var paymentRole = &Access{
	Name:        ADMIN_PAYMENT,
	Slug:        ADMIN_PAYMENT + "-role",
	Description: "can add admin payment",
}

var activityLogRole = &Access{
	Name:        ACTIVITY_LOGS,
	Slug:        ACTIVITY_LOGS + "-role",
	Description: "can view admin activity logs",
}

var managePersonelRole = &Access{
	Name:        MANAGE_PERSONEL,
	Slug:        MANAGE_PERSONEL + "-role",
	Description: "can add new admin and agent",
}

var deactivateRole = &Access{
	Name:        DEACTIVATE,
	Slug:        DEACTIVATE + "-role",
	Description: "can block and unblock user and agents",
}

var twoFARole = &Access{
	Name:        TWO_FACTOR_AUTH,
	Slug:        TWO_FACTOR_AUTH + "-role",
	Description: "can generate 2fa security code",
}
