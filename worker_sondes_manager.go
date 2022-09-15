package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

func (w *Worker) AppendSonde(sonde *Sonde) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.sondes[sonde.FileName] = sonde
}

func (w *Worker) RemoveSonde(filename string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.sondes, filename)
}

/**
 * Load a sonde from a file
 */
func LoadFromToml(fileSonde string) (*Sonde, error) {
	var sonde *Sonde
	_, err := toml.DecodeFile(fileSonde, &sonde)

	if err != nil {
		return sonde, err
	}

	sonde.FileName = fileSonde
	sonde.NextExecution = time.Now()
	sonde.Errors = make(map[SondeErrorStatus]*SondeError)

	return sonde, err
}

/*
* Affiche la liste des sondes chargées
 */
func (w *Worker) DisplaySondesList() {
	fmt.Println("Liste des sondes surveillées :")
	for _, sonde := range w.sondes {
		fmt.Printf("%s\n", sonde.Name)
	}
}

/**
* Initial load of sondes
 */
func (w *Worker) InitialLoadSondes() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, err := os.Stat(w.dirSondes); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(w.dirSondes)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		sonde, err := LoadFromToml(w.dirSondes + "/" + file.Name())
		if err != nil {
			return err
		}
		w.sondes[sonde.FileName] = sonde
	}

	w.DisplaySondesList()

	return nil
}

/**
* Observe le dossier des sondes pour détecter les ajouts et suppressions de fichiers
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
					w.RemoveSonde(event.Name)
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
							w.mutex.Lock()
							sondeExist.Update(sonde)
							w.mutex.Unlock()
							hasBeenUpdated = true
							break
						}
					}
					if !hasBeenUpdated {
						w.AppendSonde(sonde)
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