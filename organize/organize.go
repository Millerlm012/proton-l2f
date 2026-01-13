package organize

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ProtonMail/go-proton-api"
	"github.com/go-resty/resty/v2"
)

//go:generate moq -out protonclient_moq_test.go . ProtonClient
type ProtonClient interface {
	CreateLabel(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error)
	GetLabels(ctx context.Context, labelTypes ...proton.LabelType) ([]proton.Label, error)
	UnlabelMessages(ctx context.Context, messageIDs []string, labelID string) error
	LabelMessages(ctx context.Context, messageIDs []string, labelID string) error
	DeleteLabel(ctx context.Context, labelID string) error
	UpdateLabel(ctx context.Context, labelID string, req proton.UpdateLabelReq) (proton.Label, error)
	do(ctx context.Context, fn func(*resty.Request) (*resty.Response, error)) error
}

type Service struct {
	Client ProtonClient
}

func New(protonClient ProtonClient) *Service {
	return &Service{
		Client: protonClient,
	}
}

func FlattenToMessageIds(messages []proton.Message) []string {
	var messageIds []string
	for _, message := range messages {
		messageIds = append(messageIds, message.ID)
	}

	return messageIds
}

func (s *Service) createSubFolders(ctx context.Context, subFolders []string, existingFolders *map[string]proton.Label) (proton.Label, error) {
	folders := *existingFolders

	var newFolder proton.Label
	for i, subFolder := range subFolders {
		// parentID is "" for first element
		// all sub folders parent will be the previous sub
		parentID := ""
		if i != 0 {
			previousFolder := subFolders[i-1]
			parentID = folders[previousFolder].ID
		}

		newlyCreated, err := s.createMissingFolder(ctx, subFolder, parentID, existingFolders)
		if err != nil {
			return proton.Label{}, err
		}

		if i+1 == len(subFolders) {
			newFolder = newlyCreated
		}
	}

	return newFolder, nil
}

func (s *Service) createMissingFolder(ctx context.Context, labelName string, parentID string, existingFolders *map[string]proton.Label) (proton.Label, error) {
	folders := *existingFolders

	// skipping label that already has a matching folder
	if newFolder, ok := folders[labelName]; ok {
		return newFolder, nil
	}

	tmpFolderName := "l2f-tmp-" + labelName
	newFolder, err := s.Client.CreateLabel(ctx, proton.CreateLabelReq{
		Name:     tmpFolderName,
		Type:     proton.LabelTypeFolder,
		ParentID: parentID,
	})
	if err != nil {
		return proton.Label{}, fmt.Errorf("failed to create %s folder %w", labelName, err)
	}

	folders[tmpFolderName] = newFolder
	return newFolder, nil
}

func (s *Service) CreateMissingFolders(ctx context.Context, label proton.Label, existingFolders map[string]proton.Label) (proton.Label, error) {
	if strings.Contains(label.Name, "-") {
		// handle each sub
		subFolders := strings.Split(label.Name, "-")
		lastNewFolderInChain, err := s.createSubFolders(ctx, subFolders, &existingFolders)
		if err != nil {
			return proton.Label{}, err
		}

		return lastNewFolderInChain, nil
	}

	newFolder, err := s.createMissingFolder(ctx, label.Name, "", &existingFolders)
	if err != nil {
		return proton.Label{}, err
	}

	return newFolder, nil
}

/*
ListEmails returns a list of all emails with a specific labelId
*/
func (s *Service) ListEmailsWithLabel(ctx context.Context, labelId string) ([]proton.Message, error) {
	var res struct {
		Messages []proton.Message
	}

	if err := s.Client.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/mail/v4/messages")
	}); err != nil {
		return []proton.Message{}, err
	}

	return res.Messages, nil
}

/*
	MigrateLabelsToFolders migrates labels to a replicated folder

- Create a l2f-tmp-{label} folder that the the label and emails will migrate to
- Move all emails with that label to the folder
- Remove the label from each email
- Delete the label
- Update all l2f-tmp-{label} to {label}
*/
func (s *Service) MigrateLabelsToFolders(ctx context.Context, labelsToIgnore []string) error {
	labels, err := s.Client.GetLabels(ctx, proton.LabelTypeLabel)
	if err != nil {
		return fmt.Errorf("failed to get labels %w", err)
	}

	folders, err := s.Client.GetLabels(ctx, proton.LabelTypeFolder)
	if err != nil {
		return fmt.Errorf("failed to get folders %w", err)
	}

	mappedFolders := make(map[string]proton.Label)
	for _, folder := range folders {
		mappedFolders[folder.Name] = folder
	}

	for _, label := range labels {
		// skipping labels that user requested to ignore
		if slices.Contains(labelsToIgnore, label.Name) {
			continue
		}

		newFolder, err := s.CreateMissingFolders(ctx, label, mappedFolders)
		if err != nil {
			return err
		}
		fmt.Println("Created %s tmp folder to replace label", newFolder.Name)

		emails, err := s.ListEmailsWithLabel(ctx, label.Name)
		if err != nil {
			return err
		}
		fmt.Println("Moving %d emails from %s to %s", len(emails), label.Name, newFolder.Name)

		// removing old label
		emailIds := FlattenToMessageIds(emails)
		if err := s.Client.UnlabelMessages(ctx, emailIds, label.ID); err != nil {
			return err
		}

		if err := s.Client.LabelMessages(ctx, emailIds, newFolder.ID); err != nil {
			return err
		}
		fmt.Println("Moved all emails from %s to %s", label.Name, newFolder.Name)

		if err := s.Client.DeleteLabel(ctx, label.ID); err != nil {
			return err
		}
	}

	for _, label := range mappedFolders {
		if strings.Contains(label.Name, "l2f-tmp-") {
			newName := strings.Replace(label.Name, "l2f-tmp-", "", 1)
			newLabel, err := s.Client.UpdateLabel(ctx, label.ID, proton.UpdateLabelReq{Name: newName})
			if err != nil {
				return err
			}

			fmt.Println("Updated %s to %s", label.Name, newLabel.Name)
		}
	}

	return nil
}
