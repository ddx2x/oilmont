package mongo

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/dict"
	"github.com/ddx2x/oilmont/pkg/datasource/gtm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	metadata          = "metadata"
	version           = "version"
	metadataName      = "metadata.name"
	metadataWorkspace = "metadata.workspace"
	metadataUUID      = "metadata.uuid"
	metadataDelete    = "metadata.is_delete"
)

var _ datasource.IStorage = &Mongo{}

func getCtx(client *mongo.Client) (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := client.Connect(ctx); err != nil {
		return nil, cancel, err
	}
	return ctx, cancel, nil
}

type Mongo struct {
	uri    string
	client *mongo.Client
	ctx    context.Context
}

func NewMongo(ctx context.Context, uri string) (*Mongo, error, chan error) {
	client, err := connect(uri)
	if err != nil {
		return nil, err, nil
	}

	investigationErrorChannel := make(chan error)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if err := client.Ping(ctx, readpref.Primary()); err != nil {
				investigationErrorChannel <- err
			}
		}
	}()

	mongo := &Mongo{uri: uri, client: client, ctx: ctx}
	if err := common.InitResourceConfigure(mongo); err != nil {
		panic(fmt.Errorf("init resource configure error: %s", err))
	}

	return mongo, nil, investigationErrorChannel
}

func connect(uri string) (*mongo.Client, error) {
	clientOptions := options.Client()
	clientOptions.SetRegistry(
		bson.NewRegistryBuilder().
			RegisterTypeMapEntry(
				bsontype.DateTime,
				reflect.TypeOf(time.Time{})).
			Build(),
	)
	clientOptions.ApplyURI(uri)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	ctx, cancel, err := getCtx(client)
	defer func() { cancel() }()
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

func (m *Mongo) Close() error {
	ctx, cancel, err := getCtx(m.client)
	if err != nil {
		return err
	}
	defer func() { cancel() }()
	return m.client.Disconnect(ctx)
}

func (m *Mongo) List(db, table, labels string, filterDelete bool) ([]interface{}, error) {
	var filter = bson.D{{}}
	if len(labels) > 0 {
		filter = expr2labels(labels)
	}
	if filterDelete {
		filter = append(filter, bson.E{Key: metadataDelete, Value: false})
	}
	findOptions := options.Find()

	cursor, err := m.client.
		Database(db).
		Collection(table).
		Find(m.ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	var _results []bson.M
	if err := cursor.All(m.ctx, &_results); err != nil {
		return nil, err
	}
	results := make([]interface{}, 0)
	for index := range _results {
		results = append(results, _results[index])
	}
	return results, nil
}

func (m *Mongo) GetByFilter(db, table string, result interface{}, filter map[string]interface{}, filterDelete bool) error {
	findOneOptions := options.FindOne()
	if filterDelete {
		filter[metadataDelete] = false
	}
	singleResult := m.client.
		Database(db).
		Collection(table).
		FindOne(m.ctx, map2filter(filter), findOneOptions)
	if err := singleResult.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return datasource.NotFound
		}
		return err
	}
	return nil
}

func (m *Mongo) Get(db, table, name string, result interface{}, filterDelete bool) error {
	query := bson.M{metadataName: name}
	if filterDelete {
		query[metadataDelete] = false
	}
	singleResult := m.client.Database(db).Collection(table).
		FindOne(context.Background(), query)
	if err := singleResult.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return datasource.NotFound
		}
		return err
	}
	return nil
}

func (m *Mongo) GetByMetadataUUID(db, table, uuid string, result interface{}, filterDelete bool) error {
	query := bson.M{metadataUUID: uuid}
	if filterDelete {
		query[metadataDelete] = false
	}
	findOneOptions := options.FindOne()
	singleResult := m.client.
		Database(db).
		Collection(table).
		FindOne(m.ctx, query, findOneOptions)
	if err := singleResult.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return datasource.NotFound
		}
		return err
	}
	return nil
}

func versionMatchFilter(opData map[string]interface{}, resourceVersion string) bool {
	if resourceVersion == "" {
		return false
	}
	metadata, exist := opData[metadata]
	if !exist {
		return false
	}
	metadataMap := metadata.(map[string]interface{})
	version, exist := metadataMap[version]
	if !exist {
		return false
	}
	if version.(string) <= resourceVersion {
		return false
	}
	return true
}

func fieldMatchFilter(opData map[string]interface{}, key string, value interface{}) bool {
	return reflect.DeepEqual(dict.Get(opData, key), value)
}

