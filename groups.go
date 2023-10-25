package main

import (
	"context"
	"fmt"
	"log/slog"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

//nolint:gochecknoglobals
var groupMemberIds = map[string][]string{}

func getGroupMembers(ctx context.Context, client *msgraphsdk.GraphServiceClient, groupIDs []string) ([]string, error) {
	var targetUserIDs []string

	for _, groupID := range groupIDs {
		if _, ok := groupMemberIds[groupID]; ok {
			targetUserIDs = append(targetUserIDs, groupMemberIds[groupID]...)

			continue
		}

		slog.Info(fmt.Sprintf("get transitive members of groups %s", groupID))

		transitiveMembers, err := client.Groups().ByGroupId(groupID).TransitiveMembers().Get(ctx, nil)
		if err != nil {
			return nil, err
		}

		for _, transitiveMember := range transitiveMembers.GetValue() {
			targetUserIDs = append(targetUserIDs, *transitiveMember.GetId())
		}
	}

	return targetUserIDs, nil
}
