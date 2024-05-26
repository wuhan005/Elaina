<template>
  <div class="container">
    <Header/>
    <div class="container-fluid">
      <router-view/>
    </div>
  </div>
</template>

<script setup lang="ts">

import Header from "./components/Header.vue";
</script>

<style scoped>
</style>
