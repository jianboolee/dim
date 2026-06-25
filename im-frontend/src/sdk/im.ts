// 消息类型枚举
export enum MessageType {
    Text = 'text',
    Image = 'image',
    Video = 'video',
    Audio = 'audio',
    Card = 'card',
    Link = 'link',
    Ping = 'ping',
    Pong = 'pong'
  }
  
  // 消息状态枚举
  export enum MessageStatus {
    Sending = 'sending',
    Sent = 'sent',
    Delivered = 'delivered',
    Failed = 'failed'
  }
  
  // SDK 配置接口
  export interface IMSDKOptions {
    baseURL: string;
    wsURL?: string;
    token: string;
  }
  
  // 媒体信息接口
  export interface MediaInfo {
    url: string;
    size: number;
    duration?: number;
    width?: number;
    height?: number;
    format: string;
    thumbnail?: string;
    uploading?: boolean;
  }
  
  // 卡片信息接口
  export interface CardInfo {
    title: string;
    description: string;
    path: string;
    image_url: string;
    price: number;
    currency: string;
  }
  
  // 链接信息接口
  export interface LinkInfo {
    title: string;
    description: string;
    url: string;
    image_url: string;
  }
  
  // 消息接口
  export interface Message {
    id?: string;
    client_message_id?: string;
    conversation_id?: string;
    from_id?: string;
    to_id: string;
    type: MessageType;
    content: string;
    status?: MessageStatus;
    media_info?: MediaInfo;
    card_info?: CardInfo;
    link_info?: LinkInfo;
    created_at?: string;
    updated_at?: string;
  }
  
  // 会话接口
  export interface ConversationUserState {
    last_activated_at?: string;
    last_read_at?: string;
    unread_count?: number;
  }

  export interface Conversation {
    id: string;
    type: string;
    participants: string[];
    last_message: Message;
    to_user_info?: {
      id: string;
      nickname?: string;
      avatar?: string;
    };
    image_url: string;
    user_states?: { [key: string]: ConversationUserState };
    created_at: string;
    updated_at: string;
    last_activity: string;
  }

  export interface ConversationPage {
    items: Conversation[];
    next_cursor?: string;
    has_more: boolean;
  }

  export interface ConversationQueryParams {
    limit?: number;
    cursor?: string;
    q?: string;
    active_conversation_id?: string;
  }
  
  // 会话状态接口
  export interface Session {
    user_id: string;
    is_online: boolean;
    last_seen: string;
  }
  
  // 消息查询参数接口
  export interface MessageQueryParams {
    before_id?: string;
    after_id?: string;
    start_time?: string;
    end_time?: string;
    limit?: number;
  }

  interface ApiResponse<T> {
    code: number;
    message?: string;
    data: T;
  }

  function normalizeMessage(raw: Record<string, unknown>): Message {
    const payload = raw.payload as Record<string, unknown> | undefined
    let id: string | undefined
    if (raw.id != null) {
      if (typeof raw.id === 'string') {
        id = raw.id
      } else if (typeof raw.id === 'object' && raw.id !== null && '$oid' in raw.id) {
        id = String((raw.id as { $oid: string }).$oid)
      } else {
        id = String(raw.id)
      }
    }

    return {
      id,
      client_message_id: raw.client_message_id == null ? undefined : String(raw.client_message_id),
      conversation_id: raw.conversation_id == null ? undefined : String(raw.conversation_id),
      from_id: String(raw.sender_id ?? raw.from_id ?? ''),
      to_id: String(raw.receiver_id ?? raw.to_id ?? ''),
      type: (raw.type as MessageType) ?? MessageType.Text,
      content: String(raw.content ?? ''),
      status: raw.status as MessageStatus | undefined,
      media_info: raw.media_info as MediaInfo | undefined ?? payloadToMediaInfo(payload),
      card_info: raw.card_info as CardInfo | undefined ?? payloadToCardInfo(payload),
      link_info: raw.link_info as LinkInfo | undefined ?? payloadToLinkInfo(payload),
      created_at: raw.created_at as string | undefined,
      updated_at: raw.updated_at as string | undefined,
    };
  }

  function payloadToMediaInfo(payload?: Record<string, unknown>): MediaInfo | undefined {
    if (!payload?.url) return undefined

    return {
      url: String(payload.url),
      size: Number(payload.size ?? 0),
      duration: payload.duration == null ? undefined : Number(payload.duration),
      width: payload.width == null ? undefined : Number(payload.width),
      height: payload.height == null ? undefined : Number(payload.height),
      format: String(payload.ext_string ?? ''),
    }
  }

  function payloadToCardInfo(payload?: Record<string, unknown>): CardInfo | undefined {
    if (!payload?.title && !payload?.url) return undefined

    return {
      title: String(payload.title ?? ''),
      description: String(payload.description ?? ''),
      path: String(payload.url ?? ''),
      image_url: String(payload.image_url ?? ''),
      price: Number(payload.price ?? 0),
      currency: String(payload.currency ?? ''),
    }
  }

  function payloadToLinkInfo(payload?: Record<string, unknown>): LinkInfo | undefined {
    if (!payload?.url) return undefined

    return {
      title: String(payload.title ?? ''),
      description: String(payload.description ?? ''),
      url: String(payload.url),
      image_url: String(payload.image_url ?? ''),
    }
  }

  function mediaInfoToPayload(mediaInfo?: MediaInfo): Record<string, unknown> | undefined {
    if (!mediaInfo) return undefined

    return {
      url: mediaInfo.url,
      size: mediaInfo.size,
      duration: mediaInfo.duration,
      width: mediaInfo.width,
      height: mediaInfo.height,
      ext_string: mediaInfo.format,
    }
  }

  function cardInfoToPayload(cardInfo?: CardInfo): Record<string, unknown> | undefined {
    if (!cardInfo) return undefined

    return {
      title: cardInfo.title,
      description: cardInfo.description,
      url: cardInfo.path,
      image_url: cardInfo.image_url,
      price: cardInfo.price,
      currency: cardInfo.currency,
    }
  }

  function linkInfoToPayload(linkInfo?: LinkInfo): Record<string, unknown> | undefined {
    if (!linkInfo) return undefined

    return {
      title: linkInfo.title,
      description: linkInfo.description,
      url: linkInfo.url,
      image_url: linkInfo.image_url,
    }
  }

  function buildPayload(type: MessageType, mediaInfo?: MediaInfo, cardInfo?: CardInfo, linkInfo?: LinkInfo) {
    if ([MessageType.Image, MessageType.Video, MessageType.Audio].includes(type)) {
      return mediaInfoToPayload(mediaInfo)
    }
    if (type === MessageType.Card) {
      return cardInfoToPayload(cardInfo)
    }
    if (type === MessageType.Link) {
      return linkInfoToPayload(linkInfo)
    }
    return undefined
  }

  function normalizeConversation(raw: Record<string, unknown>): Conversation {
    const conv = raw as unknown as Conversation
    return {
      ...conv,
      id: String(raw.id ?? conv.id ?? ''),
    }
  }

  async function apiRequest<T>(
    baseURL: string,
    path: string,
    token: string,
    init: RequestInit = {},
  ): Promise<T> {
    const headers: Record<string, string> = {
      Authorization: `Bearer ${token}`,
      ...(init.headers as Record<string, string> | undefined),
    };

    if (init.body && !(init.headers as Record<string, string> | undefined)?.['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }

    const response = await fetch(`${baseURL}${path}`, {
      ...init,
      headers,
    });

    if (!response.ok) {
      throw new Error(`Request failed: ${response.statusText}`);
    }

    const json = await response.json();
    if (json && typeof json.code === 'number' && 'data' in json) {
      return (json as ApiResponse<T>).data;
    }

    return json as T;
  }
  
  // 连接状态接口
  export interface ConnectionStatus {
    status: 'connected' | 'disconnected' | 'error';
    error?: any;
  }
  
  // 消息处理器类型
  export type MessageHandler = (message: Message) => void;
  
  // 连接状态处理器类型
  export type ConnectionHandler = (status: ConnectionStatus) => void;
  
  /**
   * IM SDK
   * 封装了 WebSocket 连接、消息发送、接收等功能
   */
  class IMSDK {
    private baseURL: string;
    private wsURL: string;
    private token: string;
    private ws: WebSocket | null;
    private messageHandlers: MessageHandler[];
    private connectionHandlers: ConnectionHandler[];
    private heartbeatInterval: number;
    private heartbeatTimer: ReturnType<typeof setInterval> | null;
  
    /**
     * 构造函数
     */
    constructor(options: IMSDKOptions) {
      this.baseURL = options.baseURL || (typeof window !== 'undefined' ? window.location.origin : '');
      this.wsURL = options.wsURL || `${this.baseURL.replace(/^http/, 'ws')}/im/ws`;
      this.token = options.token;
      this.ws = null;
      this.messageHandlers = [];
      this.connectionHandlers = [];
      this.heartbeatInterval = 30000; // 默认30秒发送一次心跳
      this.heartbeatTimer = null;
    }
  
    /**
     * 连接 WebSocket
     */
    async connect(): Promise<void> {
      if (this.ws?.readyState === WebSocket.OPEN) {
        return;
      }

      this.ws?.close();

      return new Promise((resolve, reject) => {
        const ws = new WebSocket(`${this.wsURL}?token=${encodeURIComponent(this.token)}`);
        this.ws = ws;
        let settled = false;

        const settleResolve = () => {
          if (settled) return;
          settled = true;
          resolve();
        };

        const settleReject = (error: unknown) => {
          if (settled) return;
          settled = true;
          reject(error);
        };
        
        ws.onopen = () => {
          this._notifyConnectionHandlers({ status: 'connected' });
          settleResolve();
        };
        
        ws.onclose = () => {
          if (this.ws === ws) {
            this.ws = null;
            this._notifyConnectionHandlers({ status: 'disconnected' });
          }
          settleReject(new Error('WebSocket closed before connection was established'));
        };
        
        ws.onmessage = (event: MessageEvent) => {
          const raw = JSON.parse(event.data);
          if (raw?.type === MessageType.Ping || raw?.type === MessageType.Pong) {
            return;
          }
          const message = normalizeMessage(raw);
          this._notifyMessageHandlers(message);
        };
        
        ws.onerror = (error: Event) => {
          if (this.ws !== ws) return;
          this._notifyConnectionHandlers({ status: 'error', error });
          settleReject(error);
        };
      });
    }
  
    /**
     * 断开 WebSocket 连接
     */
    disconnect(): void {
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
    }

    isSocketOpen(): boolean {
      return this.ws?.readyState === WebSocket.OPEN;
    }

    updateToken(token: string): void {
      this.token = token;
    }
  
    /**
     * 创建消息
     */
    createMessage(message: Message): Message {
      return {
        ...message,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    }
    /**
     * 通过 WebSocket 发送消息
     */
    async sendMessageWS(
      conversationId: string,
      type: MessageType = MessageType.Text,
      content: string = '',
      mediaInfo?: MediaInfo,
      cardInfo?: CardInfo,
      linkInfo?: LinkInfo,
      clientMessageId?: string
    ): Promise<void> {
      return new Promise((resolve, reject) => {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
          reject(new Error('WebSocket is not connected'));
          return;
        }
  
        try {
          const message: Message = {
            client_message_id: clientMessageId,
            conversation_id: conversationId,
            to_id: '',
            type,
            content
          };
  
          // 根据消息类型添加必要的信息
          if (type === MessageType.Text && !content) {
            reject(new Error('Content is required for text message'));
            return;
          }
  
          if ([MessageType.Image, MessageType.Video, MessageType.Audio].includes(type)) {
            if (!mediaInfo?.url) {
              reject(new Error('MediaInfo with URL is required for media message'));
              return;
            }
            message.media_info = mediaInfo;
          }
  
          if (type === MessageType.Card) {
            if (!cardInfo?.path) {
              reject(new Error('CardInfo with path is required for card message'));
              return;
            }
            message.card_info = cardInfo;
          }
  
          if (type === MessageType.Link) {
            if (!linkInfo?.url) {
              reject(new Error('LinkInfo with URL is required for link message'));
              return;
            }
            message.link_info = linkInfo;
          }
  
          this.ws.send(JSON.stringify(message));
          resolve();
        } catch (error) {
          reject(new Error(`Failed to send message: ${error}`));
        }
      });
    }
  
    /**
     * 发送消息
     */
    async sendMessage(
      conversationId: string,
      type: MessageType = MessageType.Text,
      content: string = '',
      mediaInfo?: MediaInfo,
      cardInfo?: CardInfo,
      linkInfo?: LinkInfo,
      clientMessageId?: string
    ): Promise<Message> {
      const message: Message = {
        client_message_id: clientMessageId,
        conversation_id: conversationId,
        to_id: '',
        type,
        content
      };
  
      if (type === MessageType.Text && !content) {
        throw new Error('Content is required for text message');
      }
  
      if ([MessageType.Image, MessageType.Video, MessageType.Audio].includes(type)) {
        if (!mediaInfo?.url) {
          throw new Error('MediaInfo with URL is required for media message');
        }
        message.media_info = mediaInfo;
      }
  
      if (type === MessageType.Card) {
        if (!cardInfo?.path) {
          throw new Error('CardInfo with path is required for card message');
        }
        message.card_info = cardInfo;
      }
  
      if (type === MessageType.Link) {
        if (!linkInfo?.url) {
          throw new Error('LinkInfo with URL is required for link message');
        }
        message.link_info = linkInfo;
      }

      const payload = buildPayload(type, mediaInfo, cardInfo, linkInfo);
  
      const response = await fetch(`${this.baseURL}/im/api/conversations/${conversationId}/messages`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.token}`
        },
        body: JSON.stringify({
          client_message_id: clientMessageId,
          type,
          content,
          payload,
        })
      });

      if (!response.ok) {
        throw new Error(`Failed to send message: ${response.statusText}`);
      }

      const json = await response.json();
      if (json && typeof json.code === 'number' && json.data) {
        return normalizeMessage(json.data);
      }

      return normalizeMessage(json);
    }
  
    async getConversationMessages(conversationId: string, params: MessageQueryParams = {}): Promise<Message[]> {
      const queryParams = new URLSearchParams();
      if (params.before_id) queryParams.append('before_id', params.before_id);
      if (params.after_id) queryParams.append('after_id', params.after_id);
      if (params.start_time) queryParams.append('start_time', params.start_time);
      if (params.end_time) queryParams.append('end_time', params.end_time);
      if (params.limit) queryParams.append('limit', params.limit.toString());

      const suffix = queryParams.toString() ? `?${queryParams}` : '';
      const data = await apiRequest<Record<string, unknown>[]>(
        this.baseURL,
        `/im/api/conversations/${conversationId}/messages${suffix}`,
        this.token,
      );

      return (data ?? []).map((item) => normalizeMessage(item));
    }
  
    /**
     * 获取会话列表
     */
    async getConversationPage(params: ConversationQueryParams = {}): Promise<ConversationPage> {
      const queryParams = new URLSearchParams();
      if (params.limit) queryParams.append('limit', params.limit.toString());
      if (params.cursor) queryParams.append('cursor', params.cursor);
      if (params.q) queryParams.append('q', params.q);
      if (params.active_conversation_id) {
        queryParams.append('active_conversation_id', params.active_conversation_id);
      }

      const suffix = queryParams.toString() ? `?${queryParams}` : '';
      const data = await apiRequest<Record<string, unknown>[] | Record<string, unknown>>(
        this.baseURL,
        `/im/api/conversations${suffix}`,
        this.token,
      );

      if (Array.isArray(data)) {
        return {
          items: data.map((item) => normalizeConversation(item)),
          has_more: false,
        };
      }

      const items = Array.isArray(data.items)
        ? (data.items as Record<string, unknown>[]).map((item) => normalizeConversation(item))
        : [];

      return {
        items,
        next_cursor: typeof data.next_cursor === 'string' ? data.next_cursor : undefined,
        has_more: Boolean(data.has_more),
      };
    }

    async getConversations(params: ConversationQueryParams = {}): Promise<Conversation[]> {
      const page = await this.getConversationPage(params);
      return page.items;
    }

    async getConversation(conversationId: string): Promise<Conversation> {
      const data = await apiRequest<Record<string, unknown>>(
        this.baseURL,
        `/im/api/conversations/${conversationId}`,
        this.token,
      );

      return normalizeConversation(data);
    }

    async activateConversation(conversationId: string): Promise<Conversation> {
      const data = await apiRequest<Record<string, unknown>>(
        this.baseURL,
        `/im/api/conversations/${conversationId}/activate`,
        this.token,
        { method: 'POST' },
      );

      return normalizeConversation(data);
    }
  
    /**
     * 获取用户在线状态
     */
    async getUserStatus(userID: string): Promise<Session> {
      const data = await apiRequest<Session>(
        this.baseURL,
        `/im/api/sessions/${userID}`,
        this.token,
      );

      return data;
    }
  
    /**
     * 保持在线状态
     */
    async keepAlive(): Promise<void> {
      await apiRequest<void>(
        this.baseURL,
        '/im/api/sessions/keepalive',
        this.token,
        { method: 'POST' },
      );
    }
  
    /**
     * 监听新消息
     */
    onMessage(handler: MessageHandler): void {
      this.messageHandlers.push(handler);
    }
  
    /**
     * 监听连接状态
     */
    onConnection(handler: ConnectionHandler): void {
      this.connectionHandlers.push(handler);
    }
  
    /**
     * 移除消息监听器
     */
    offMessage(handler: MessageHandler): void {
      const index = this.messageHandlers.indexOf(handler);
      if (index > -1) {
        this.messageHandlers.splice(index, 1);
      }
    }
  
    /**
     * 移除连接状态监听器
     */
    offConnection(handler: ConnectionHandler): void {
      const index = this.connectionHandlers.indexOf(handler);
      if (index > -1) {
        this.connectionHandlers.splice(index, 1);
      }
    }
  
    /**
     * 内部方法：通知所有消息处理器
     */
    private _notifyMessageHandlers(message: Message): void {
      this.messageHandlers.forEach(handler => handler(message));
    }
  
    /**
     * 内部方法：通知所有连接状态处理器
     */
    private _notifyConnectionHandlers(status: ConnectionStatus): void {
      this.connectionHandlers.forEach(handler => handler(status));
    }
  
    /**
     * 启动心跳机制
     */
    startHeartbeat(): void {
      if (this.heartbeatTimer) {
        clearInterval(this.heartbeatTimer);
      }
  
      this.heartbeatTimer = setInterval(() => {
        this.send({ type: MessageType.Ping, to_id: '', content: '' });
      }, this.heartbeatInterval);
    }
  
    /**
     * 停止心跳机制
     */
    stopHeartbeat(): void {
      if (this.heartbeatTimer) {
        clearInterval(this.heartbeatTimer);
        this.heartbeatTimer = null;
      }
    }
  
    /**
     * 发送消息或心跳
     */
    private async send(message: Message): Promise<void> {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket is not connected');
      }
  
      try {
        this.ws.send(JSON.stringify(message));
      } catch (error) {
        console.error('Error sending message:', error);
        throw new Error(`Failed to send message: ${error}`);
      }
    }
  
    /**
     * 发送文本消息的快捷方法
     */
    async sendTextMessage(conversationId: string, content: string): Promise<Message> {
      return this.sendMessage(conversationId, MessageType.Text, content);
    }
  
    /**
     * 发送图片消息的快捷方法
     */
    async sendImageMessage(
      conversationId: string,
      content: string,
      url: string,
      width: number,
      height: number,
      size: number,
      format: string
    ): Promise<Message> {
      const mediaInfo: MediaInfo = {
        url,
        width,
        height,
        size,
        format
      };
      return this.sendMessage(conversationId, MessageType.Image, content, mediaInfo);
    }
  
    /**
     * 发送视频消息的快捷方法
     */
    async sendVideoMessage(
      conversationId: string,
      content: string,
      url: string,
      duration: number,
      width: number,
      height: number,
      size: number,
      format: string
    ): Promise<Message> {
      const mediaInfo: MediaInfo = {
        url,
        duration,
        width,
        height,
        size,
        format
      };
      return this.sendMessage(conversationId, MessageType.Video, content, mediaInfo);
    }
  
    /**
     * 发送音频消息的快捷方法
     */
    async sendAudioMessage(
      conversationId: string,
      content: string,
      url: string,
      duration: number,
      size: number,
      format: string
    ): Promise<Message> {
      const mediaInfo: MediaInfo = {
        url,
        duration,
        size,
        format
      };
      return this.sendMessage(conversationId, MessageType.Audio, content, mediaInfo);
    }
  
    /**
     * 发送卡片消息的快捷方法
     */
    async sendCardMessage(
      conversationId: string,
      content: string,
      title: string,
      description: string,
      path: string,
      imageUrl: string,
      price: number = 0,
      currency: string = 'CNY'
    ): Promise<Message> {
      const cardInfo: CardInfo = {
        title,
        description,
        path,
        image_url: imageUrl,
        price,
        currency
      };
      
      return this.sendMessage(conversationId, MessageType.Card, content, undefined, cardInfo);
    }
  
    /**
     * 发送链接消息的快捷方法
     */
    async sendLinkMessage(
      conversationId: string,
      content: string,
      title: string,
      description: string,
      url: string,
      imageUrl: string
    ): Promise<Message> {
      const linkInfo: LinkInfo = {
        title,
        description,
        url,
        image_url: imageUrl
      };
      
      return this.sendMessage(conversationId, MessageType.Link, content, undefined, undefined, linkInfo);
    }
  }
  
  export default IMSDK; 
