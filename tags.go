package main

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func syncTeamsTag(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string, tag TagConfigStruct) error {
	targetUserIDs, err := getGroupMembers(ctx, client, tag.Groups)
	if err != nil {
		return err
	}

	if len(targetUserIDs) == 0 {
		return fmt.Errorf("no users found for tag %s. Empty tags are not supported", tag.Name)
	} else if len(targetUserIDs) > 25 {
		return fmt.Errorf("more then 25 users found for tag %s. A MS Teams tag only support up to 25 members", tag.Name)
	}

	tagID, err := findTeamsTagByDisplayName(ctx, client, teamID, tag.Name)
	if err != nil {
		return err
	}

	if tagID == "" {
		tagID, err = createTeamsTag(ctx, client, teamID, tag, targetUserIDs)
		if err != nil {
			return err
		}
	} else {
		err = updateTeamsTag(ctx, client, teamID, tagID, tag)
		if err != nil {
			return err
		}
	}

	tagMembers, err := getTeamsTagMembers(ctx, client, teamID, tagID)
	if err != nil {
		return err
	}

	err = syncTeamsTagMembers(ctx, client, teamID, tagID, tag, tagMembers, targetUserIDs)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Finish sync of tag %s in teams %s", tag.Name, teamID))

	return nil
}

func getTeamsTagMembers(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string, tagID string,
) (models.TeamworkTagMemberCollectionResponseable, error) {
	return client.Teams().ByTeamId(teamID).Tags().ByTeamworkTagId(tagID).Members().Get(ctx, nil)
}

func syncTeamsTagMembers(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string, tagID string, tag TagConfigStruct,
	tagMembers models.TeamworkTagMemberCollectionResponseable, targetUserIDs []string,
) error {
	tagUserIDs := make([]string, len(tagMembers.GetValue()))

	for i, tagMember := range tagMembers.GetValue() {
		tagUserIDs[i] = *tagMember.GetUserId()

		if slices.Contains(targetUserIDs, *tagMember.GetUserId()) {
			continue
		}

		err := client.
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
		targetUserID := targetUserID

		if slices.Contains(tagUserIDs, targetUserID) {
			continue
		}

		requestBody := models.NewTeamworkTagMember()
		requestBody.SetUserId(&targetUserID)

		_, err := client.
			Teams().ByTeamId(teamID).
			Tags().ByTeamworkTagId(tagID).
			Members().Post(ctx, requestBody, nil)
		if err != nil {
			return err
		}

		slog.Info(fmt.Sprintf("Adder user %s to tag %s in teams %s", targetUserID, tag.Name, teamID))
	}

	return nil
}

func findTeamsTagByDisplayName(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string, tagDisplayName string) (string, error) {
	existingTags, err := client.Teams().ByTeamId(teamID).Tags().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	for _, existingTag := range existingTags.GetValue() {
		if tagDisplayName == *existingTag.GetDisplayName() {
			slog.Info(fmt.Sprintf("Tag %s in teams %s found.", tagDisplayName, teamID))

			return *existingTag.GetId(), nil
		}
	}

	return "", nil
}

func createTeamsTag(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string, tag TagConfigStruct, targetUserIDs []string) (string, error) {
	slog.Info(fmt.Sprintf("Tag %s in teams %s not found. Creating.", tag.Name, teamID))

	teamworkTagMember := models.NewTeamworkTagMember()
	teamworkTagMember.SetUserId(&targetUserIDs[0])

	requestBody := models.NewTeamworkTag()
	requestBody.SetDisplayName(&tag.Name)
	requestBody.SetDescription(&tag.Description)
	requestBody.SetMembers([]models.TeamworkTagMemberable{teamworkTagMember})

	slog.Info(fmt.Sprintf("Adder user %s to tag %s in teams %s", targetUserIDs[0], tag.Name, teamID))

	cratedTag, err := client.Teams().ByTeamId(teamID).Tags().Post(ctx, requestBody, nil)
	if err != nil {
		return "", err
	}

	return *cratedTag.GetId(), nil
}

func updateTeamsTag(ctx context.Context, client *msgraphsdk.GraphServiceClient, teamID string, tagID string, tag TagConfigStruct) error {
	slog.Info(fmt.Sprintf("Updating tag %s in teams %s.", tag.Name, teamID))

	requestBody := models.NewTeamworkTag()
	requestBody.SetDisplayName(&tag.Name)
	requestBody.SetDescription(&tag.Description)

	_, err := client.Teams().ByTeamId(teamID).Tags().ByTeamworkTagId(tagID).Patch(ctx, requestBody, nil)

	return err
}
