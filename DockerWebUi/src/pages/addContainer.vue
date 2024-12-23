<template>
  <div>
    <h1>Add Container</h1>
    <form @submit.prevent="addContainer">
      <div class="form-group mb-3">
        <label for="containerName">Container Name</label>
        <input type="text" class="form-control" id="containerName" v-model="containerName" required>
      </div>
      <div class="form-group mb-3">
        <label for="containerImage">Container Image</label>
        <input type="text" class="form-control" id="containerImage" v-model="containerImage" required>
      </div>

</template>

<script lang="ts">
export default {
  data() {
    return {
      containerName: '',
      containerImage: '',
    };
  },
  methods: {
    addContainer() {
      this.$store.dispatch('addContainer', {
        containerName: this.containerName,
        containerImage: this.containerImage,
      });
    },
  },
};
</script>
