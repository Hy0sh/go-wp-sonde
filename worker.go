package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Worker struct {
	sondes    []*Sonde
	dirSondes string
}

/**
* Initial load of sondes
 */
func (w *Worker) InitialLoadSondes() error {
	if _, err := os.Stat(w.dirSondes); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(w.dirSondes)
	if err != nil {
		return err
	}
	sondes := make([]*Sonde, 0)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		sonde, err := LoadFromToml(w.dirSondes + "/" + file.Name())
		if err != nil {
			return err
		}

		sondes = append(sondes, sonde)
	}

	w.sondes = sondes

	w.DisplaySondesList()

	return nil
}

/**
* Observe the directory for update / create /remove sondes
 */
func (w *Worker) ObserveSondeDir() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				if !strings.HasSuffix(event.Name, ".toml") {
					continue
				}

				if event.Op == fsnotify.Remove {
					for i, sonde := range w.sondes {
						if sonde.FileName == event.Name {
							w.sondes = append(w.sondes[:i], w.sondes[i+1:]...)
							break
						}
					}
					fmt.Printf("Sonde %s supprimée\n", event.Name)
					w.DisplaySondesList()
				} else {
					sonde, err := LoadFromToml(event.Name)
					if err != nil {
						fmt.Println(err)
						continue
					}
					hasBeenUpdated := false
					for _, sondeExist := range w.sondes {
						if sondeExist.FileName == event.Name {
							sondeExist.Update(sonde)
							hasBeenUpdated = true
							break
						}
					}
					if !hasBeenUpdated {
						w.sondes = append(w.sondes, sonde)
					}
					fmt.Printf("Sonde %s ajoutée ou mise à jour\n", sonde.Name)
					w.DisplaySondesList()
				}
				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	if err := watcher.Add(w.dirSondes); err != nil {
		fmt.Println("ERROR", err)
		panic(err)
	}

	<-done
}

func (w *Worker) DisplaySondesList() {
	fmt.Println("Liste des sondes surveillées :")
	for _, sonde := range w.sondes {
		fmt.Printf("%s\n", sonde.Name)
	}
}

func (w *Worker) Run() error {
	chSonde := make(chan *Sonde)
	var hashErrSonde map[string]*SondeError = make(map[string]*SondeError)
	defer close(chSonde)

	for {
		for _, sonde := range w.sondes {
			if time.Now().After(sonde.NextExecution) {
				time.Sleep(time.Millisecond * time.Duration((100 / len(w.sondes))))
				go sonde.Check(chSonde)
				go func() {
					sonde := <-chSonde
					// Détection des erreurs qui ont disparu
					for hash, oldSerr := range hashErrSonde {
						if oldSerr.FileName == sonde.FileName && oldSerr.IsErrorSolved(sonde.Errors) {
							delete(hashErrSonde, hash)
							oldSerr.DisplayResolvedError(sonde)
						}
					}

					// On ajoute les nouvelles erreurs
					for _, sondeError := range sonde.Errors {
						if hashErrSonde[sondeError.Hash] == nil {
							hashErrSonde[sondeError.Hash] = sondeError
							sondeError.DisplayNewError(sonde)
						}
					}

				}()

			}
		}

		time.Sleep(time.Second * 1)
	}
}

func NewWorker(dirSondes string) *Worker {
	return &Worker{
		dirSondes: dirSondes,
	}
}
