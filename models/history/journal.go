package history

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"go.mongodb.org/mongo-driver/bson"
)

// FileHistoryCollectionKey is the MongoDB collection name for storing file history.
const FileHistoryCollectionKey = "fileHistory"

// DoesFileExistInHistory checks if a file exists at the given filepath according to the file history.
func DoesFileExistInHistory(ctx context.Context, filepath file_system.Filepath) (*FileAction, error) {
	pipe := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "$or",
						Value: bson.A{
							bson.D{
								{Key: "actionType", Value: "fileCreate"},
								{Key: "filepath", Value: filepath.String()},
							},
							bson.D{
								{Key: "actionType", Value: "fileMove"},
								{Key: "destinationPath", Value: filepath.String()},
							},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "fileHistory"},
					{Key: "let",
						Value: bson.D{
							{Key: "fileID", Value: "$fileID"},
							{Key: "arrivalTime", Value: "$timestamp"},
						},
					},
					{Key: "pipeline",
						Value: bson.A{
							bson.D{
								{Key: "$match",
									Value: bson.D{
										{Key: "$expr",
											Value: bson.D{
												{Key: "$and",
													Value: bson.A{
														bson.D{
															{Key: "$eq",
																Value: bson.A{
																	"$fileID",
																	"$$fileID",
																},
															},
														},
														bson.D{
															{Key: "$gt",
																Value: bson.A{
																	"$timestamp",
																	"$$arrivalTime",
																},
															},
														},
														bson.D{
															{Key: "$in",
																Value: bson.A{
																	"$actionType",
																	bson.A{
																		"fileMove",
																		"fileDelete",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							bson.D{{Key: "$limit", Value: 1}},
						},
					},
					{Key: "as", Value: "subsequentActions"},
				},
			},
		},
		bson.D{{Key: "$match", Value: bson.D{{Key: "subsequentActions", Value: bson.D{{Key: "$size", Value: 0}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 1}},
		bson.D{
			{Key: "$project",
				Value: bson.D{
					{Key: "fileID", Value: 1},
					{Key: "filepath",
						Value: bson.D{
							{Key: "$ifNull",
								Value: bson.A{
									"$destinationPath",
									"$filepath",
								},
							},
						},
					},
					{Key: "arrivedAt", Value: "$timestamp"},
					{Key: "actionType", Value: 1},
					{Key: "exists", Value: bson.D{{Key: "$literal", Value: true}}},
				},
			},
		},
	}

	col, err := db.GetCollection[FileAction](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	res, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	var target []*FileAction

	err = res.All(ctx, &target)
	if err != nil {
		return nil, err
	}

	if len(target) == 0 {
		return nil, nil
	}

	return target[0], nil
}
