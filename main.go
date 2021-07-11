package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/andersfylling/snowflake"
	"github.com/nickname32/discordhook"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var backupData string

func main() {
	// loading configuration file path
	configPath := flag.String("config", "/etc/mongobackup.yml", "path to configuration file")
	flag.Parse()
	log.Println("loading configuration file from " + *configPath)

	// loading configuration
	viper.SetConfigFile(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("an error occurred while reading the configuration: ", err)
	}

	// opening the log file
	logPath := viper.GetString("logPath")
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("an error occurred while opening the log file: ", err)
	}
	// formatting log output
	log.SetFlags(log.Ltime)
	log.SetOutput(io.MultiWriter(file, os.Stdout))

	// prepare mongo context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// create a mongo client
	mongoUri := viper.GetString("mongoUri")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal("an error occurred while creating the mongo client: ", err)
	}

	databaseNames, err := client.ListDatabaseNames(ctx, bson.D{}, options.ListDatabases())
	if err != nil {
		log.Fatal("an error occurred while reading databases names: ", err)
	}

	for _, databaseName := range databaseNames {
		database := client.Database(databaseName)
		collectionNames, err := database.ListCollectionNames(ctx, bson.D{}, options.ListCollections())
		if err != nil {
			log.Fatal("an error occurred while reading collection names of "+databaseName+" database: ", err)
		}
		backupData += "// database " + databaseName + "\n"
		for _, collectionName := range collectionNames {
			collection := database.Collection(collectionName)
			cursor, err := collection.Find(ctx, bson.D{}, options.Find())
			if err != nil {
				log.Fatal("an error occurred while reading documents of "+collectionName+" collection: ", err)
			}
			backupData += "// collection " + collectionName + "\n"
			backupCollection(cursor, ctx)
		}
	}
	webhookID := viper.GetString("webhook.id")
	webhookToken := viper.GetString("webhook.token")
	sendToDiscord(webhookID, webhookToken)
}

func backupCollection(cursor *mongo.Cursor, ctx context.Context) {
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatal("an error occurred while closing mongo cursor: ", err)
		}
	}(cursor, ctx)
	for cursor.Next(ctx) {
		var document bson.M
		err := cursor.Decode(&document)
		if err != nil {
			log.Fatal("an error occurred while decoding a mongo document: ", err)
		}
		marshal, err := json.Marshal(document)
		if err != nil {
			return
		}
		backupData += string(marshal) + "\n"
	}
	if err := cursor.Err(); err != nil {
		log.Fatal("an error occured with the mongo cursor: ", err)
	}
}

func sendToDiscord(webhookID string, webhookToken string) {
	api, err := discordhook.NewWebhookAPI(snowflake.ParseSnowflakeString(webhookID), webhookToken, true, nil)
	if err != nil {
		log.Fatal("an error occured while the webhook api creation: ", err)
	}

	_, err = api.Execute(context.TODO(), &discordhook.WebhookExecuteParams{
		Content: "backup: " + time.Now().String(),
	}, strings.NewReader(backupData), "backup.json")

	if err != nil {
		log.Fatal("an error occured while executing the webhook: ", err)
	}

}
