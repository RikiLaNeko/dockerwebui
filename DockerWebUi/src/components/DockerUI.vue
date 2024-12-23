<template>
  <div class="flex h-screen">
    <!-- Barre latérale -->
    <div :class="['transition-width duration-300', sidebarOpen ? 'w-64' : 'w-16']" class="bg-gray-800 text-white h-full">
      <button @click="toggleSidebar" class="m-2 p-2 bg-gray-700 rounded hover:bg-gray-600">
        {{ sidebarOpen ? 'Close' : 'Open' }}
      </button>
      <div v-if="sidebarOpen" class="mt-4">
        <ul>
          <li v-for="container in containers" :key="container.id" @click="selectContainer(container)" class="p-2 hover:bg-gray-700 cursor-pointer">
            {{ container.name }}
          </li>
        </ul>
        <button @click="addContainer" class="mt-4 p-2 bg-green-600 rounded hover:bg-green-500">
          + ADD
        </button>
      </div>
    </div>

    <!-- Contenu principal -->
    <div v-if="selectedContainer" class="flex-1 bg-gray-100 p-4">
      <h1 class="text-2xl font-bold mb-4">{{ selectedContainer.name }}</h1>
      <div class="flex gap-4">
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-500">EDIT</button>
        <button class="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-500">Console</button>
        <button class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-500">LOGS</button>
        <button class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-500">BACKUP</button>
      </div>
    </div>

    <!-- Placeholder si aucun conteneur sélectionné -->
    <div v-else class="flex-1 flex items-center justify-center text-gray-500">
      <p>Sélectionnez un conteneur pour commencer.</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import axios from 'axios';

interface Container {
  id: string;
  name: string;
  status: string;
}

export default defineComponent({
  name: 'DockerUI',
  setup() {
    const sidebarOpen = ref(true);
    const containers = ref<Container[]>([]);
    const selectedContainer = ref<Container | null>(null);

    const toggleSidebar = () => {
      sidebarOpen.value = !sidebarOpen.value;
    };

    const selectContainer = (container: Container) => {
      selectedContainer.value = container;
    };

    const fetchContainers = async () => {
      try {
        const response = await axios.get('/api/containers/json');
        containers.value = response.data;
      } catch (error) {
        console.error('Error fetching containers:', error);
      }
    };

    const connectContainer = async (containerId: string) => {
      try {
        await axios.post(`/api/containers/${containerId}/start`);
        fetchContainers();
      } catch (error) {
        console.error('Error connecting container:', error);
      }
    };

    const disconnectContainer = async (containerId: string) => {
      try {
        await axios.post(`/api/containers/${containerId}/stop`);
        fetchContainers();
      } catch (error) {
        console.error('Error disconnecting container:', error);
      }
    };

    const addContainer = async () => {
      try {
        const response = await axios.post('/api/containers/create', {
          Image: 'your-image-name',
          name: 'new-container',
        });
        await axios.post(`/api/containers/${response.data.Id}/start`);
        fetchContainers();
      } catch (error) {
        console.error('Error adding container:', error);
      }
    };

    onMounted(() => {
      fetchContainers();
    });

    return {
      sidebarOpen,
      containers,
      selectedContainer,
      toggleSidebar,
      selectContainer,
      fetchContainers,
      connectContainer,
      disconnectContainer,
      addContainer,
    };
  },
});
</script>