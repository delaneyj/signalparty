package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/delaneyj/signalparty/cmd/codegen/templates"
	"github.com/urfave/cli/v3"
)

const (
	threadSafeKey        = "safe"
	genericParamCountKey = "count"
)

func main() {
	cmd := &cli.Command{
		Name:  "generate",
		Usage: "Generate code for 🚀 signals",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  threadSafeKey,
				Usage: "Generate thread-safe code",
				Value: true,
			},
			&cli.UintFlag{
				Name:  genericParamCountKey,
				Usage: "Number of generic parameters to generate",
				Value: 8,
			},
		},
		Action: generate,
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func generate(ctx context.Context, cmd *cli.Command) error {
	start := time.Now()
	log.Printf("Codegen for rocket started !")
	defer func() {
		log.Printf("Codegen for rocket finished in %v", time.Since(start))
	}()

	log.Printf("Thread safe: %v", cmd.Bool(threadSafeKey))

	threadSafe := cmd.Bool(threadSafeKey)
	genericParamCount := cmd.Uint(genericParamCountKey)

	contents := templates.DumbdumbGen(threadSafe, int(genericParamCount))
	if err := os.WriteFile("dumbdumb/signals.go", []byte(contents), 0644); err != nil {
		return err
	}

	contents = templates.RocketGen(threadSafe, int(genericParamCount))
	if err := os.WriteFile("rocket/signals.go", []byte(contents), 0644); err != nil {
		return err
	}

	return nil
}
