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

type PrivateConversationServiceOption interface {
	applyPrivateConversationServiceOption(*privateConversationServiceOptions)
}

type privateConversationServiceOptions struct {
	ensureUsers           bool
	initialSenderMuted    *bool
	initialPeerMuted      *bool
	initialMemberSettings map[string]ConversationInitialMemberSettings
}

func defaultPrivateConversationServiceOptions() privateConversationServiceOptions {
	return privateConversationServiceOptions{
		ensureUsers: true,
	}
}

func (o *privateConversationServiceOptions) ensureInitialMemberSettings() {
	if o.initialMemberSettings == nil {
		o.initialMemberSettings = map[string]ConversationInitialMemberSettings{}
	}
}

type GroupConversationServiceOption interface {
	applyGroupConversationServiceOption(*groupConversationServiceOptions)
}

type groupConversationServiceOptions struct {
	ensureUsers bool
}

func defaultGroupConversationServiceOptions() groupConversationServiceOptions {
	return groupConversationServiceOptions{
		ensureUsers: true,
	}
}

type GroupMemberServiceOption interface {
	applyGroupMemberServiceOption(*groupMemberServiceOptions)
}

type groupMemberServiceOptions struct {
	ensureUsers bool
}

func defaultGroupMemberServiceOptions() groupMemberServiceOptions {
	return groupMemberServiceOptions{
		ensureUsers: true,
	}
}

type withoutEnsureUsersOption struct{}

func WithoutEnsureUsers() withoutEnsureUsersOption {
	return withoutEnsureUsersOption{}
}

func (o withoutEnsureUsersOption) applyPrivateConversationServiceOption(options *privateConversationServiceOptions) {
	options.ensureUsers = false
}

func (o withoutEnsureUsersOption) applyGroupConversationServiceOption(options *groupConversationServiceOptions) {
	options.ensureUsers = false
}

func (o withoutEnsureUsersOption) applyGroupMemberServiceOption(options *groupMemberServiceOptions) {
	options.ensureUsers = false
}

type initialPeerMutedOption struct {
	muted bool
}

func WithInitialPeerMuted(muted bool) initialPeerMutedOption {
	return initialPeerMutedOption{muted: muted}
}

func (o initialPeerMutedOption) applyPrivateConversationServiceOption(options *privateConversationServiceOptions) {
	options.initialPeerMuted = &o.muted
}

type initialSenderMutedOption struct {
	muted bool
}

func WithInitialSenderMuted(muted bool) initialSenderMutedOption {
	return initialSenderMutedOption{muted: muted}
}

func (o initialSenderMutedOption) applyPrivateConversationServiceOption(options *privateConversationServiceOptions) {
	options.initialSenderMuted = &o.muted
}
