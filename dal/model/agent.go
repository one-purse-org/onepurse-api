package model

type Agent struct {
	ID          string `bson:"_id" json:"id"`
	FullName    string `bson:"full_name" json:"full_name"`
	UserName    string `bson:"username" json:"username"`
	Email       string `bson:"email" json:"email"`
	Phone       string `bson:"phone" json:"phone"`
	Wallet      Wallet `bson:"wallet" json:"wallet"` // A Merchant can only trade one currency hence just one wallet is needed
	Approved    bool   `bson:"approved" json:"approved"`
	DeviceToken string `bson:"device_token" json:"device_token"`
}
