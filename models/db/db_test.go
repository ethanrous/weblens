package db

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testCollectionKey = "testCollection"

// TestDocument is a sample document structure for testing
type TestDocument struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Value     int                `bson:"value"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func TestConnectToMongoWithCancelledContext(t *testing.T) {
	// Create a context and immediately cancel it
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Attempt to connect with cancelled context
	mongodb, err := ConnectToMongo(ctx, config.GetMongoDBUri(), testCollectionKey)

	assert.Error(t, err)
	assert.Nil(t, mongodb)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestContextualizedCollection_Basic(t *testing.T) {
	ctx := SetupTestDB(t, testCollectionKey)

	t.Run("GetCollection", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
	})

	t.Run("GetCollection_NoDatabase", func(t *testing.T) {
		collection, err := GetCollection(context.Background(), testCollectionKey)
		assert.Error(t, err)
		assert.Nil(t, collection)
		assert.True(t, errors.Is(err, ErrNoDatabase))
	})

	t.Run("GetCollection_NilContext", func(t *testing.T) {
		collection, err := GetCollection(nil, testCollectionKey)
		assert.Error(t, err)
		assert.Nil(t, collection)
		assert.True(t, errors.Is(err, ErrNoDatabase))
	})
}

func TestContextualizedCollection_CRUD(t *testing.T) {
	ctx := SetupTestDB(t, testCollectionKey)

	t.Run("InsertOne", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		doc := TestDocument{
			ID:        primitive.NewObjectID(),
			Name:      "test",
			Value:     42,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result, err := collection.InsertOne(ctx, doc)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, doc.ID, result.InsertedID)
	})

	t.Run("InsertMany", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		docs := []interface{}{
			TestDocument{
				ID:        primitive.NewObjectID(),
				Name:      "test1",
				Value:     1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			TestDocument{
				ID:        primitive.NewObjectID(),
				Name:      "test2",
				Value:     2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		result, err := collection.InsertMany(ctx, docs)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result.InsertedIDs))
	})

	t.Run("FindOne", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		doc := TestDocument{
			ID:        primitive.NewObjectID(),
			Name:      "findTest",
			Value:     42,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = collection.InsertOne(ctx, doc)
		require.NoError(t, err)

		var found TestDocument
		err = collection.FindOne(ctx, bson.M{"name": "findTest"}).Decode(&found)
		assert.NoError(t, err)
		assert.Equal(t, doc.ID, found.ID)
		assert.Equal(t, doc.Name, found.Name)
		assert.Equal(t, doc.Value, found.Value)
	})

	t.Run("UpdateOne", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		doc := TestDocument{
			ID:        primitive.NewObjectID(),
			Name:      "updateTest",
			Value:     42,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = collection.InsertOne(ctx, doc)
		require.NoError(t, err)

		update := bson.M{"$set": bson.M{"value": 43}}
		result, err := collection.UpdateOne(ctx, bson.M{"_id": doc.ID}, update)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		var updated TestDocument
		err = collection.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&updated)
		assert.NoError(t, err)
		assert.Equal(t, 43, updated.Value)
	})

	t.Run("DeleteOne", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		doc := TestDocument{
			ID:        primitive.NewObjectID(),
			Name:      "deleteTest",
			Value:     42,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = collection.InsertOne(ctx, doc)
		require.NoError(t, err)

		result, err := collection.DeleteOne(ctx, bson.M{"_id": doc.ID})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)

		err = collection.FindOne(ctx, bson.M{"_id": doc.ID}).Err()
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})
}

// func TestContextualizedCollection_Concurrency(t *testing.T) {
// 	mongodb, ctx := setupTestDB(t)
// 	ctx = context.WithValue(ctx, DatabaseContextKey, mongodb)
//
// 	t.Run("ConcurrentInserts", func(t *testing.T) {
// 		collection, err := GetCollection(ctx, testCollectionKey)
// 		require.NoError(t, err)
//
// 		const numGoroutines = 10
// 		const insertsPerGoroutine = 100
//
// 		var wg sync.WaitGroup
// 		wg.Add(numGoroutines)
//
// 		for i := 0; i < numGoroutines; i++ {
// 			go func(routineNum int) {
// 				defer wg.Done()
//
// 				for j := 0; j < insertsPerGoroutine; j++ {
// 					doc := TestDocument{
// 						ID:        primitive.NewObjectID(),
// 						Name:      fmt.Sprintf("concurrent_%d_%d", routineNum, j),
// 						Value:     j,
// 						CreatedAt: time.Now(),
// 						UpdatedAt: time.Now(),
// 					}
//
// 					_, err := collection.InsertOne(ctx, doc)
// 					require.NoError(t, err)
// 				}
// 			}(i)
// 		}
//
// 		wg.Wait()
//
// 		count, err := collection.CountDocuments(ctx, bson.M{})
// 		assert.NoError(t, err)
// 		assert.Equal(t, int64(numGoroutines*insertsPerGoroutine), count)
// 	})
//
// 	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
// 		collection, err := GetCollection(ctx, testCollectionKey)
// 		require.NoError(t, err)
//
// 		const numDocs = 100
// 		docs := make([]interface{}, numDocs)
// 		for i := 0; i < numDocs; i++ {
// 			docs[i] = TestDocument{
// 				ID:        primitive.NewObjectID(),
// 				Name:      fmt.Sprintf("doc_%d", i),
// 				Value:     i,
// 				CreatedAt: time.Now(),
// 				UpdatedAt: time.Now(),
// 			}
// 		}
//
// 		_, err = collection.InsertMany(ctx, docs)
// 		require.NoError(t, err)
//
// 		var wg sync.WaitGroup
// 		const numReaders = 5
// 		const numWriters = 3
// 		wg.Add(numReaders + numWriters)
//
// 		// Start readers
// 		for i := 0; i < numReaders; i++ {
// 			go func() {
// 				defer wg.Done()
// 				for j := 0; j < 50; j++ {
// 					var doc TestDocument
// 					err := collection.FindOne(ctx, bson.M{}).Decode(&doc)
// 					require.NoError(t, err)
// 				}
// 			}()
// 		}
//
// 		// Start writers
// 		for i := 0; i < numWriters; i++ {
// 			go func(writerNum int) {
// 				defer wg.Done()
// 				for j := 0; j < 20; j++ {
// 					doc := TestDocument{
// 						ID:        primitive.NewObjectID(),
// 						Name:      fmt.Sprintf("writer_%d_%d", writerNum, j),
// 						Value:     j,
// 						CreatedAt: time.Now(),
// 						UpdatedAt: time.Now(),
// 					}
// 					_, err := collection.InsertOne(ctx, doc)
// 					require.NoError(t, err)
// 				}
// 			}(i)
// 		}
//
// 		wg.Wait()
// 	})
// }

func TestContextualizedCollection_EdgeCases(t *testing.T) {
	ctx := SetupTestDB(t, testCollectionKey)

	t.Run("EmptyDocument", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		_, err = collection.InsertOne(ctx, bson.M{})
		assert.NoError(t, err)
	})

	t.Run("LargeDocument", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		// Create a large document (close to 16MB limit)
		largeString := strings.Repeat("a", 15*1024*1024) // 15MB
		doc := bson.M{"data": largeString}

		_, err = collection.InsertOne(ctx, doc)
		assert.NoError(t, err)
	})

	t.Run("NestedDocument", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		doc := bson.M{
			"level1": bson.M{
				"level2": bson.M{
					"level3": bson.M{
						"level4": bson.M{
							"value": 42,
						},
					},
				},
			},
		}

		_, err = collection.InsertOne(ctx, doc)
		assert.NoError(t, err)

		var found bson.M
		err = collection.FindOne(ctx, bson.M{"level1.level2.level3.level4.value": 42}).Decode(&found)
		assert.NoError(t, err)
	})
}

func TestContextualizedCollection_Transactions(t *testing.T) {
	ctx := SetupTestDB(t, testCollectionKey)

	t.Run("SuccessfulTransaction", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		err = WithTransaction(ctx, func(ctx context.Context) error {
			_, err := collection.InsertOne(ctx, TestDocument{
				ID:   primitive.NewObjectID(),
				Name: "transactionTest",
			})
			return err
		})
		assert.NoError(t, err)

		var doc TestDocument
		err = collection.FindOne(ctx, bson.M{"name": "transactionTest"}).Decode(&doc)
		assert.NoError(t, err)
	})

	t.Run("RollbackTransaction", func(t *testing.T) {
		err := WithTransaction(ctx, func(ctx context.Context) error {
			collection, err := GetCollection(ctx, testCollectionKey)
			require.NoError(t, err)

			_, err = collection.InsertOne(ctx, TestDocument{
				ID:   primitive.NewObjectID(),
				Name: "rollbackTest",
			})
			require.NoError(t, err)

			return errors.New("force rollback")
		})
		assert.Error(t, err)

		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		err = collection.FindOne(ctx, bson.M{"name": "rollbackTest"}).Decode(&TestDocument{})
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})
}

func TestContextualizedCollection_Indexes(t *testing.T) {
	ctx := SetupTestDB(t, testCollectionKey)

	t.Run("CreateAndVerifyIndex", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		indexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		}

		_, err = collection.GetCollection().Indexes().CreateOne(ctx, indexModel)
		require.NoError(t, err)

		// Try to insert duplicate documents
		doc1 := TestDocument{
			ID:   primitive.NewObjectID(),
			Name: "uniqueTest",
		}
		doc2 := TestDocument{
			ID:   primitive.NewObjectID(),
			Name: "uniqueTest",
		}

		_, err = collection.InsertOne(ctx, doc1)
		require.NoError(t, err)

		_, err = collection.InsertOne(ctx, doc2)
		assert.Error(t, err) // Should fail due to unique index
	})
}

func TestContextualizedCollection_ErrorHandling(t *testing.T) {
	ctx := SetupTestDB(t, testCollectionKey)

	t.Run("InvalidObjectID", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		err = collection.FindOne(ctx, bson.M{"_id": "invalid"}).Err()
		assert.Error(t, err)
	})

	t.Run("InvalidQuery", func(t *testing.T) {
		collection, err := GetCollection(ctx, testCollectionKey)
		require.NoError(t, err)

		_, err = collection.Find(ctx, bson.M{"$invalid": 1})
		assert.Error(t, err)
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		collection, err := GetCollection(ctxTimeout, testCollectionKey)
		require.NoError(t, err)

		_, err = collection.Find(ctxTimeout, bson.M{})
		assert.Error(t, err)
	})
}
