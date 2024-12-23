<template>
  <div class="flex h-screen">
    <!-- Sidebar -->
    <div :class="['transition-width duration-300', sidebarOpen ? 'w-64' : 'w-16']" class="bg-gray-800 text-white h-full">
      <button @click="toggleSidebar" class="m-2 p-2 bg-gray-700 rounded hover:bg-gray-600">
        {{ sidebarOpen ? 'Close' : 'Open' }}
      </button>
      <div v-if="sidebarOpen" class="mt-4">
        <ul>
          <li v-for="container in containers" :key="container.ID" @click="selectContainer(container)" class="p-2 hover:bg-gray-700 cursor-pointer">
            {{ container.Names }}
          </li>
        </ul>
        <button @click="addContainer" class="mt-4 p-2 bg-green-600 rounded hover:bg-green-500">
          + ADD
        </button>
      </div>
    </div>

    <!-- Main content -->
    <div v-if="selectedContainer" class="flex-1 bg-gray-100 p-4">
      <h1 class="text-2xl font-bold mb-4">{{ selectedContainer.Names }}</h1>
      <div class="flex gap-4">
        <button @click="sendMessage('start', selectedContainer.ID)" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-500">START</button>
        <button @click="sendMessage('stop', selectedContainer.ID)" class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-500">STOP</button>
        <button @click="fetchContainers" class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-500">REFRESH</button>
        <button @click="closeContainer" class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-500">CLOSE</button>
        <button @click="showContainerConsole" class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-500">Console</button>
        <button @click="showContainerLogs" class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-500">Logs</button>
      </div>
      <div v-if="showConsole" class="console-output">
        <Console :containerId="selectedContainer.ID" />
      </div>
      <div v-if="showLogs" class="logs-output">
        <Logs :containerId="selectedContainer.ID" />
      </div>
    </div>

    <!-- Placeholder if no container selected -->
    <div v-else class="flex-1 flex items-center justify-center text-gray-500">
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
    const sidebarOpen = ref(true);
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

<style scoped>
.console-output, .logs-output {
  background: #1a1a1a;
  color: #f1f1f1;
  padding: 1rem;
  border-radius: 0.5rem;
}
</style>
