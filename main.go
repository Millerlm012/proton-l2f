package main

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/millerlm012/proton-l2f/organize"
	"github.com/millerlm012/proton-l2f/proton"
	"github.com/millerlm012/proton-l2f/utils"
)

func loadEnv(envPath string) error {
	if err := godotenv.Load(envPath); err != nil {
		return err
	}

	return nil
}

// TODO: replace all panics with something more reasonable
func main() {
	args, err := utils.ParseFlags()
	if err != nil {
		panic(err)
	}

	if err := loadEnv(args.EnvPath); err != nil {
		panic(err)
	}

	env, err := utils.GetEnv()
	if err != nil {
		panic(err)
	}

	client, err := proton.CreateClient(env.Username, env.Password)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	orgService := organize.New(client)
	orgService.MigrateLabelsToFolders(ctx, args.IgnoreLabels)
}
