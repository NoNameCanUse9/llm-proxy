<script setup lang="ts">
// The main app entry only handles the router view
// To enforce dark mode, we can add a class to the body here
import { onMounted } from 'vue'
import { Toaster } from 'vue-sonner'

onMounted(() => {
  document.documentElement.classList.add('dark')
})
</script>

<template>
  <router-view />
  <Toaster position="top-right" theme="dark" rich-colors />
</template>

<style>
/* Any global styles that shouldn't be in Tailwind index.css can go here */
body {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  background-color: hsl(var(--background));
  color: hsl(var(--foreground));
  font-family: 'Inter', sans-serif;
}
</style>
