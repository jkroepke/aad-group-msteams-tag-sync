package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"gopkg.in/yaml.v3"
)

func main() {
	ctx := context.Background()

	configFile := flag.String("config", os.Getenv("CONFIG"), "path to config file")
	flag.Parse()

	yamlFile, err := os.ReadFile(*configFile)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	var c configRoot
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	client, err := authenticate()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = syncTeams(ctx, client, c)
	if err != nil {
		slog.Error(GetOdataError(err).Error())
		os.Exit(1)
	}
}

func authenticate() (*msgraphsdk.GraphServiceClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create credential: %w", err)
	}

	return msgraphsdk.NewGraphServiceClientWithCredentials(cred, nil)
}
