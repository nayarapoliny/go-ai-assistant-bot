package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nayarapoliny/go-ai-assistant-bot/internal/aibot"
)

func main() {
	cliMode := flag.Bool("cli", false, "Executar o bot em modo interativo no terminal")
	port := flag.String("port", "8081", "Porta HTTP do microsserviço do bot")
	flag.Parse()

	apiKey := os.Getenv("GEMINI_API_KEY")
	provider := aibot.NewGeminiProvider(apiKey, "gemini-2.5-flash", nil)
	memory := aibot.NewSessionMemory(12)
	botService := aibot.NewBotService(provider, memory)

	if *cliMode {
		fmt.Println("🤖 Go AI Assistant Bot (Modo Terminal). Escreva 'sair' para encerrar.")
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("\nVocê > ")
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(scanner.Text())
			if strings.EqualFold(input, "sair") || strings.EqualFold(input, "exit") {
				break
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			reply, _, err := botService.Chat(ctx, "cli-session", input)
			cancel()
			if err != nil {
				fmt.Printf("❌ Erro: %v\n", err)
				continue
			}
			fmt.Printf("🤖 Bot > %s\n", reply)
		}
		return
	}

	mux := http.NewServeMux()
	botService.RegisterRoutes(mux)

	fmt.Printf("🤖 Microsserviço Go AI Bot ativo em http://localhost:%s/api/v1/chat\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
