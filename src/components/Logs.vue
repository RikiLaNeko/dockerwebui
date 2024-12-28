<template>
  <div class="logs-container">
    <h2 class="logs-title">Logs for {{ containerId }}</h2>
    <pre class="logs-output">{{ logs }}</pre>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineProps } from 'vue';
import axios from 'axios';

const props = defineProps<{ containerId: string }>();

const logs = ref('');

const fetchLogs = async () => {
  try {
    const response = await axios.get(`http://localhost:3000/api/containers/${props.containerId}/logs`);
    logs.value = response.data;
  } catch (error) {
    console.error('Error fetching logs:', error);
  }
};

onMounted(() => {
  fetchLogs();
});
</script>

<style src="../css/Logs.css" scoped></style>
