package interfaces

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
	ID 				primitive.ObjectID 	`bson:"_id"`
	FileHash		string				`bson:"fileHash"`
	Filename 		string 				`bson:"filename"`
	BlurHash 		string 				`bson:"blurHash"`
	Thumbnail64 	string		 		`bson:"thumbnail"`
	ThumbWidth 		int					`bson:"width"`
	ThumbHeight 	int 				`bson:"height"`
	CreateDate		time.Time			`bson:"createDate"`
}