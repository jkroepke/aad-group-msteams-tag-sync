package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
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
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	client, err := authenticate()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	sync(ctx, client, c)
}

func authenticate() (*msgraphsdk.GraphServiceClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create credential: %w", err)
	}

	return msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{})
}

func sync(ctx context.Context, client *msgraphsdk.GraphServiceClient, config configRoot) error {
	for _, tenant := range config.Tenants {
		for _, team := range tenant.Teams {
			var msTeamsTeamIds []string

			if team.ID != "" {
				team, err := client.Teams().ByTeamId(tenant.ID).Get(ctx, nil)
				if err != nil {
					return err
				}
				msTeamsTeamIds = append(msTeamsTeamIds, *team.GetId())
			} else if team.Filter != "" {
				configuration := &graphteams.TeamsRequestBuilderGetRequestConfiguration{
					QueryParameters: &graphteams.TeamsRequestBuilderGetQueryParameters{
						Filter: &team.Filter,
						Select: []string{"id"},
					},
				}
				teams, err := client.Teams().Get(context.Background(), configuration)
				if err != nil {
					return err
				}
				teams.GetValue()
			}
		}
	}
}
