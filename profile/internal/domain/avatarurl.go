package domain

import (
	"context"
	"fmt"
	"sync"

	"github.com/TATAROmangol/mess/profile/internal/model"
)

func (d *Domain) GetAvatarURL(ctx context.Context, key *string) (string, error) {
	if key == nil {
		return "", nil
	}

	avatarURL, err := d.Avatar.GetAvatarURL(ctx, *key)
	if err != nil {
		return "", fmt.Errorf("get avatar url: %v", err)
	}

	return avatarURL, nil
}

func (d *Domain) GetAvatarsURL(ctx context.Context, profiles []*model.Profile) (map[string]string, []error) {
	res := make(map[string]string)
	errors := make([]error, 0, len(profiles))

	wg := sync.WaitGroup{}
	wg.Add(len(profiles))

	mu := sync.Mutex{}

	ch := make(chan error)
	for _, profile := range profiles {
		go func() {
			defer wg.Done()

			url, err := d.GetAvatarURL(ctx, profile.AvatarKey)
			if err != nil {
				ch <- err
				return
			}

			mu.Lock()
			res[profile.SubjectID] = url
			mu.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for err := range ch {
		errors = append(errors, err)
	}

	return res, errors
}
