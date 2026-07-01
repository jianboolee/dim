package dim

type LoginUserOption interface {
	applyLoginUserOption(*loginUserOptions)
}

type loginUserOptions struct {
	ensureUser bool
}

func defaultLoginUserOptions() loginUserOptions {
	return loginUserOptions{
		ensureUser: true,
	}
}

type SendMessageOption interface {
	applySendMessageOption(*sendMessageOptions)
}

type sendMessageOptions struct {
	ensureUsers           bool
	initialSenderMuted    *bool
	initialPeerMuted      *bool
	initialMemberSettings map[string]ConversationInitialMemberSettings
}

func defaultSendMessageOptions() sendMessageOptions {
	return sendMessageOptions{
		ensureUsers: true,
	}
}

func (o *sendMessageOptions) ensureInitialMemberSettings() {
	if o.initialMemberSettings == nil {
		o.initialMemberSettings = map[string]ConversationInitialMemberSettings{}
	}
}

type withoutEnsureUsersOption struct{}

func WithoutEnsureUsers() SendMessageOption {
	return withoutEnsureUsersOption{}
}

func (o withoutEnsureUsersOption) applySendMessageOption(options *sendMessageOptions) {
	options.ensureUsers = false
}

type initialPeerMutedOption struct {
	muted bool
}

func WithInitialPeerMuted(muted bool) SendMessageOption {
	return initialPeerMutedOption{muted: muted}
}

func (o initialPeerMutedOption) applySendMessageOption(options *sendMessageOptions) {
	options.initialPeerMuted = &o.muted
}

type initialSenderMutedOption struct {
	muted bool
}

func WithInitialSenderMuted(muted bool) SendMessageOption {
	return initialSenderMutedOption{muted: muted}
}

func (o initialSenderMutedOption) applySendMessageOption(options *sendMessageOptions) {
	options.initialSenderMuted = &o.muted
}
