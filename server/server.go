package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type Dolar struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Cotacao struct {
	USDBRL Dolar `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite", "./meubanco.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Cria a tabela se não existir
	createTable := `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cotacao TEXT NOT NULL
	);`

	if _, err := db.Exec(createTable); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		bid, err := buscarCotacao()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = salvarCotacao(db, bid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response, _ := json.Marshal(map[string]interface{}{
			"bid": bid})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)

	})
	http.ListenAndServe(":8080", mux)
}

func salvarCotacao(db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	query := `INSERT INTO cotacoes (cotacao) VALUES (?)`

	_, err := db.ExecContext(ctx, query, bid)
	return err
}

func buscarCotacao() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("Erro ao preparar a requisição: %v\n", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msgError := ""
		if ctx.Err() == context.DeadlineExceeded {
			msgError = fmt.Sprintf("Timeout excedido ao realizar a requisição: %v\n", err)
		} else {
			msgError = fmt.Sprintf("Erro ao realizar a requisição: %v\n", err)
		}
		return "", errors.New(msgError)

	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {

		return "", fmt.Errorf("Erro ao realizar a requisição, status HTTP inválido: %d\n", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Erro ao ler o retorno: %v\n", err)
	}

	var resultado Cotacao
	err = json.Unmarshal(data, &resultado)
	if err != nil {
		return "", fmt.Errorf("Erro ao fazer parse do retorno: %v\n", err)
	}

	return resultado.USDBRL.Bid, nil
}
