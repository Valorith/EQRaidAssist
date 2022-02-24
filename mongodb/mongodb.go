package mongodb

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

/*
type database struct {
	client  *mongo.Client
	context context.Context
}
*/

func Connect() error {
	// Set client options
	fmt.Println("Configuring database connection...")
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Connect(): error reading config file:%s", err)
	}
	USERNAME := getEnvVarString("MONGODB_USERNAME")
	PASSWORD := getEnvVarString("MONGODB_PASSWORD")
	DBString := "mongodb+srv://" + USERNAME + ":" + PASSWORD + "@cluster0.t0wb9.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	client, err := mongo.NewClient(options.Client().ApplyURI(DBString))
	if err != nil {
		return fmt.Errorf("Connect(): error creating mongo client: %v", err)
	}
	// Set context
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	fmt.Println("Attempting to connect to database...")
	err = client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("Connect(): error connecting to mongo: %v", err)
	}
	fmt.Println("Connected to database.")
	defer client.Disconnect(ctx)
	fmt.Println("Testing database connetion...")
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("Connect(): error pinging mongo: %v", err)
	}
	fmt.Println("Database test succesful.")

	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("Connect(): error listing databases: %v", err)
	}
	fmt.Println("Databases:")
	for _, database := range databases {
		fmt.Println("\t", database)
	}

	return nil
}

func getEnvVarString(key string) string {
	value := viper.GetString(key)
	return value
}
