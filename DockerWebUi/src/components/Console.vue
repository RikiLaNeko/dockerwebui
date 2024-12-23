<template>
  <div class="console-container">
    <h2 class="console-title">Console Output for {{ containerId }}</h2>
    <div class="console-output" ref="consoleOutput"></div>
    <textarea
      v-model="command"
      @keyup.enter="executeCommand"
      placeholder="Enter command"
      class="command-input"
    ></textarea>
    <button @click="executeCommand" class="execute-button">Execute</button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineProps } from 'vue';

const props = defineProps<{ containerId: string }>();

const command = ref('');
const socket = ref<WebSocket | null>(null);
const consoleOutput = ref<HTMLDivElement | null>(null);

const displayMessage = (message: string) => {
  const messageElement = document.createElement('div');
  messageElement.innerText = message;
  if (consoleOutput.value) {
    consoleOutput.value.appendChild(messageElement);
    consoleOutput.value.scrollTop = consoleOutput.value.scrollHeight;
  }
};

const connectWebSocket = () => {
  socket.value = new WebSocket(`ws://localhost:3000/ws/${props.containerId}`);

  socket.value.onmessage = (event) => {
    displayMessage(event.data);
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

<style src="../css/Console.css" scoped></style>
