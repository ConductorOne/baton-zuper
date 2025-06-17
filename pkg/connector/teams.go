package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

// Entitlement value representing team membership.
const (
	entitlementTeamMember = "member"
)

type teamsClientInterface interface {
	GetTeams(ctx context.Context, options client.PageOptions) ([]*client.Team, string, annotations.Annotations, error)
	GetTeamUsers(ctx context.Context, teamID string) ([]*client.ZuperUser, string, annotations.Annotations, error)
}

// teamBuilder is a builder for team resources.
type teamBuilder struct {
	resourceType *v2.ResourceType
	client       teamsClientInterface
}

func (t *teamBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return teamResourceType
}

// parseIntoTeamResource converts a Team into a Baton v2.Resource.
func parseIntoTeamResource(team *client.Team) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"team_name":        team.TeamName,
		"team_color":       team.TeamColor,
		"team_description": team.TeamDescription,
		"team_timezone":    team.TeamTimezone,
		"user_count":       team.UserCount,
		"is_active":        team.IsActive,
		"created_at":       team.CreatedAt,
		"updated_at":       team.UpdatedAt,
	}
	return resource.NewGroupResource(
		team.TeamName,
		teamResourceType,
		team.TeamUID,
		[]resource.GroupTraitOption{
			resource.WithGroupProfile(profile),
		},
	)
}

// List returns the teams as Baton resources, with pagination.
func (t *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource
	bag, pageToken, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: teamResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}
	teams, nextPageToken, annos, err := t.client.GetTeams(ctx, client.PageOptions{
		PageSize:  pToken.Size,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", nil, err
	}
	for _, team := range teams {
		teamResource, err := parseIntoTeamResource(team)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, teamResource)
	}
	var outToken string
	if nextPageToken != "" {
		outToken, err = bag.NextToken(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}
	}
	return resources, outToken, annos, nil
}

// Entitlements returns a "member" entitlement for each team, grantable to users.
func (t *teamBuilder) Entitlements(ctx context.Context, teamResource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	ent := entitlement.NewAssignmentEntitlement(
		teamResource,
		entitlementTeamMember,
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDisplayName(fmt.Sprintf("Member of %s", teamResource.DisplayName)),
		entitlement.WithDescription(fmt.Sprintf("Member of team %s", teamResource.DisplayName)),
	)
	return []*v2.Entitlement{ent}, "", annos, nil
}

// Grants returns grants for the "member" entitlement for each user in the team.
func (t *teamBuilder) Grants(ctx context.Context, teamResource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	teamID := teamResource.Id.Resource
	users, _, _, err := t.client.GetTeamUsers(ctx, teamID)
	if err != nil {
		return nil, "", annos, fmt.Errorf("failed to get team users for %s: %w", teamID, err)
	}
	var grants []*v2.Grant
	for _, user := range users {
		userResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     user.UserUID,
			},
		}
		grantObj := grant.NewGrant(
			teamResource,
			entitlementTeamMember,
			userResource.Id,
			grant.WithGrantMetadata(map[string]interface{}{
				"team_id":   teamID,
				"team_name": teamResource.DisplayName,
				"user_id":   user.UserUID,
				"username":  user.Email,
			}),
		)
		grants = append(grants, grantObj)
	}
	return grants, "", annos, nil
}

// newTeamBuilder creates a new instance of teamBuilder.
func newTeamBuilder(client *client.Client) *teamBuilder {
	return &teamBuilder{
		resourceType: teamResourceType,
		client:       client,
	}
}
