<script setup>
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import Header from 'picocrank/vue/components/Header.vue'
import { initState } from './composables/useInit'

const route = useRoute()
const router = useRouter()

const siteTitle = computed(() => initState.siteTitle || 'Greyvar Dat Editor')
const metaText = computed(() => {
  if (initState.error) {
    return initState.error
  }
  if (!initState.ready) {
    return 'Connecting…'
  }
  return `dat: ${initState.datDir} · res: ${initState.resDir}`
})

function goTab (name) {
  router.push({ name })
}
</script>

<template>
  <div class="app-shell">
    <Header :title="siteTitle" :sidebar-enabled="false" :show-branding="true" :fixed="false" />
    <p class="meta">
      {{ metaText }}
    </p>
    <nav class="tabs">
      <button
        type="button"
        class="tab"
        :class="{ active: route.name === 'entdefs' }"
        @click="goTab('entdefs')"
      >
        Entdefs
      </button>
      <button
        type="button"
        class="tab"
        :class="{ active: route.name === 'tiledefs' }"
        @click="goTab('tiledefs')"
      >
        Tiledefs
      </button>
    </nav>
    <main class="layout">
      <RouterView />
    </main>
  </div>
</template>
