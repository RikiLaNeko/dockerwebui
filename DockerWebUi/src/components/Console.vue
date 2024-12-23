<template>
  <div>
    <h2>Console Output for {{ containerId }}</h2>
    <div class="console-output" ref="consoleOutput"></div>
    <textarea v-model="command" @keyup.enter="executeCommand" placeholder="Enter command" class="command-input"></textarea>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineProps } from 'vue';

const props = defineProps<{ containerId: string }>();

const command = ref('');
const socket = ref<WebSocket | null>(null);
const consoleOutput = ref<HTMLDivElement | null>(null);

const connectWebSocket = () => {
  socket.value = new WebSocket(`ws://localhost:3000/ws/${props.containerId}`);

  socket.value.onmessage = (event) => {
    const message = document.createElement('div');
    message.innerText = event.data;
    if (consoleOutput.value) {
      consoleOutput.value.appendChild(message);
      consoleOutput.value.scrollTop = consoleOutput.value.scrollHeight;
    }
  };

  socket.value.onclose = () => {
    console.error('WebSocket closed');
  };

  socket.value.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
};

const executeCommand = () => {
  if (socket.value && socket.value.readyState === WebSocket.OPEN) {
    socket.value.send(command.value + '\n');
    command.value = '';
  }
};

onMounted(() => {
  connectWebSocket();
});
</script>

<style scoped>
.console-output {
  background: #1a1a1a;
  color: #f1f1f1;
  padding: 1rem;
  border-radius: 0.5rem;
  height: 300px;
  overflow-y: auto;
  margin-bottom: 1rem;
}

pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}

.command-input {
  display: block;
  width: 100%;
  padding: 0.5rem;
  border-radius: 0.5rem;
  border: 1px solid #ccc;
  background: #1a1a1a;
  color: #f1f1f1;
}
</style>
