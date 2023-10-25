package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"
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
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func authenticate() (*msgraphsdk.GraphServiceClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create credential: %w", err)
	}

	return msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{})
}

func syncTeams(ctx context.Context, client *msgraphsdk.GraphServiceClient, config configRoot) error {
	for _, team := range config.Teams {
		var teamIDs []string

		if team.ID != "" {
			teamID, err := getTeamsByTeamId(ctx, client, team.ID)
			if err != nil {
				return err
			}

			teamIDs = append(teamIDs, teamID)
		} else if team.Filter != "" {
			teamID, err := getTeamsByFilter(ctx, client, team.Filter)
			if err != nil {
				return err
			}

			if len(teamIDs) == 0 {
				return fmt.Errorf("no teams found with filter '%s'", team.Filter)
			}

			teamIDs = append(teamIDs, teamID...)
		} else {
			return fmt.Errorf("neither 'id' nor 'filter' defined")
		}

		for _, teamID := range teamIDs {
			for _, tag := range team.Tags {
				graphApiTeamTags, err := client.Teams().ByTeamId(teamID).Tags().Get(ctx, nil)
				if err != nil {
					return err
				}

				var tagID string

				for _, graphApiTeamTag := range graphApiTeamTags.GetValue() {
					if tag.Name == *graphApiTeamTag.GetDisplayName() {
						slog.Info(fmt.Sprintf("Tag %s in teams %s found.", tag.Name, teamID))
						tagID = *graphApiTeamTag.GetId()
						break
					}
				}

				var (
					targetUserIDs     []string
					tagUserIDs        []string
					transitiveMembers models.DirectoryObjectCollectionResponseable
				)

				for _, groupID := range tag.Groups {
					slog.Info(fmt.Sprintf("get transitive members of groups %s", groupID))
					transitiveMembers, err = client.Groups().ByGroupId(groupID).TransitiveMembers().Get(ctx, nil)
					if err != nil {
						return err
					}

					for _, transitiveMember := range transitiveMembers.GetValue() {
						targetUserIDs = append(targetUserIDs, *transitiveMember.GetId())
					}
				}

				if len(targetUserIDs) == 0 {
					return fmt.Errorf("no users found for tag %s. Empty tags are not supported", tag.Name)
				} else if len(targetUserIDs) > 25 {
					return fmt.Errorf("more then 25 users found for tag %s. A MS Teams tag only support up to 25 members", tag.Name)
				}

				if tagID == "" {
					slog.Info(fmt.Sprintf("Tag %s in teams %s not found. Creating.", tag.Name, teamID))

					teamworkTagMember := models.NewTeamworkTagMember()
					teamworkTagMember.SetUserId(&targetUserIDs[0])

					requestBody := models.NewTeamworkTag()
					requestBody.SetDisplayName(&tag.Name)
					requestBody.SetMembers([]models.TeamworkTagMemberable{teamworkTagMember})

					cratedTag, err := client.Teams().ByTeamId(teamID).Tags().Post(ctx, requestBody, nil)
					if err != nil {
						return err
					}

					tagID = *cratedTag.GetId()
				}

				requestBody := models.NewTeamworkTag()
				requestBody.SetDescription(&tag.Description)

				_, err = client.Teams().ByTeamId(teamID).Tags().ByTeamworkTagId(tagID).Patch(ctx, requestBody, nil)
				if err != nil {
					return err
				}

				tagMembers, err := client.
					Teams().ByTeamId(teamID).
					Tags().ByTeamworkTagId(tagID).
					Members().Get(ctx, nil)

				if err != nil {
					return err
				}

				for _, tagMember := range tagMembers.GetValue() {
					tagUserIDs = append(tagUserIDs, *tagMember.GetUserId())

					if slices.Contains(targetUserIDs, *tagMember.GetUserId()) {
						continue
					}

					err = client.
						Teams().ByTeamId(teamID).
						Tags().ByTeamworkTagId(tagID).
						Members().ByTeamworkTagMemberId(*tagMember.GetId()).
						Delete(ctx, nil)

					if err != nil {
						return err
					}

					slog.Info(fmt.Sprintf("Removed user %s to tag %s in teams %s", *tagMember.GetUserId(), tag.Name, teamID))
				}

				for _, targetUserID := range targetUserIDs {
					if slices.Contains(tagUserIDs, targetUserID) {
						continue
					}

					requestBody := models.NewTeamworkTagMember()
					requestBody.SetUserId(&targetUserID)

					_, err = client.
						Teams().ByTeamId(teamID).
						Tags().ByTeamworkTagId(tagID).
						Members().Post(ctx, requestBody, nil)

					if err != nil {
						return err
					}

					slog.Info(fmt.Sprintf("Adder user %s to tag %s in teams %s", targetUserID, tag.Name, teamID))
				}

				slog.Info(fmt.Sprintf("Finish sync of tag %s in teams %s", tag.Name, teamID))
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

	graphApiTeams, err := client.Teams().Get(ctx, configuration)
	if err != nil {
		return nil, err
	}

	if len(graphApiTeams.GetValue()) == 0 {
		return nil, fmt.Errorf("no teams with filter '%s' found", filter)
	}

	var teamIDs []string
	for _, graphApiTeam := range graphApiTeams.GetValue() {
		teamIDs = append(teamIDs, *graphApiTeam.GetId())
	}

	return teamIDs, nil
}

func getTeamsByTeamId(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string) (string, error) {
	graphApiTeam, err := client.Teams().ByTeamId(teamID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return *graphApiTeam.GetId(), nil
}
