package interfaces

type User struct {
	Id			string `bson:"_id"`
	Username	string `bson:"username"`
	Password 	string `bson:"password"`
	Tokens 		[]string `bson:"tokens"`
}