<template>
  <div>
    <h1>Add New Container</h1>
    <form @submit.prevent="submitForm">
      <div>
        <label for="name">Container Name:</label>
        <input type="text" v-model="name" id="name" required>
      </div>
      <div>
        <label for="image">Image:</label>
        <input type="text" v-model="image" id="image" required>
      </div>
      <button type="submit">Add Container</button>
    </form>
  </div>
</template>

// Todo - Make this page accesible via the route /add-container

<script lang="ts">
import { defineComponent, ref } from 'vue';
import axios from 'axios';

export default defineComponent({
  name: 'AddContainer',
  setup() {
    const name = ref('');
    const image = ref('');
    const serverURL = 'http://localhost:3000';

    const submitForm = async () => {
      try {
        await axios.post(serverURL + '/api/containers/create', {
          name: name.value,
          image: image.value
        });
        // Redirect to home page after successful creation
        this.$router.push('/');
      } catch (error) {
        console.error('Error adding container:', error);
      }
    };

    return {
      name,
      image,
      submitForm
    };
  }
});
</script>

<style scoped>
/* Add your styles here */
</style>
