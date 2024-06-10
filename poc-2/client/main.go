package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Inicializa o watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	// Monitora mudanças no arquivo output.txt
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("Arquivo modificado:", event.Name)
					// Lê o conteúdo do arquivo
					data, err := ioutil.ReadFile("output.txt")
					if err != nil {
						log.Println("Erro ao ler o arquivo:", err)
						continue
					}
					// Envia o conteúdo do arquivo para o servidor
					err = sendToServer(data)
					if err != nil {
						log.Println("Erro ao enviar para o servidor:", err)
						continue
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Erro do watcher:", err)
			}
		}
	}()

	// Adiciona o arquivo ao watcher
	err = watcher.Add("output.txt")
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

// Envia os dados para o servidor
func sendToServer(data []byte) error {
	url := "http://localhost:8000/save"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}

	return nil
}
