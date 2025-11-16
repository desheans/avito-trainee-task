package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/tests"

	"github.com/stretchr/testify/require"
)

const (
	schemeMigrationsPath = "../../../migrations/"
)

var serverURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	s, err := tests.CreatePostgresStorage(ctx, schemeMigrationsPath)
	if err != nil {
		log.Fatal(err)
	}

	var cleanup func()
	serverURL, cleanup, err = tests.StartServer(s)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	code := m.Run()
	os.Exit(code)
}

func TestTeamPRReassign(t *testing.T) {
	team := api.Team{
		TeamName: "t1",
		Members: []api.TeamMember{
			{UserId: "u1", Username: "u1", IsActive: true},
			{UserId: "u2", Username: "u2", IsActive: true},
			{UserId: "u3", Username: "u3", IsActive: true},
			{UserId: "u4", Username: "u4", IsActive: true},
		},
	}

	teamJSON, _ := json.Marshal(team)
	resp, err := http.Post(serverURL+"/team/add", "application/json", bytes.NewReader(teamJSON))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)
	defer resp.Body.Close()

	prReq := api.PostPullRequestCreateJSONBody{
		PullRequestId:   "01",
		PullRequestName: "test",
		AuthorId:        "u1",
	}

	prJSON, _ := json.Marshal(prReq)
	resp, err = http.Post(serverURL+"/pullRequest/create", "application/json", bytes.NewReader(prJSON))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)
	defer resp.Body.Close()

	var prResp map[string]api.PullRequest
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &prResp)
	require.NoError(t, err)

	pr := prResp["pr"]
	require.Len(t, pr.AssignedReviewers, 2)

	oldReviewer := pr.AssignedReviewers[0]
	reassignReq := api.PostPullRequestReassignJSONBody{
		OldUserId:     oldReviewer,
		PullRequestId: "01",
	}

	reassignJSON, _ := json.Marshal(reassignReq)
	resp, err = http.Post(serverURL+"/pullRequest/reassign", "application/json", bytes.NewReader(reassignJSON))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()

	type ReassignResponse struct {
		PR         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}

	var reassignResp ReassignResponse
	body, _ = io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &reassignResp)
	require.NoError(t, err)

	updatedPR := reassignResp.PR
	replacedBy := reassignResp.ReplacedBy

	reviewers := updatedPR.AssignedReviewers
	require.NotContains(t, reviewers, oldReviewer)
	require.Contains(t, reviewers, replacedBy)
	require.Len(t, reviewers, 2)
}
