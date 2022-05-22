package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/alexmullins/zip"
	"github.com/andersfylling/snowflake"
	"github.com/nickname32/discordhook"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

var backupName = time.Now().Format("2006-01-02T150405")

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

	runMongoDump()

	webhookID := viper.GetString("webhook.id")
	webhookToken := viper.GetString("webhook.token")
	zipPassword := viper.GetString("zipPassword")

	createZipFile(zipPassword)
	sendToDiscord(webhookID, webhookToken)
}

func runMongoDump() {
	subProcess := exec.Command("mongodump", "--uri=\""+viper.GetString("mongoUri")+"\"", "--archive=backups/dump."+backupName+".archive")
	if err := subProcess.Run(); err != nil {
		fmt.Println("An error occured: ", err)
	}
}

func createZipFile(zipPassword string) {
	archive, err := os.Create("backups/" + backupName + ".zip")
	if err != nil {
		log.Fatal("an error occurred while creating the archive: ", err)
	}
	defer func(archive *os.File) {
		err = archive.Close()
		if err != nil {
			log.Println(err)
		}
	}(archive)

	zipWriter := zip.NewWriter(archive)
	writer, err := zipWriter.Encrypt("backups/dump."+backupName+".archive", zipPassword)
	if err != nil {
		log.Fatal("an error occurred while creating the archive: ", err)
	}

	file, err := os.Open("backups/dump." + backupName + ".archive")
	if err != nil {
		return
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		log.Fatal("an error occurred while creating the archive: ", err)
	}
	defer func(zipWriter *zip.Writer) {
		err = zipWriter.Close()
		if err != nil {
			log.Println(err)
		}
	}(zipWriter)
}

func sendToDiscord(webhookID string, webhookToken string) {
	api, err := discordhook.NewWebhookAPI(snowflake.ParseSnowflakeString(webhookID), webhookToken, true, nil)
	if err != nil {
		log.Fatal("an error occurred while the webhook api creation: ", err)
	}

	file, err := os.Open("backups/" + backupName + ".zip")
	if err != nil {
		return
	}

	_, err = api.Execute(context.TODO(), &discordhook.WebhookExecuteParams{
		Content: "backup: " + time.Now().Format(time.RFC3339),
	}, file, time.Now().Format(time.RFC3339)+".zip")

	if err != nil {
		log.Fatal("an error occured while executing the webhook: ", err)
	}

}
