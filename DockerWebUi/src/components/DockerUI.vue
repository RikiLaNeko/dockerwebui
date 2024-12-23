<template>
  <div class="docker-ui-container">
    <!-- Sidebar -->
    <div :class="['sidebar', sidebarOpen ? 'sidebar-open' : 'sidebar-closed']">
      <button @click="toggleSidebar" class="sidebar-toggle-btn">
        {{ sidebarOpen ? 'Close' : 'Open' }}
      </button>
      <div v-if="sidebarOpen" class="sidebar-content">
        <ul>
          <li v-for="container in containers" :key="container.ID" @click="selectContainer(container)" class="sidebar-item">
            {{ container.Names }}
          </li>
        </ul>
        <button @click="addContainer" class="add-container-btn">+ ADD</button>
      </div>
    </div>

    <!-- Main content -->
    <div v-if="selectedContainer" class="main-content">
      <h1 class="container-title">{{ selectedContainer.Names }}</h1>
      <div class="container-actions">
        <button @click="sendMessage('start', selectedContainer.ID)" class="action-btn start-btn">START</button>
        <button @click="sendMessage('stop', selectedContainer.ID)" class="action-btn stop-btn">STOP</button>
        <button @click="fetchContainers" class="action-btn refresh-btn">REFRESH</button>
        <button @click="closeContainer" class="action-btn close-btn">CLOSE</button>
        <button @click="showContainerConsole" class="action-btn console-btn">Console</button>
        <button @click="showContainerLogs" class="action-btn logs-btn">Logs</button>
      </div>
      <div v-if="showConsole" class="console-output">
        <Console :containerId="selectedContainer.ID" />
      </div>
      <div v-if="showLogs" class="logs-output">
        <Logs :containerId="selectedContainer.ID" />
      </div>
    </div>

    <!-- Placeholder if no container selected -->
    <div v-else class="placeholder">
      <p>Select a container to get started.</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import axios from 'axios';
import Console from './Console.vue';
import Logs from './Logs.vue';

interface Container {
  ID: string;
  Names: string;
  Image: string;
  Status: string;
}

export default defineComponent({
  name: 'DockerUI',
  components: {
    Console,
    Logs,
  },
  setup() {
    const sidebarOpen = ref(false);
    const containers = ref<Container[]>([]);
    const selectedContainer = ref<Container | null>(null);
    const showConsole = ref(false);
    const showLogs = ref(false);
    const serverURL = 'http://localhost:3000';
    const ws = new WebSocket('ws://localhost:3000/ws');

    ws.onmessage = (event) => {
      // Handle incoming WebSocket messages
      console.log('WebSocket message:', event.data);
    };

    const toggleSidebar = () => {
      sidebarOpen.value = !sidebarOpen.value;
    };

    const selectContainer = (container: Container) => {
      selectedContainer.value = container;
      showConsole.value = false;
      showLogs.value = false;
    };

    const fetchContainers = async () => {
      try {
        const response = await axios.get(serverURL + '/api/containers/json');
        containers.value = response.data;
      } catch (error) {
        console.error('Error fetching containers:', error);
      }
    };

    const sendMessage = (action: string, containerID: string, command?: string) => {
      const message = {
        action: action,
        container: containerID,
        command: command || '',
      };
      ws.send(JSON.stringify(message));
    };

    const closeContainer = () => {
      selectedContainer.value = null;
      showConsole.value = false;
      showLogs.value = false;
    };

    const showContainerConsole = () => {
      showConsole.value = true;
      showLogs.value = false;
    };

    const showContainerLogs = () => {
      showLogs.value = true;
      showConsole.value = false;
    };

    onMounted(() => {
      fetchContainers();
    });

    return {
      sidebarOpen,
      containers,
      selectedContainer,
      showConsole,
      showLogs,
      toggleSidebar,
      selectContainer,
      fetchContainers,
      sendMessage,
      closeContainer,
      showContainerConsole,
      showContainerLogs,
    };
  },
});
</script>

<style src="../css/DockerUI.css" scoped></style>
