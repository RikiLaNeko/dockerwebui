<template>
  <DockerUI />
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import DockerUI from './components/DockerUI.vue';

export default defineComponent({
  name: 'App',
  components: {
    DockerUI,
  },
});
</script>

<style src="./style.css"></style>
