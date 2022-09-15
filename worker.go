package main

import (
	"log"
	"os"
	"sync"
	"time"
)

type Worker struct {
	sondes    map[string]*Sonde
	errors    map[string]*SondeError
	dirSondes string
	mutex     *sync.Mutex
}

func (w *Worker) CheckRequiredEnv() {
	requieredEnv := []string{
		"SONDE_SLACK_WEBHOOK_URL",
		"SONDE_NOSEE_URL",
	}
	missingEnv := []string{}
	for _, env := range requieredEnv {
		if os.Getenv(env) == "" {
			missingEnv = append(missingEnv, env)
		}
	}
	if len(missingEnv) > 0 {
		log.Fatalf("Missing required env vars: %s", missingEnv)
	}
}

/*
* Lance le check sur toutes les sondes
 */
func (w *Worker) RunAllCheck() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, sonde := range w.sondes {
		if time.Now().After(sonde.NextExecution) {
			time.Sleep(time.Millisecond * time.Duration((100 / len(w.sondes))))
			go sonde.CheckAll()
		}
	}
}

/*
* Point d'entrée du worker
 */
func (w *Worker) Run() error {

	for {
		w.RunAllCheck()
		time.Sleep(time.Second * 1)
	}
}

func NewWorker(dirSondes string) *Worker {
	return &Worker{
		dirSondes: dirSondes,
		mutex:     &sync.Mutex{},
		sondes:    make(map[string]*Sonde),
		errors:    make(map[string]*SondeError),
	}
}
