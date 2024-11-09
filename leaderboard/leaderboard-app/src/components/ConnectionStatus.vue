// components/ConnectionStatus.vue
<template>
  <div :class="['connection-status', statusClass]">
    {{ statusText }}
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  isConnected: {
    type: Boolean,
    required: true
  }
})

const statusClass = computed(() => props.isConnected ? 'connected' : 'disconnected')
const statusText = computed(() => props.isConnected ? 'Connected' : 'Disconnected')
</script>

<style scoped>
.connection-status {
  padding: var(--spacing-xs) var(--spacing-md);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  font-weight: 500;
}

.connected {
  background: var(--color-connected);
  color: var(--color-connected-text);
}

.disconnected {
  background: var(--color-disconnected);
  color: var(--color-disconnected-text);
}
</style>