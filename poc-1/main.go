package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Estrutura para armazenar os dados em memória
type DataStore struct {
	mu    sync.Mutex
	lines []string
}

var dataStore = DataStore{
	lines: make([]string, 0),
}

// Websocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Função para ler o arquivo em tempo real
func readFileInRealTime(filePath string, done chan bool) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		select {
		case <-done:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(1 * time.Second) // espera 1 segundo antes de tentar novamente
				continue
			}
			dataStore.mu.Lock()
			dataStore.lines = append(dataStore.lines, line)
			dataStore.mu.Unlock()
			fmt.Println("Linha lida:", line)
		}
	}
}

// Handler para WebSocket
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Falha ao atualizar conexão: %v", err)
		return
	}
	defer conn.Close()

	for {
		dataStore.mu.Lock()
		for _, line := range dataStore.lines {
			err := conn.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				log.Printf("Erro ao escrever mensagem: %v", err)
				dataStore.mu.Unlock()
				return
			}
		}
		dataStore.lines = make([]string, 0) // Limpa os dados após enviar
		dataStore.mu.Unlock()
		time.Sleep(1 * time.Second) // espera 1 segundo antes de enviar novos dados
	}
}

// Handler para servir a página HTML
func htmlHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Real-time Log Viewer</title>
</head>
<body>
    <h1>Real-time Log Viewer</h1>
    <div id="logContainer" style="white-space: pre-wrap;"></div>
    <script>
        const logContainer = document.getElementById('logContainer');
        const socket = new WebSocket('ws://' + window.location.host + '/ws');
        socket.onmessage = function(event) {
            logContainer.innerHTML += event.data;
        };
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}

func main() {
	// Caminho para o arquivo a ser lido
	filePath := "output.txt"
	done := make(chan bool)
	go readFileInRealTime(filePath, done)

	http.HandleFunc("/", htmlHandler)
	http.HandleFunc("/ws", wsHandler)

	log.Println("Servidor iniciado em :4000")
	log.Fatal(http.ListenAndServe(":4000", nil))
}
