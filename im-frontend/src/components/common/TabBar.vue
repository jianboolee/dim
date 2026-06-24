<template>
  <nav class="tab-bar">
    <div class="tab-content">
      <RouterLink to="/" class="tab-item" replace active-class="active">
        <span>首页</span>
      </RouterLink>
      <RouterLink to="/favorites" class="tab-item" replace active-class="active">
        <span>收藏</span>
      </RouterLink>
      <RouterLink to="/publish" class="tab-item publish" replace active-class="active">
        <i class="bi bi-plus"></i>
        <span class="hidden">发布</span>
      </RouterLink>
      <RouterLink to="/messages" class="tab-item" replace active-class="active">
        <span>消息</span>
        <span v-if="unreadCount" class="badge badge-danger">
          {{ unreadCount }}
        </span>
      </RouterLink>
      <RouterLink to="/my" class="tab-item" replace active-class="active">
        <span>我的</span>
      </RouterLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUnreadMessageStore } from '@/stores/unreadMessage'
const unreadMessageStore = useUnreadMessageStore()

// 修改模板中的未读消息数显示
const unreadCount = computed(() => {
  return unreadMessageStore.unreadCount > 99 ? '99+' : unreadMessageStore.unreadCount
})
</script>

<style scoped>
.tab-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  /* box-shadow: 0 -1px 4px rgba(0,0,0,0.1); */
  z-index: 100;
  border-top: 1px solid var(--border-color);
  padding-bottom: env(safe-area-inset-bottom);
}

.tab-content {
  height: 59px;
  display: flex;
  justify-content: space-around;
  align-items: center;
}

.tab-item {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-color-secondary);
  text-decoration: none;
  font-size: 14px;
  font-weight: 600;
  padding: 8px 0;
  position: relative;
}

.tab-item.active {
  color: var(--primary-color);
  font-weight: 600;
}

.tab-item.publish {
  width: 52px;
  height: 36px;
  border: 2px solid #fff;
  border-radius: 12px;
  padding: 0;
  background: var(--primary-color);
  color: white;
}
.tab-item.publish.active {
  border-color: var(--primary-color);
}

.tab-item.publish i {
  font-size: 26px;
  font-weight: 800;
}

.hidden {
  display: none;
}
.tab-bar .badge {
  position: absolute;
  top: 4px;
  right: -10px;
  font-size: 12px;
  line-height: 1;
  padding: 2px 4px;
  border-radius: 10px;
  z-index: 10;
  min-width: 16px;
  text-align: center; 
}
.tab-bar .badge-danger {
  background: var(--danger-color);
  color: white;
}
</style> 
