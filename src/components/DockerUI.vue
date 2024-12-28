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
      <div class="container-status">
        Status: {{ selectedContainer.Status }}
      </div>
      <div class="container-actions">
        <button v-if="!isContainerRunning(selectedContainer.Status)" @click="startContainer" class="action-btn start-btn">START</button>
        <button v-if="isContainerRunning(selectedContainer.Status)" @click="stopContainer" class="action-btn stop-btn">STOP</button>
        <button @click="fetchContainers" class="action-btn refresh-btn">REFRESH</button>
        <button @click="closeContainer" class="action-btn close-btn">CLOSE</button>
        <button v-if="isContainerRunning(selectedContainer.Status)" @click="showContainerConsole" class="action-btn console-btn">Console</button>
        <button v-if="isContainerRunning(selectedContainer.Status)" @click="showContainerLogs" class="action-btn logs-btn">Logs</button>
        <button @click="deleteContainer" class="action-btn delete-btn">DELETE</button>
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

    ws.onopen = () => {
      console.log('WebSocket connection opened');
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

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
        // Update the status of the selected container
        if (selectedContainer.value) {
          const updatedContainer = containers.value.find(c => c.ID === selectedContainer.value?.ID);
          if (updatedContainer) {
            selectedContainer.value = updatedContainer;
          }
        }
      } catch (error) {
        console.error('Error fetching containers:', error);
      }
    };

    const deleteContainer = () => {
      console.log('Deleting container:', selectedContainer.value);
      if (selectedContainer.value) {
        sendMessage('delete', selectedContainer.value.ID);
        axios.delete(serverURL + '/api/containers/' + selectedContainer.value.ID)
          .then(() => {
            selectedContainer.value = null;
            fetchContainers();  // Fetch containers after deleting
          });
      }
    };

    const sendMessage = (action: string, containerID: string, command?: string) => {
      const message = {
        action: action,
        container: containerID,
        command: command || '',
      };
      ws.send(JSON.stringify(message));
      console.log('Sent message:', message);
    };

    // Add a new container , when the user clicks on the + ADD button redirct to the AddContainer page 
    const addContainer = () => {
      console.log('Adding a new container');
      // Redirect to the AddContainer page
      this.$router.push('/addContainer');
    };

    const startContainer = () => {
      console.log('Starting container:', selectedContainer.value);
      if (selectedContainer.value) {
        sendMessage('start', selectedContainer.value.ID);
        axios.post(serverURL + '/api/containers/' + selectedContainer.value.ID + '/start')
          .then(fetchContainers);  // Fetch containers after starting
      }
    };

    const stopContainer = () => {
      console.log('Stopping container:', selectedContainer.value);
      if (selectedContainer.value) {
        sendMessage('stop', selectedContainer.value.ID);
        axios.post(serverURL + '/api/containers/' + selectedContainer.value.ID + '/stop')
          .then(fetchContainers);  // Fetch containers after stopping
      }
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

    const isContainerRunning = (status: string) => {
      return status.toLowerCase().includes('up') || status.toLowerCase().includes('running');
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
      startContainer,
      stopContainer,
      isContainerRunning,
    };
  },
});
</script>

<style src="../css/DockerUI.css" scoped></style>
