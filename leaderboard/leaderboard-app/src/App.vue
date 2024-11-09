<template>
  <div class="app-container">
    <div class="leaderboard card">
      <div class="leaderboard-header">
        <h1 class="leaderboard-title">Real-time Leaderboard</h1>
        <ConnectionStatus :isConnected="isConnected" />
      </div>

      <LeaderboardTable 
        :rankings="rankings"
        :recentlyUpdated="recentlyUpdated"
        :loading="loading"
      />
    </div>

    <ScoreForm 
      @submit="handleScoreSubmit" 
      :disabled="!isConnected"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useLeaderboard } from '@/composables/useLeaderboard'
import ConnectionStatus from '@/components/ConnectionStatus.vue'
import LeaderboardTable from '@/components/LeaderboardTable.vue'
import ScoreForm from '@/components/ScoreForm.vue'

// State
const loading = ref(false)

// Initialize leaderboard
const { 
  rankings, 
  recentlyUpdated, 
  isConnected, 
  updateScore 
} = useLeaderboard()

// Handle score submission
const handleScoreSubmit = async (scoreData) => {
  if (!isConnected.value) {
    console.error('Not connected to server')
    return
  }

  try {
    loading.value = true
    await updateScore(scoreData)
  } catch (error) {
    console.error('Failed to update score:', error)
    // Use a proper toast/notification system instead of alert
    alert(error.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.app-container {
  max-width: 800px;
  margin: 0 auto;
  padding: var(--spacing-lg);
}

.card {
  background: var(--color-surface);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-lg);
  margin-bottom: var(--spacing-lg);
}

@media (max-width: 600px) {
  .app-container {
    padding: var(--spacing-sm);
  }

  .card {
    padding: var(--spacing-md);
  }
}
</style>