func (m *Mongo) checkExistAndCreate(ctx context.Context, db, table string) error {
	names, err := m.client.Database(db).ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return err
	}
	exist := false
	for _, name := range names {
		if name == table {
			exist = true
		}
	}
	if !exist {
		if err := m.client.Database(db).CreateCollection(ctx, table); err != nil {
			return err
		}

		indexModel := mongo.IndexModel{
			Keys:    bson.D{{"metadata.name", 1}, {"metadata.workspace", 1}},
			Options: options.Index().SetUnique(true),
		}
		if _, err = m.client.Database(db).Collection(table).Indexes().CreateOne(ctx, indexModel); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mongo) WatchEvent(ctx context.Context, db, table string, resourceVersion string, filters ...datasource.Filter) (<-chan core.Event, error) {
	if err := m.checkExistAndCreate(ctx, db, table); err != nil {
		return nil, err
	}
	ns := fmt.Sprintf("%s.%s", db, table)
	directReadFilter := func(op *gtm.Op) bool {
		pass := true
		if pass = versionMatchFilter(op.Data, resourceVersion); !pass {
			return pass
		}
		for _, filter := range filters {
			if pass = fieldMatchFilter(op.Data, filter.Key, filter.Value); !pass {
				return false
			}
		}
		return true
	}
	gtmCtx := gtm.Start(m.client, &gtm.Options{
		DirectReadNs:     []string{ns},
		ChangeStreamNs:   []string{ns},
		MaxAwaitTime:     10,
		DirectReadFilter: directReadFilter,
	})

	result := make(chan core.Event, 0)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(result)
				gtmCtx.Stop()
				return
			case <-gtmCtx.ErrC:
				close(result)
				return
			case op, ok := <-gtmCtx.OpC:
				if !ok {
					return
				}
				var opType core.EventType
				switch {
				case op.IsInsert():
					opType = core.ADDED
					if isDelete := dict.Get(op.Data, "metadata.is_delete"); isDelete != nil {
						if value, ok := isDelete.(bool); ok && value {
							continue
						}
					}
				case op.IsUpdate():
					opType = core.MODIFIED
					if isDelete := dict.Get(op.Data, "metadata.is_delete"); isDelete != nil {
						if value, ok := isDelete.(bool); ok && value {
							opType = core.DELETED
						}
					}
				case op.IsDelete():
					opType = core.DELETED
				}

				defaultObj := &core.DefaultObject{}
				if err := core.UnmarshalToIObject(op.Data, defaultObj); err != nil {
					continue
				}

				result <- core.Event{
					Type:   opType,
					Object: defaultObj,
				}
			}
		}
	}()

	return result, nil

}

func (m *Mongo) Watch(db, table string, resourceVersion string, watch datasource.WatchInterface, filters ...datasource.Filter) {
	ns := fmt.Sprintf("%s.%s", db, table)
	directReadFilter := func(op *gtm.Op) bool {
		pass := true
		if pass = versionMatchFilter(op.Data, resourceVersion); !pass {
			return pass
		}
		for _, filter := range filters {
			if pass = fieldMatchFilter(op.Data, filter.Key, filter.Value); !pass {
				return pass
			}
		}
		return pass
	}
	ctx := gtm.Start(m.client,
		&gtm.Options{
			DirectReadNs:     []string{ns},
			ChangeStreamNs:   []string{ns},
			MaxAwaitTime:     100,
			DirectReadFilter: directReadFilter,
		})

	go func(watch datasource.WatchInterface) {
		for {
			select {
			case err := <-ctx.ErrC:
				watch.ErrorStop() <- err
				return
			case <-watch.CloseStop():
				ctx.Stop()
				return
			case op, ok := <-ctx.OpC:
				if !ok {
					return
				}
				if err := watch.Handle(op.Data); err != nil {
					watch.ErrorStop() <- err
					return
				}
			}
		}
	}(watch)
}

