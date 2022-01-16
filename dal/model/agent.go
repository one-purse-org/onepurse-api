package model

type Agent struct {
	ID          string `bson:"id" json:"id"`
	Name        string `bson:"name" json:"name"`
	Address     string `bson:"address" json:"address"`
	Wallet      Wallet `bson:"wallet" json:"wallet"`
	DeviceToken string `bson:"device_token" json:"device_token"`
	IDType      string `bson:"id_type" json:"id_type"`
	IDNumber    string `bson:"id_number" json:"id_number"`
	IDImage     string `bson:"id_image" json:"id_image"`
}

type AgentAccount struct {
	ID          string `bson:"_id" json:"id"`
	Agent       *Agent `bson:"agent" json:"agent"`
	Name        string `bson:"name" json:"name"`
	Wallet      Wallet `bson:"wallet" json:"wallet"`
	DeviceToken string `bson:"device_token" json:"device_token"`
}
