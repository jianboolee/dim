import { storeToRefs } from 'pinia'
import { useConversationListStore } from '@/stores/conversationList'

export function useConversationList() {
  const store = useConversationListStore()

  return {
    ...storeToRefs(store),
    loadConversations: store.loadConversations,
    loadMoreConversations: store.loadMoreConversations,
    searchConversations: store.searchConversations,
    loadMoreSearchConversations: store.loadMoreSearchConversations,
    handleIncomingMessage: store.handleIncomingMessage,
    clearUnreadForPeer: store.clearUnreadForPeer,
    upsertConversation: store.upsertConversation,
    ensureConversationInList: store.ensureConversationInList,
    requestScrollToConversation: store.requestScrollToConversation,
    getPeerUserIds: store.getPeerUserIds,
    resetConversations: store.resetConversations,
  }
}
