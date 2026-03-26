package history

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	"github.com/ethanrous/weblens/modules/wlfs"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var fileIDIndexKey = "fileID_index"
var filepathIndexKey = "filepath_index"
var originPathIndexKey = "originPath_index"
var destinationPathIndexKey = "destinationPath_index"

// IndexModels defines MongoDB indexes for the fileHistory collection.
var IndexModels = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "fileID", Value: 1}},
		Options: options.Index().SetName(fileIDIndexKey),
	},
	{
		Keys:    bson.D{{Key: "filepath", Value: 1}},
		Options: options.Index().SetName(filepathIndexKey),
	},
	{
		Keys:    bson.D{{Key: "originPath", Value: 1}},
		Options: options.Index().SetName(originPathIndexKey),
	},
	{
		Keys:    bson.D{{Key: "destinationPath", Value: 1}},
		Options: options.Index().SetName(destinationPathIndexKey),
	},
}

func init() {
	startup.RegisterHook(registerIndexes)
}

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
							bson.D{
								{Key: "actionType", Value: "fileRestore"},
								{Key: "filepath", Value: filepath.String()},
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

// GetLifetimesOptions specifies filtering options for retrieving file lifetimes.
type GetLifetimesOptions struct {
	ActiveOnly bool
	PathPrefix wlfs.Filepath
	Depth      int
	TowerID    string
}

// compileLifetimeOpts merges multiple GetLifetimesOptions into a single compiled option set.
// The function applies the last non-zero value for each option field, with a minimum depth of 1.
func compileLifetimeOpts(opts ...GetLifetimesOptions) GetLifetimesOptions {
	o := GetLifetimesOptions{}
	o.Depth = 1 // Minimum depth

	for _, opt := range opts {
		if !opt.PathPrefix.IsZero() {
			o.PathPrefix = opt.PathPrefix
		}

		if opt.ActiveOnly {
			o.ActiveOnly = opt.ActiveOnly
		}

		if opt.Depth != 0 && opt.Depth > o.Depth {
			o.Depth = opt.Depth
		}

		if opt.TowerID != "" {
			o.TowerID = opt.TowerID
		}
	}

	return o
}

// GetLifetimes retrieves file lifetimes based on the provided filtering options.
func GetLifetimes(ctx context.Context, opts ...GetLifetimesOptions) ([]FileLifetime, error) {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return nil, err
	}

	o := compileLifetimeOpts(opts...)

	pipe := bson.A{bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: pathPrefixReFilter(o.PathPrefix, o.Depth)["$or"]}}}}}

	if o.TowerID != "" {
		pipe = append(pipe, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "towerID", Value: o.TowerID},
			}},
		})
	}

	fileIDGroup := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$fileID"},
			{Key: "actions", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		},
		},
	}

	if o.ActiveOnly {
		pipe = append(pipe,
			bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: 1}}}},
			fileIDGroup,
			bson.D{{Key: "$match", Value: bson.D{{Key: "actions", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$elemMatch", Value: bson.D{{Key: "actionType", Value: "fileDelete"}}}}}}}}}},
			// Compute the file's current path from its last action (destinationPath for moves, filepath otherwise)
			bson.D{
				{Key: "$addFields", Value: bson.D{
					{Key: "lastAction", Value: bson.D{{Key: "$last", Value: "$actions"}}},
				}},
			},
			bson.D{
				{Key: "$addFields", Value: bson.D{
					{Key: "currentPath", Value: bson.D{
						{Key: "$ifNull", Value: bson.A{"$lastAction.destinationPath", "$lastAction.filepath"}},
					}},
				}},
			},
			// Dedup by current path: if multiple files claim the same path, keep the most recently active one
			bson.D{{Key: "$sort", Value: bson.D{{Key: "lastAction.timestamp", Value: -1}}}},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$currentPath"},
					{Key: "doc", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
				}},
			},
			bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$doc"}}}},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "lastAction", Value: 0},
					{Key: "currentPath", Value: 0},
				}},
			},
		)
	} else {
		pipe = append(pipe, fileIDGroup)
	}

	cur, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	var lifetimes []FileLifetime

	err = cur.All(ctx, &lifetimes)
	if err != nil {
		return nil, db.WrapError(err, "GetLifetimes")
	}

	return lifetimes, nil
}

func registerIndexes(ctx context.Context, _ config.Provider) error {
	col, err := db.GetCollection[any](ctx, FileHistoryCollectionKey)
	if err != nil {
		return err
	}

	for _, idx := range IndexModels {
		if err := col.NewIndex(idx); err != nil {
			return err
		}
	}

	return nil
}
