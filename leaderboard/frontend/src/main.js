import { createApp } from 'vue'
import App from './App.vue'
import './assets/styles/main.css'

const app = createApp(App)
app.mount('#app')

// src/composables/useWebSocket.js
import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(url) {
  const socket = ref(null)
  const isConnected = ref(false)
  const lastMessage = ref(null)

  const connect = () => {
    socket.value = new WebSocket(url)

    socket.value.onopen = () => {
      isConnected.value = true
      console.log('WebSocket connected')
    }

    socket.value.onclose = () => {
      isConnected.value = false
      console.log('WebSocket disconnected')
      setTimeout(connect, 5000)
    }

    socket.value.onmessage = (event) => {
      lastMessage.value = JSON.parse(event.data)
    }
  }

  const send = (data) => {
    if (socket.value?.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify(data))
      return true
    }
    return false
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    socket.value?.close()
  })

  return {
    isConnected,
    lastMessage,
    send
  }
}