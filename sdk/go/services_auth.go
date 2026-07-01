package dim

import "context"

func (s *Services) LoginUser(ctx context.Context, user UserInput, options ...LoginUserOption) (*Session, error) {
	opts := defaultLoginUserOptions()
	for _, option := range options {
		option.applyLoginUserOption(&opts)
	}
	if opts.ensureUser {
		if err := s.client.EnsureUser(ctx, user); err != nil {
			return nil, err
		}
	}
	return s.client.Login(ctx, user.ID)
}