func (m *Mongo) Create(db, table string, object core.IObject) (core.IObject, error) {
	if err := m.checkExistAndCreate(m.ctx, db, table); err != nil {
		return nil, err
	}
	if datasource.GetCoder(table) == nil {
		return nil, fmt.Errorf("not register code table %s", table)
	}
	object.SetKind(core.Kind(table))
	object.GenerateVersion()
	_, err := m.client.Database(db).Collection(table).InsertOne(m.ctx, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func isMongoDupKey(err error) bool {
	wces, ok := err.(mongo.WriteException)
	if !ok {
		return false
	}
	if len(wces.WriteErrors) > 0 {
		wce := wces.WriteErrors[0]
		return wce.Code == 11000 || wce.Code == 11001 || wce.Code == 12582 || wce.Code == 16460 && strings.Contains(wce.Message, " E11000 ")
	}

	if wces.WriteConcernError != nil {
		wce := wces.WriteConcernError
		return wce.Code == 11000 || wce.Code == 11001 || wce.Code == 12582 || wce.Code == 16460 && strings.Contains(wce.Message, " E11000 ")
	}

	return false
}

func (m *Mongo) InsertUnique(db, table string, id interface{}, data interface{}) error {
	_, err := m.client.Database(db).Collection(table).InsertOne(m.ctx, bson.M{"_id": id, "data": data})
	if isMongoDupKey(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) GetById(db, table, id string, result interface{}) error {
	return m.client.Database(db).
		Collection(table).
		FindOne(m.ctx, bson.M{"_id": id}).
		Decode(result)
}

func (m *Mongo) Bulk(db, table string, objects []core.IObject) error {
	if err := m.checkExistAndCreate(m.ctx, db, table); err != nil {
		return err
	}
	docs := make([]interface{}, len(objects))
	for i := range objects {
		docs[i] = objects[i]
	}
	_, err := m.client.Database(db).
		Collection(table).
		InsertMany(m.ctx, docs)
	return err
}

func (m *Mongo) RemoveTable(db, table string) error {
	return m.client.Database(db).Collection(table).Drop(m.ctx)
}

func (m *Mongo) ListToObject(db, table string, filter map[string]interface{}, result interface{}, filterDelete bool) error {
	if filter == nil {
		filter = make(map[string]interface{})
	}
	findOptions := options.Find()
	if filterDelete {
		filter[metadataDelete] = false
	}
	cursor, err := m.client.
		Database(db).
		Collection(table).
		Find(m.ctx, map2filter(filter), findOptions)
	if err != nil {
		return err
	}
	return cursor.All(m.ctx, result)
}

func (m *Mongo) ListByFilter(db, table string, filter map[string]interface{}, filterDelete bool) ([]interface{}, error) {
	findOptions := options.Find()
	if filterDelete {
		filter[metadataDelete] = false
	}
	cursor, err := m.client.
		Database(db).
		Collection(table).
		Find(m.ctx, map2filter(filter), findOptions)
	if err != nil {
		return nil, err
	}

	var _results []bson.M
	if err := cursor.All(m.ctx, &_results); err != nil {
		return nil, err
	}

	results := make([]interface{}, len(_results))
	for index := range _results {
		results[index] = _results[index]
	}
	return results, nil
}

func (m *Mongo) Apply(db, table, name string, newObject core.IObject, forceApply bool, paths ...string) (core.IObject, bool, error) {
	if err := m.checkExistAndCreate(m.ctx, db, table); err != nil {
		return nil, false, err
	}

	var update = false
	var query = bson.M{metadataName: name}
	if newObject.GetWorkspace() != "" {
		query[metadataWorkspace] = newObject.GetWorkspace()
	}
	if newObject.GetUUID() != "" {
		query[metadataUUID] = newObject.GetUUID()
	}

	singleResult := m.client.Database(db).Collection(table).FindOne(m.ctx, query)

	if singleResult.Err() == mongo.ErrNoDocuments {
		newObject.GenerateVersion()
		_, err := m.client.Database(db).Collection(table).InsertOne(m.ctx, newObject)
		if err != nil {
			return nil, false, err
		}
		return newObject, false, nil
	}

	old := newObject.Clone()
	if err := singleResult.Decode(old); err != nil {
		return nil, false, err
	}

	oldMap, err := core.ToMap(old)
	if err != nil {
		return nil, false, err
	}

	newMap, err := core.ToMap(newObject)
	if err != nil {
		return nil, false, err
	}

	if len(paths) == 0 {
		paths = []string{"spec"}
	}

	for _, path := range paths {
		if dict.CompareMergeObject(oldMap, newMap, path) {
			update = true
		}
	}

	if !update && !forceApply {
		return old, false, nil
	}

	if err := core.EncodeFromMap(newObject, oldMap); err != nil {
		return old, false, err
	}

	upsert := true
	newObject.GenerateVersion() //update version
	_, err = m.client.Database(db).Collection(table).
		ReplaceOne(m.ctx, query, newObject,
			options.MergeReplaceOptions(
				&options.ReplaceOptions{Upsert: &upsert},
			),
		)
	if err != nil {
		return old, true, err
	}

	return newObject, true, nil
}

func (m *Mongo) DeleteByIObject(db, table string, object core.IObject) error {
	query := bson.M{metadataName: object.GetName()}
	if object.GetWorkspace() != "" {
		query[metadataWorkspace] = object.GetWorkspace()
		query[metadataUUID] = object.GetUUID()
	}
	object.Delete()
	upsert := true
	_, err := m.client.Database(db).Collection(table).ReplaceOne(m.ctx, query, object,
		options.MergeReplaceOptions(
			&options.ReplaceOptions{Upsert: &upsert},
		),
	)
	if err != nil {
		return err
	}

	_, err = m.client.Database(db).Collection(table).DeleteOne(m.ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) Delete(db, table, name, workspace string) error {
	query := bson.M{metadataName: name}
	if workspace != "" {
		query[metadataWorkspace] = workspace
	}

	object := &core.DefaultObject{}
	singleResult := m.client.Database(db).Collection(table).FindOne(m.ctx, query)
	if singleResult.Err() == mongo.ErrNoDocuments {
		return nil
	}
	if err := singleResult.Decode(object); err != nil {
		return err
	}

	object.Delete()
	_, err := m.client.Database(db).Collection(table).ReplaceOne(m.ctx, query, object)
	if err != nil {
		return err
	}

	_, err = m.client.Database(db).Collection(table).DeleteOne(m.ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) DeleteByUUID(db, table, uuid string) error {
	query := bson.M{metadataUUID: uuid}
	_, err := m.client.Database(db).Collection(table).DeleteOne(m.ctx, query)
	if err != nil {
		return err
	}
	return nil
}
