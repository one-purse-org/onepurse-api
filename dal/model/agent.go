package model

type Agent struct {
	ID          string `bson:"_id" json:"id"`
	FullName    string `bson:"full_name" json:"full_name"`
	Email       string `bson:"email" json:"email"`
	Phone       string `bson:"phone" json:"phone"`
	Wallet      Wallet `bson:"wallet" json:"wallet"` // A Merchant can only trade one currency hence just one wallet is needed
	DeviceToken string `bson:"device_token" json:"device_token"`
}
