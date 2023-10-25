package main

import (
	"context"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"
)

func syncTeams(ctx context.Context, client *msgraphsdk.GraphServiceClient, config configRoot) error {
	for _, team := range config.Teams {
		var teamIDs []string

		switch {
		case team.ID != "":
			teamID, err := getTeamsByTeamID(ctx, client, team.ID)
			if err != nil {
				return err
			}

			teamIDs = append(teamIDs, teamID)
		case team.Filter != "":
			teamID, err := getTeamsByFilter(ctx, client, team.Filter)
			if err != nil {
				return err
			}

			teamIDs = append(teamIDs, teamID...)
		default:
			return fmt.Errorf("neither 'id' nor 'filter' defined")
		}

		for _, teamID := range teamIDs {
			for _, tag := range team.Tags {
				if err := syncTeamsTag(ctx, client, teamID, tag); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getTeamsByFilter(ctx context.Context, client *msgraphsdk.GraphServiceClient, filter string) ([]string, error) {
	configuration := &graphteams.TeamsRequestBuilderGetRequestConfiguration{
		QueryParameters: &graphteams.TeamsRequestBuilderGetQueryParameters{
			Filter: &filter,
			Select: []string{"id"},
		},
	}

	teams, err := client.Teams().Get(ctx, configuration)
	if err != nil {
		return nil, err
	}

	if len(teams.GetValue()) == 0 {
		return nil, fmt.Errorf("no teams with filter '%s' found", filter)
	}

	teamIDs := make([]string, len(teams.GetValue()))

	for i, team := range teams.GetValue() {
		teamIDs[i] = *team.GetId()
	}

	return teamIDs, nil
}

func getTeamsByTeamID(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string) (string, error) {
	team, err := client.Teams().ByTeamId(teamID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return *team.GetId(), nil
}
