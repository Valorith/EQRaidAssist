package mongodb

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

var ( // Databases
	// Database storing alias data
	AliasDB = database{}
	RaidsDB = database{}
)

// Represents data associated with a single mongodb connection
type database struct {
	Client         *mongo.Client     `json:"client"`
	Context        context.Context   `json:"context"`
	Username       string            `json:"username"`
	Password       string            `json:"password"`
	ClusterName    string            `json:"clusterName"`
	DatabaseName   string            `json:"databaseName"`
	CollectionName string            `json:"collectionName"`
	Collection     *mongo.Collection `json:"collection"`
	DbString       string            `json:"dbString"`
	Connected      bool              `json:"connected"`
}

func Init() {
	// Connect to alias database
	AliasDB.ClusterName = "cluster0"
	AliasDB.DatabaseName = "CWRaidAssist"
	AliasDB.CollectionName = "aliases"
	err := AliasDB.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
	}

	// Connect to raids database
	RaidsDB.ClusterName = "cluster0"
	RaidsDB.DatabaseName = "CWRaidAssist"
	RaidsDB.CollectionName = "raids"
	err = RaidsDB.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
	}

}

func DisconnectALL() {
	err := AliasDB.Disconnect()
	if err != nil {
		fmt.Println("Error disconnecting from alias database:", err)
	}
	err = RaidsDB.Disconnect()
	if err != nil {
		fmt.Println("Error disconnecting from raids database:", err)
	}
}

func (db *database) Connect() error {
	// Set client options
	fmt.Println("Configuring database connection...")
	// Load .env
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Connect(): error reading config file:%s", err)
	}
	// If credentials are not set, attempt to load them from .env
	if db.Username == "" {
		db.Username = getEnvVarString("MONGODB_USERNAME")
		db.Password = getEnvVarString("MONGODB_PASSWORD")
	}

	// Set dbString
	db.DbString = "mongodb+srv://" + db.Username + ":" + db.Password + "@" + db.ClusterName + ".t0wb9.mongodb.net/" + db.DatabaseName + "?retryWrites=true&w=majority"

	// Set client
	db.Client, err = mongo.NewClient(options.Client().ApplyURI(db.DbString))
	if err != nil {
		return fmt.Errorf("Connect(): error creating mongo client: %v", err)
	}

	// Set context
	db.Context, _ = context.WithTimeout(context.Background(), 10*time.Second)

	//Set Collection
	db.Collection = db.Client.Database(db.DatabaseName).Collection(db.CollectionName)

	// Connect to MongoDB
	fmt.Printf("Attempting to connect to database(%s\\%s\\%s)...\n", db.ClusterName, db.DatabaseName, db.CollectionName)
	err = db.Client.Connect(db.Context)
	if err != nil {
		return fmt.Errorf("Connect(): error connecting to mongo: %v", err)
	}
	fmt.Println("Connected to database.")

	// Test connection
	fmt.Println("Testing database connetion...")
	err = db.Client.Ping(db.Context, nil)
	if err != nil {
		return fmt.Errorf("Connect(): error pinging mongo: %v", err)
	}
	fmt.Println("Database test succesful.")
	db.Connected = true
	return nil
}

func (db *database) Disconnect() error {
	defer db.Client.Disconnect(db.Context)
	db.Connected = false
	return nil
}

// Inserts a struct of data into the database
func (db *database) Insert(data interface{}) error {
	if db.Connected {
		result, err := db.Collection.InsertOne(db.Context, data)
		if err != nil {
			return fmt.Errorf("Insert(): error inserting data: %v", err)
		}
		fmt.Printf("Inserted data into database(%s\\%s\\%s)...Result ID: %s...\n", db.ClusterName, db.DatabaseName, db.CollectionName, result.InsertedID)
		return nil
	}
	return fmt.Errorf("Insert(): database not connected")
}

func getEnvVarString(key string) string {
	value := viper.GetString(key)
	return value
}
