# MongoBackup 🍃
Fast and efficient tool to backup Mongo databases via a Discord webhook.

### It's fast ⚡
Totally written in Go, you benefit from all the advantages of these languages, optimized for performance. This program also uses the official Mongo drivers.

<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/2/23/Go_Logo_Aqua.svg/1200px-Go_Logo_Aqua.svg.png" height=40 width=auto>

### It's simple 🤓
All you have to do is download the application binaries, fill in a configuration file and create a cron job. Simple, isn't it?
```
curl https://github.com/Romitou/MongoBackup/releases/download/latest/mongobackup-linux -o mongobackup 
```
```yml
# Where the logs are located
logPath: "./logs.log"

# The Mongo URI used to connect to the database
mongoUri: "mongodb://127.0.0.1/"

# The archive password
zipPassword: "supersecretpassword"

# Example: https://discord.com/api/webhooks/763413460892641305/gFx3ERX1IVmyVJ2etAXfQ8OVIYARz06JUmXzJOW8Z5ALv4-GE5lkW
webhook:
  id: "763413460892641305"
  token: "gFx3ERX1IVmyVJ2etAXfQ8OVIYARz06JUmXzJOW8Z5ALv4-GE5lkW"
```

### It's secure 🔒
When storing or uploading to Discord, the archive is encrypted with a password you set in the configuration. Only those who know the password can read the archive containing the backup.
