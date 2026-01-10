package organize

import (
	"context"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateMissingSubFolders(t *testing.T) {
	subFolders := []string{"Parent", "Child", "Grandchild"}
	existingFolders := make(map[string]proton.Label)
	expectedFolders := map[string]proton.Label{
		"Parent":     {Name: "Parent"},
		"Child":      {Name: "Child"},
		"Grandchild": {Name: "Granchild"},
	}
	expectedParentIds := []string{"", "ParentID", "ChildID"}

	var parentIds []string
	mock := &ProtonClientMock{
		CreateLabelFunc: func(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error) {
			id := uuid.New().String()
			switch req.Name {
			case "Parent":
				id = "ParentID"
			case "Child":
				id = "ChildID"
			}

			parentIds = append(parentIds, req.ParentID)
			return proton.Label{
				ID:   id,
				Name: req.Name,
			}, nil
		},
	}

	ctx := context.Background()
	service := New(mock)
	service.createSubFolders(ctx, subFolders, &existingFolders)

	assert.Equal(t, 3, len(mock.CreateLabelCalls()))
	assert.Equal(t, expectedFolders["Parent"].Name, existingFolders["Parent"].Name)
	assert.Equal(t, expectedFolders["Child"].Name, existingFolders["Child"].Name)
	assert.Equal(t, expectedParentIds, parentIds)
}

// func TestCreateMissingParentFolders(t *testing.T) {
// 	mock := &ProtonClientMock{
// 		CreateLabelFunc: func(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error) {
// 			return proton.Label{
// 				Name: req.Name,
// 			}, nil
// 		},
// 	}
// 	ctx := context.Background()
// 	labels := []string{}
// 	existing
//
// 	service := New(mock)
// 	allFolders, err := service.CreateMissingFolders(ctx, subFolders, &existingFolders, []string{})
// }
