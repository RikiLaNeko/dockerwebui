<template>
  <div class="console-container">
    <h2 class="console-title">Console Output for {{ containerId }}</h2>
    <div ref="terminalContainer" class="terminal-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineProps, onBeforeUnmount } from 'vue';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

const props = defineProps<{ containerId: string }>();

const terminalContainer = ref<HTMLDivElement | null>(null);
let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let socket: WebSocket | null = null;

const connectWebSocket = () => {
  socket = new WebSocket(`ws://localhost:3000/ws/${props.containerId}`);

  socket.onmessage = (event) => {
    if (terminal) {
      terminal.write(event.data);
    }
  };

  socket.onclose = () => {
    console.error('WebSocket closed');
  };

  socket.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
};

const initializeTerminal = () => {
  terminal = new Terminal({
    cursorBlink: true,
    fontFamily: 'monospace',
    theme: {
      background: '#121212',
      foreground: '#f1f1f1'
    }
  });

  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);

  if (terminalContainer.value) {
    terminal.open(terminalContainer.value);
    fitAddon.fit();
  }

  terminal.onData((data) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(data);
    }
  });
};

onMounted(() => {
  initializeTerminal();
  connectWebSocket();
});

onBeforeUnmount(() => {
  if (terminal) {
    terminal.dispose();
  }
  if (socket) {
    socket.close();
  }
});
</script>

<style src="../css/Console.css" scoped></style>
