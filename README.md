# Nosee Sonde sites web

Sonde ayant a vocation a surveiller des métrics sur un site web.
Temps de réponse, contenu du site, si le site est indexé, code HTTP.

### valeurs par défaut config :
NbRetentionsWarning = 2
NbRetentionsCritical = 1

### Options
``` text
  -d string Directory with sondes
  -t	Test mode - execute test part only
  -v	Print version
```
### Signaux écoutés
```text
USR1 : Va lire le dossier des sondes pour mettre à jour la liste des sondes.
USR2 : débug des sondes en cours avec des informations sur leurs satus
QUIT : renvoie la liste des go routines en cours
```
- kill -USR1 $(pidof go-wp-sonde)
- kill -USR2 $(pidof go-wp-sonde)

### Envs Requis
- SONDE_SLACK_WEBHOOK_URL
- SONDE_NOSEE_URL
- SONDE_NOSEE_INFLUXDB_URL