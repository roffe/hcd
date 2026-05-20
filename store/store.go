package store

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/roffe/hcd/ent"
	"github.com/roffe/hcd/ent/car"
	"github.com/roffe/hcd/ent/valuehistory"
)

type Store struct {
	c *ent.Client
}

func New() (*Store, error) {
	client, err := ent.Open("sqlite3", "file:db.sqlite?cache=shared&_fk=1")
	if err != nil {
		//log.Fatalf("failed opening connection to sqlite: %v", err)
		return nil, fmt.Errorf("failed opening connection to sqlite: %w", err)
	}
	//defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(
		context.Background(),
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
	); err != nil {
		//log.Fatalf("failed creating schema resources: %v", err)
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	return &Store{c: client}, nil
}

func (s *Store) Close() error {
	return s.c.Close()
}

func (s *Store) CreateCar(ctx context.Context, vin string) (*ent.Car, error) {
	return s.c.Car.Create().SetVIN(vin).Save(ctx)
}

func (s *Store) GetCar(ctx context.Context, vin string) (*ent.Car, error) {
	return s.c.Car.Query().Where(car.VIN(vin)).WithValueHistories().Only(ctx)
}

func (s *Store) GetCars(ctx context.Context) ([]*ent.Car, error) {
	return s.c.Car.Query().Order(ent.Desc(car.FieldLastAccessed)).All(ctx)
}

func (s *Store) CreateValueHistoryRead(ctx context.Context, vin string, uec, uecPWM, rec []byte) (*ent.ValueHistory, error) {
	return s.CreateValueHistory(ctx, vin, uec, uecPWM, rec, valuehistory.TypeREAD)
}

func (s *Store) CreateValueHistoryWrite(ctx context.Context, vin string, uec, uecPWM, rec []byte) (*ent.ValueHistory, error) {
	return s.CreateValueHistory(ctx, vin, uec, uecPWM, rec, valuehistory.TypeWRITE)
}

func (s *Store) CreateValueHistory(ctx context.Context, vin string, uec, uecPWM, rec []byte, typ valuehistory.Type) (*ent.ValueHistory, error) {
	// Select existing car or create new one
	car, err := s.c.Car.Query().Where(car.VIN(vin)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			car, err = s.CreateCar(ctx, vin)
			if err != nil {
				return nil, fmt.Errorf("failed creating car: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed querying car: %w", err)
		}
	}

	if err := car.Update().SetLastAccessed(time.Now()).Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed updating car last accessed: %w", err)
	}

	return s.c.ValueHistory.Create().
		SetCar(car).
		SetType(typ).
		SetUEC(uec).
		SetUECPWM(uecPWM).
		SetREC(rec).
		Save(ctx)
}

func (s *Store) GetValueHistory(ctx context.Context, vin string) ([]*ent.ValueHistory, error) {
	car, err := s.c.Car.Query().Where(car.VIN(vin)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying car: %w", err)
	}

	return car.QueryValueHistories().Order(ent.Desc(valuehistory.FieldCreatedAt)).All(ctx)
}

func (s *Store) UpdateCarName(ctx context.Context, vin string, name string) error {
	car, err := s.c.Car.Query().Where(car.VIN(vin)).Only(ctx)
	if err != nil {
		return fmt.Errorf("failed querying car: %w", err)
	}

	return s.c.Car.UpdateOne(car).SetName(name).Exec(ctx)
}
