package mongo

import (
	"context"
	"errors"
	"fmt"

	"BulkaVPN/client/internal"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto"
	"BulkaVPN/pkg/errx"
	"BulkaVPN/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientConfig struct {
	DBName     string `envconfig:"BULKAVPN_MONGO_DBNAME" required:"true" default:"bulkavpn"`
	Collection string `envconfig:"BULKAVPN_MONGO_COLL" required:"true" default:"clients"`
}

type clientRepo struct {
	dbName   string
	collName string

	collection *mongo.Collection
	client     mongox.Client
}

func NewClientRepo(cfg ClientConfig, client mongox.Client) repository.ClientRepo {
	return &clientRepo{
		dbName:     cfg.DBName,
		collName:   cfg.Collection,
		client:     client,
		collection: client.Database(cfg.DBName).Collection(cfg.Collection),
	}
}

func (r *clientRepo) buildFilter(filter *repository.ClientSearchOpts) (bson.D, error) {
	f := bson.D{}

	if filter == nil {
		return f, nil
	}

	if len(filter.Filter.ClientId) > 0 {
		f = append(f, bson.E{Key: "client_id", Value: bson.M{"$in": filter.Filter.ClientId}})
	}

	if len(filter.Filter.OvpnConfig) > 0 {
		f = append(f, bson.E{Key: "ovpn_config", Value: bson.M{"$in": filter.Filter.OvpnConfig}})
	}

	if len(filter.Filter.CountryServer) > 0 {
		f = append(f, bson.E{Key: "country_server", Value: bson.M{"$in": []string{filter.Filter.CountryServer}}})
	}

	if filter.Filter.TelegramId > 0 {
		f = append(f, bson.E{Key: "telegram_id", Value: bson.M{"$in": filter.Filter.TelegramId}})
	}

	if filter.Filter.HasTrialBeenUsed {
		f = append(f, bson.E{Key: "has_trial_been_used", Value: bson.M{"$in": filter.Filter.HasTrialBeenUsed}})
	}

	if filter.Filter.IsTrialActiveNow {
		f = append(f, bson.E{Key: "is_trial_active_now", Value: bson.M{"$in": filter.Filter.IsTrialActiveNow}})
	}

	return f, nil
}

func (r *clientRepo) Create(ctx context.Context, client *internal.Client) error {
	_, err := r.collection.InsertOne(ctx, client)
	if err != nil {
		return fmt.Errorf("client.Create: failed: %w", err)
	}

	return nil
}

func (r *clientRepo) Get(ctx context.Context, opts repository.ClientGetOpts) (*internal.Client, error) {
	f := bson.D{}

	if len(opts.ClientID) > 0 {
		f = append(f, bson.E{Key: "client_id", Value: bson.M{"$in": []string{opts.ClientID}}})
	}

	if len(opts.OvpnConfig) > 0 {
		f = append(f, bson.E{Key: "ovpn_config", Value: bson.M{"$in": []string{opts.OvpnConfig}}})
	}

	if opts.TelegramID > 0 {
		f = append(f, bson.E{Key: "telegram_id", Value: bson.M{"$in": []int64{opts.TelegramID}}})
	}

	var c internal.Client
	if err := r.collection.FindOne(ctx, f).Decode(&c); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errx.NotFound(pb.ErrNotFound, "client.Get: client not found")
		}

		return nil, err
	}

	return &c, nil
}

func (r *clientRepo) Search(ctx context.Context, opts repository.ClientSearchOpts) ([]*internal.Client, error) {
	f, err := r.buildFilter(&opts)
	if err != nil {
		return nil, fmt.Errorf("client.Search: failed to build filter: %w", err)
	}

	o := options.Find()

	if len(opts.AfterID) > 0 {
		objectID, err := primitive.ObjectIDFromHex(opts.AfterID)
		if err != nil {
			return nil, fmt.Errorf("invalid ObjectID: %v", err)
		}
		f = append(f, bson.E{Key: "_id", Value: bson.M{"$gt": objectID}})
	}

	cur, err := r.collection.Find(ctx, f, o)
	if err != nil {
		return nil, fmt.Errorf("client.Search: failed: %w", err)
	}

	var res []*internal.Client

	if err = cur.All(ctx, &res); err != nil {
		return nil, fmt.Errorf("client.Search failed: %w", err)
	}

	return res, nil
}

func (r *clientRepo) Update(ctx context.Context, client *internal.Client, versionCheck int64) error {
	filter := bson.M{
		"client_id":   client.ClientID,
		"ver":         versionCheck - 1,
		"telegram_id": client.TelegramID,
	}
	update := bson.M{
		"$set": bson.M{
			"ovpn_config":         client.OvpnConfig,
			"country_server":      client.CountryServer,
			"ver":                 client.Ver,
			"telegram_id":         client.TelegramID,
			"time_left":           client.TimeLeft,
			"has_trial_been_used": client.HasTrialBeenUsed,
			"is_trial_active_now": client.IsTrialActiveNow,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("client.Update: failed: %w", err)
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("client.Update: no client matched for update")
	}

	return nil
}

func (r *clientRepo) Delete(ctx context.Context, clientID string) error {
	filter := bson.M{"client_id": clientID}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("clientRepo.Delete: failed to delete client: %w", err)
	}
	return nil
}

func (r *clientRepo) Count(ctx context.Context, opts repository.ClientSearchOpts) (int64, error) {
	f, err := r.buildFilter(&opts)
	if err != nil {
		return 0, fmt.Errorf("Client.Count: failed to build filter: %w", err)
	}

	count, err := r.collection.CountDocuments(ctx, f)
	if err != nil {
		return 0, fmt.Errorf("Client.Count: failed: %w", err)
	}

	return count, nil
}
