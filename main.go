package main

// An example Bubble Tea server. This will put an ssh session into alt screen
// and continually print up to date terminal information.

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	"github.com/muesli/termenv"
)

const (
	host = "localhost"
	port = 22
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			myCustomBubbleteaMiddleware(),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

// You can write your own custom bubbletea middleware that wraps tea.Program.
// Make sure you set the program input and output to ssh.Session.
func myCustomBubbleteaMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		m := InitialiseWelcomeScreen(pty)

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	return bm.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
