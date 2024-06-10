package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	// Cria um servidor HTTP
	http.HandleFunc("/save", saveHandler)
	fmt.Println("Servidor escutando em http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}

// saveHandler salva os dados recebidos no arquivo out.txt
func saveHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erro ao ler o corpo da requisição", http.StatusInternalServerError)
		return
	}

	// Escreve os dados no arquivo out.txt
	err = ioutil.WriteFile("out.txt", data, os.ModePerm)
	if err != nil {
		http.Error(w, "Erro ao salvar os dados no arquivo", http.StatusInternalServerError)
		return
	}

	fmt.Println("Dados salvos com sucesso em out.txt")
	w.WriteHeader(http.StatusOK)
}
