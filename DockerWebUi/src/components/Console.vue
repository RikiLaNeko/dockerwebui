<template>
  <div>
    <h2>Console Output for {{ containerId }}</h2>
    <textarea v-model="command" placeholder="Enter command" class="command-input"></textarea>
    <button @click="executeCommand" class="execute-button">Execute</button>
    <pre>{{ consoleOutput }}</pre>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineProps } from 'vue';
import axios from 'axios';

const props = defineProps<{ containerId: string }>();

const consoleOutput = ref('');
const command = ref('');

const fetchConsoleOutput = async () => {
  try {
    const response = await axios.get(`http://localhost:3000/api/containers/${props.containerId}/logs`);
    consoleOutput.value = response.data;
  } catch (error) {
    console.error('Error fetching console output:', error);
  }
};

const executeCommand = async () => {
  try {
    const response = await axios.post(`http://localhost:3000/api/containers/${props.containerId}/exec`, {
      command: command.value,
    });
    consoleOutput.value = response.data;
  } catch (error) {
    console.error('Error executing command:', error);
  }
};

onMounted(() => {
  fetchConsoleOutput();
});
</script>

<style scoped>
pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}

.command-input {
  display: block;
  width: 100%;
  padding: 0.5rem;
  margin-bottom: 1rem;
  border-radius: 0.5rem;
  border: 1px solid #ccc;
  background: #1a1a1a;
  color: #f1f1f1;
}

.execute-button {
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  border: none;
  background-color: #38a169;
  color: white;
  cursor: pointer;
}

.execute-button:hover {
  background-color: #48bb78;
}
</style>
