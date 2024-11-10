import { ref, watch } from 'vue'
import { useWebSocket } from './useWebSocket'

export function useLeaderboard() {
  const rankings = ref([])
  const recentlyUpdated = ref({})
  const { isConnected, lastMessage, send } = useWebSocket('ws://localhost:9002/ws')

  watch(lastMessage, (update) => {
    if (!update) return

    if (update.type === 'full_update') {
      rankings.value = update.rankings
    } else if (update.type === 'update') {
      rankings.value = update.rankings
      
      if (update.player) {
        recentlyUpdated.value[update.player.id] = true
        setTimeout(() => {
          recentlyUpdated.value[update.player.id] = false
        }, 2000)
      }
    }
  })

  const updateScore = (scoreData) => {
    if (!isConnected.value) {
      throw new Error('Not connected to server')
    }
    return send(scoreData)
  }

  return {
    rankings,
    recentlyUpdated,
    isConnected,
    updateScore
  }
}