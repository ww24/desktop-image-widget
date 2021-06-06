package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/ww24/desktop-image-widget/widget"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	w := widget.NewWidget()
	if err := w.Run(ctx); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
