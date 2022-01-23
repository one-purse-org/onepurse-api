package model

type Agent struct {
	ID          string `bson:"id" json:"id"`
	Agent       *Agent `bson:"agent" json:"agent"`
	Name        string `bson:"name" json:"name"`
	Wallet      Wallet `bson:"wallet" json:"wallet"` // A Merchant can only trade one currency hence just one wallet is needed
	DeviceToken string `bson:"device_token" json:"device_token"`
}
