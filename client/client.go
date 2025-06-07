package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type dolar struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	url := "http://localhost:8080/cotacao"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Erro ao preparar a requisição: %v\n", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Timeout excedido ao realizar a requisição: %v\n", err)
		} else {
			fmt.Printf("Erro ao realizar a requisição: %v\n", err)
		}
		return

	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Erro ao realizar a requisição, status HTTP inválido: %d\n", resp.StatusCode)
		return
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler o retorno: %v\n", err)
		return
	}

	var cotacao dolar
	err = json.Unmarshal(res, &cotacao)
	if err != nil {
		fmt.Printf("Erro ao fazer parse do retorno: %v\n", err)
	}

	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s", cotacao.Bid)), 0644)
	if err != nil {
		fmt.Printf("Erro ao criar/atualizar arquivo cotacao.txt: %v\n", err)
	}

	fmt.Printf("Arquivo atualizado, cotação atual: Dolar: %s\n", cotacao.Bid)
}
