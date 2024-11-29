// components/LeaderboardTable.vue
<template>
  <div class="leaderboard-table-container">
    <div v-if="loading" class="loading-overlay">
      Loading...
    </div>
    
    <table class="rankings-table" :class="{ 'is-loading': loading }">
      <thead>
        <tr>
          <th>Rank</th>
          <th>Player</th>
          <th>Score</th>
          <th>Last Updated</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="player in rankings" 
            :key="player.id" 
            :class="['player-row', { highlighted: recentlyUpdated[player.id] }]">
          <td class="rank">#{{ player.rank }}</td>
          <td>{{ player.name }}</td>
          <td class="score">{{ formatScore(player.score) }}</td>
          <td class="timestamp">{{ formatDateTime(player.updated_at) }}</td>  
        </tr>
        <tr v-if="!loading && rankings.length === 0">
          <td colspan="4" class="empty-state">
            No players found
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { formatScore, formatDateTime } from '@/utils/formatters'

defineProps({
  rankings: {
    type: Array,
    required: true
  },
  recentlyUpdated: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.leaderboard-table-container {
  position: relative;
  min-height: 200px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

.is-loading {
  opacity: 0.6;
}

.empty-state {
  text-align: center;
  color: var(--color-text-secondary);
  padding: var(--spacing-lg);
}
</style>