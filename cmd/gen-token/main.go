package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt"

	"github.com/sunary/emu-game/pkg"
)

func main() {
	subject := flag.String("sub", "", "JWT subject / user id (random if empty)")
	name := flag.String("name", "", "Optional name claim")
	email := flag.String("email", "", "Optional email claim")
	phone := flag.String("phone", "", "Optional phone claim")
	quizID := flag.String("quiz", "quiz-42", "Quiz ID used in sample curl commands")
	score := flag.Float64("score", 150, "Score used in sample curl commands")
	host := flag.String("host", "http://localhost:8080", "Server base URL (http)")
	wsHost := flag.String("ws", "ws://localhost:8080/ws", "WebSocket URL")
	showCurl := flag.Bool("curl", false, "Show curl commands")

	flag.Parse()

	if *subject == "" {
		*subject = randomSubject()
	}

	token, err := pkg.EncodeJWT(pkg.StandardPayload{
		StandardClaims: jwt.StandardClaims{
			Subject: *subject,
		},
		Name:  *name,
		Email: *email,
		Phone: *phone,
		Sub:   *subject,
	})
	if err != nil {
		log.Fatalf("failed to encode JWT: %v", err)
	}

	fmt.Println("Generated JWT token:")
	fmt.Printf("TOKEN='%s'\n\n", token)

	if !*showCurl {
		return
	}

	fmt.Println("# Join quiz")
	fmt.Printf("curl -i -X POST %s/user/quiz/%s/join \\\n", *host, *quizID)
	fmt.Println("  -H \"Authorization: Bearer $TOKEN\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{}'")
	fmt.Println()

	fmt.Println("# Submit quiz score")
	fmt.Printf("curl -i -X POST %s/user/quiz/%s/submit \\\n", *host, *quizID)
	fmt.Println("  -H \"Authorization: Bearer $TOKEN\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Printf("  -d '{\"score\":%v}'\n\n", *score)

	fmt.Println("# Watch websocket events (requires wscat or similar)")
	fmt.Printf("wscat -c %s\n\n", *wsHost)

	fmt.Println("# Fetch leaderboard segment")
	fmt.Printf("curl -s -X GET %s/leaderboard \\\n", *host)
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"from\":0,\"limit\":10}'")
}

func randomSubject() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("failed to generate random subject: %v", err)
	}
	return "stress-" + hex.EncodeToString(buf)
}
