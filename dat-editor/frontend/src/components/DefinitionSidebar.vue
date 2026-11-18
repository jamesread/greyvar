<script setup>
defineProps({
  title: { type: String, required: true },
  items: { type: Array, default: () => [] },
  selected: { type: String, default: null },
  isEntdefs: { type: Boolean, default: true },
})

const emit = defineEmits(['select', 'create'])

function subtitle (item) {
  if (item.stateCount !== undefined) {
    const solid = item.solid ? ' · solid' : ''
    return `${item.title || item.name}${solid} · ${item.stateCount || 0} states`
  }
  return item.traversable ? 'walkable' : 'blocked'
}

function createNew () {
  const label = prompt('New definition name:')
  if (label) {
    emit('create', label.trim())
  }
}
</script>

<template>
  <section class="def-sidebar">
    <div class="def-sidebar-header">
      <strong>{{ title }}</strong>
      <button type="button" @click="createNew">
        New
      </button>
    </div>
    <ul class="list">
      <li v-for="item in items" :key="item.name">
        <button
          type="button"
          class="list-item"
          :class="{ active: selected === item.name }"
          @click="emit('select', item.name)"
        >
          <span>{{ item.name }}</span>
          <small>{{ subtitle(item) }}</small>
        </button>
      </li>
    </ul>
  </section>
</template>
