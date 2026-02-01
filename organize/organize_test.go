package organize

import (
	"context"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateMissingFolderWithFolderThatAlreadyExists(t *testing.T) {
	mock := &ProtonClientMock{
		CreateLabelFunc: func(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error) {
			return proton.Label{
				Name: req.Name,
			}, nil
		},
	}
	existingFolders := map[string]proton.Label{
		"Label": {Name: "Label"},
	}
	expectedFolders := existingFolders

	ctx := context.Background()
	service := New(mock)
	_, err := service.createMissingFolder(ctx, "Label", "", &existingFolders)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedFolders), len(existingFolders))
	assert.Equal(t, expectedFolders["Label"], existingFolders["Label"])
}

func TestCreateMissingFolder(t *testing.T) {
	mock := &ProtonClientMock{
		CreateLabelFunc: func(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error) {
			return proton.Label{
				Name: req.Name,
			}, nil
		},
	}
	existingFolders := make(map[string]proton.Label)
	expectedFolders := map[string]proton.Label{
		"Label": {Name: TmpFolderName + "Label"},
	}

	ctx := context.Background()
	service := New(mock)
	newFolder, err := service.createMissingFolder(ctx, "Label", "", &existingFolders)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedFolders), len(existingFolders))
	assert.Equal(t, expectedFolders["Label"], existingFolders["Label"])
	assert.Equal(t, expectedFolders["Label"].Name, newFolder.Name)
}

func TestCreateMissingSubFolders(t *testing.T) {
	subFolders := []string{"Parent", "Child", "Grandchild"}
	existingFolders := make(map[string]proton.Label)
	expectedFolders := map[string]proton.Label{
		"Parent":     {Name: TmpFolderName + "Parent"},
		"Child":      {Name: TmpFolderName + "Child"},
		"Grandchild": {Name: TmpFolderName + "Granchild"},
	}
	expectedParentIds := []string{"", "ParentID", "ChildID"}

	var parentIds []string
	mock := &ProtonClientMock{
		CreateLabelFunc: func(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error) {
			id := uuid.New().String()
			switch req.Name {
			case TmpFolderName + "Parent":
				id = "ParentID"
			case TmpFolderName + "Child":
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
	_, err := service.createSubFolders(ctx, subFolders, &existingFolders)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(mock.CreateLabelCalls()))
	assert.Equal(t, expectedFolders["Parent"].Name, existingFolders["Parent"].Name)
	assert.Equal(t, expectedFolders["Child"].Name, existingFolders["Child"].Name)
	assert.Equal(t, expectedParentIds, parentIds)
}

func TestCreateMissingFolderAndSubFolders(t *testing.T) {
	subFolders := []string{"Parent", "Child", "Grandchild"}
	existingFolders := make(map[string]proton.Label)
	expectedFolders := map[string]proton.Label{
		"Parent":     {Name: TmpFolderName + "Parent"},
		"Child":      {Name: TmpFolderName + "Child"},
		"Grandchild": {Name: TmpFolderName + "Grandchild"},
		"Label":      {Name: TmpFolderName + "Label"},
	}

	mock := &ProtonClientMock{
		CreateLabelFunc: func(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error) {
			return proton.Label{
				Name: req.Name,
			}, nil
		},
	}

	ctx := context.Background()
	service := New(mock)
	_, err := service.createMissingFolder(ctx, "Label", "", &existingFolders)
	assert.NoError(t, err)

	_, err = service.createSubFolders(ctx, subFolders, &existingFolders)
	assert.NoError(t, err)

	assert.Equal(t, len(expectedFolders), len(existingFolders))
	assert.Equal(t, expectedFolders["Parent"], existingFolders["Parent"])
	assert.Equal(t, expectedFolders["Child"], existingFolders["Child"])
	assert.Equal(t, expectedFolders["Grandchild"], existingFolders["Grandchild"])
	assert.Equal(t, expectedFolders["Label"], existingFolders["Label"])
}

// func TestCreateMissingFolders(t *testing.T) {
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
