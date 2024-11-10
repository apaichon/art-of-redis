// components/ScoreForm.vue
<template>
  <form class="score-form card" @submit.prevent="handleSubmit">
    <h2>Update Score</h2>
    
    <div class="form-group">
      <label for="playerId">Player ID</label>
      <input 
        id="playerId"
        v-model="form.id" 
        type="text" 
        placeholder="Enter player ID"
        :disabled="disabled"
        required
      >
    </div>
    
    <div class="form-group">
      <label for="playerName">Name</label>
      <input 
        id="playerName"
        v-model="form.name" 
        type="text" 
        placeholder="Enter player name"
        :disabled="disabled"
        required
      >
    </div>
    
    <div class="form-group">
      <label for="playerScore">Score</label>
      <input 
        id="playerScore"
        v-model.number="form.score" 
        type="number" 
        min="0"
        placeholder="Enter score"
        :disabled="disabled"
        required
      >
    </div>
    
    <button 
      type="submit" 
      class="submit-button"
      :disabled="disabled || !isFormValid"
    >
      Update Score
    </button>
  </form>
</template>

<script setup>
import { reactive, computed } from 'vue'

const props = defineProps({
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['submit'])

const form = reactive({
  id: '',
  name: '',
  score: 0
})

const isFormValid = computed(() => {
  return form.id && form.name && form.score > 0
})

const handleSubmit = () => {
  if (!isFormValid.value) return
  
  emit('submit', { ...form })
  
  // Clear form
  form.id = ''
  form.name = ''
  form.score = 0
}
</script>

<style scoped>
.score-form {
  margin-top: var(--spacing-lg);
}

.submit-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>