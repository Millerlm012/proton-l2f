package organize

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ProtonMail/go-proton-api"
)

//go:generate moq -out protonclient_moq_test.go . ProtonClient
type ProtonClient interface {
	CreateLabel(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error)
	GetLabels(ctx context.Context, labelTypes ...proton.LabelType) ([]proton.Label, error)
}

type Service struct {
	Client ProtonClient
}

func New(protonClient ProtonClient) *Service {
	return &Service{
		Client: protonClient,
	}
}

func (s *Service) createSubFolders(ctx context.Context, subFolders []string, existingFolders *map[string]proton.Label) error {
	folders := *existingFolders

	for i, subFolder := range subFolders {
		// parentID is "" for first element
		// all sub folders parent will be the previous sub
		parentID := ""
		if i != 0 {
			previousFolder := subFolders[i-1]
			parentID = folders[previousFolder].ID
		}
		s.createMissingFolder(ctx, subFolder, parentID, existingFolders)
	}

	return nil
}

func (s *Service) createMissingFolder(ctx context.Context, labelName string, parentID string, existingFolders *map[string]proton.Label) error {
	folders := *existingFolders

	// skipping label that already has a matching folder
	if _, ok := folders[labelName]; ok {
		return nil
	}

	newFolder, err := s.Client.CreateLabel(ctx, proton.CreateLabelReq{
		Name:     labelName,
		Type:     proton.LabelTypeFolder,
		ParentID: parentID,
	})
	if err != nil {
		return fmt.Errorf("failed to create %s folder %w", labelName, err)
	}

	folders[newFolder.Name] = newFolder
	return nil
}

func (s *Service) CreateMissingFolders(ctx context.Context, labels []proton.Label, folders []proton.Label, labelsToIgnore []string) (map[string]proton.Label, error) {
	existingFolders := make(map[string]proton.Label)
	for _, folder := range folders {
		existingFolders[folder.Name] = folder
	}

	for _, label := range labels {
		// skipping labels that user requested to ignore
		if slices.Contains(labelsToIgnore, label.Name) {
			continue
		}

		if strings.Contains(label.Name, "-") {
			// handle each sub
			subFolders := strings.Split(label.Name, "-")
			s.createSubFolders(ctx, subFolders, &existingFolders)
			continue
		}

		s.createMissingFolder(ctx, label.Name, "", &existingFolders)
	}

	return existingFolders, nil
}

func (s *Service) MigrateLabelsToFolders(ctx context.Context, labelsToIgnore []string) error {
	existingLabels, err := s.Client.GetLabels(ctx, proton.LabelTypeLabel)
	if err != nil {
		return fmt.Errorf("failed to get labels %w", err)
	}

	existingFolders, err := s.Client.GetLabels(ctx, proton.LabelTypeFolder)
	if err != nil {
		return fmt.Errorf("failed to get folders %w", err)
	}

	s.CreateMissingFolders(ctx, existingLabels, existingFolders, labelsToIgnore)

	// for each existing label:
	// - get all emails for that label
	// - move them to the new folder
	// - remove OG label and important label
	// - delete OG label

	return nil
}
