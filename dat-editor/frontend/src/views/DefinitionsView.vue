<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { client } from '../composables/client.js'
import DefinitionSidebar from '../components/DefinitionSidebar.vue'
import EntdefEditor from '../components/EntdefEditor.vue'
import TiledefEditor from '../components/TiledefEditor.vue'

const props = defineProps({
  kind: {
    type: String,
    required: true,
  },
})

const items = ref([])
const entityTextures = ref([])
const tileTextures = ref([])
const selectedName = ref(null)
const entdefDraft = ref(null)
const tiledefDraft = ref(null)
const status = ref('')
const error = ref('')

const isEntdefs = computed(() => props.kind === 'entdefs')

async function loadLists () {
  const [entdefsRes, tiledefsRes] = await Promise.all([
    client.listEntdefs({}),
    client.listTiledefs({}),
  ])
  items.value = isEntdefs.value ? entdefsRes.items : tiledefsRes.items

  const [entitiesRes, tilesRes] = await Promise.allSettled([
    client.listTextures({ kind: 'entities' }),
    client.listTextures({ kind: 'tiles' }),
  ])
  entityTextures.value = entitiesRes.status === 'fulfilled' ? entitiesRes.value.filenames : []
  tileTextures.value = tilesRes.status === 'fulfilled' ? tilesRes.value.filenames : []
}

async function refresh () {
  try {
    await loadLists()
    error.value = ''
  } catch (err) {
    error.value = err.message || 'Failed to load definitions'
  }
}

async function selectItem (name) {
  try {
    if (isEntdefs.value) {
      const item = await client.getEntdef({ name })
      selectedName.value = name
      entdefDraft.value = structuredClone(item)
      status.value = `Loaded entdef ${name}`
    } else {
      const item = await client.getTiledef({ name })
      selectedName.value = name
      tiledefDraft.value = structuredClone(item)
      status.value = `Loaded tiledef ${name}`
    }
    error.value = ''
  } catch (err) {
    error.value = err.message || 'Failed to load definition'
  }
}

async function createItem (name) {
  try {
    if (isEntdefs.value) {
      await client.createEntdef({
        name,
        title: name,
        initialState: 'idle',
        solid: false,
        states: { idle: { name: 'idle' } },
      })
    } else {
      await client.createTiledef({
        name,
        texture: '',
        traversable: true,
      })
    }
    await refresh()
    await selectItem(name)
    status.value = `Created ${name}`
  } catch (err) {
    error.value = err.message || 'Failed to create definition'
  }
}

async function saveEntdef () {
  const draft = entdefDraft.value
  await client.updateEntdef({
    name: selectedName.value,
    title: draft.title,
    initialState: draft.initialState,
    solid: draft.solid,
    states: draft.states,
  })
  await refresh()
  status.value = `Saved ${selectedName.value}`
}

async function deleteEntdef () {
  if (!confirm(`Delete entdef ${selectedName.value}?`)) {
    return
  }
  await client.deleteEntdef({ name: selectedName.value })
  selectedName.value = null
  entdefDraft.value = null
  await refresh()
  status.value = 'Deleted entdef'
}

async function saveTiledef () {
  const draft = tiledefDraft.value
  await client.updateTiledef({
    name: selectedName.value,
    texture: draft.texture,
    traversable: draft.traversable,
  })
  await refresh()
  status.value = `Saved ${selectedName.value}`
}

async function deleteTiledef () {
  if (!confirm(`Delete tiledef ${selectedName.value}?`)) {
    return
  }
  await client.deleteTiledef({ name: selectedName.value })
  selectedName.value = null
  tiledefDraft.value = null
  await refresh()
  status.value = 'Deleted tiledef'
}

async function handleSave () {
  try {
    if (isEntdefs.value) {
      await saveEntdef()
    } else {
      await saveTiledef()
    }
    error.value = ''
  } catch (err) {
    error.value = err.message || 'Failed to save'
  }
}

async function handleDelete () {
  try {
    if (isEntdefs.value) {
      await deleteEntdef()
    } else {
      await deleteTiledef()
    }
    error.value = ''
  } catch (err) {
    error.value = err.message || 'Failed to delete'
  }
}

watch(() => props.kind, async () => {
  selectedName.value = null
  entdefDraft.value = null
  tiledefDraft.value = null
  await refresh()
})

onMounted(refresh)
</script>

<template>
  <DefinitionSidebar
    :title="isEntdefs ? 'Entity definitions' : 'Tile definitions'"
    :items="items"
    :selected="selectedName"
    :is-entdefs="isEntdefs"
    @select="selectItem"
    @create="createItem"
  />
  <section class="editor-panel">
    <div class="panel-header">
      <strong>{{ selectedName || (isEntdefs ? 'Entity editor' : 'Tile editor') }}</strong>
    </div>
    <div class="panel-body">
      <EntdefEditor
        v-if="isEntdefs"
        v-model="entdefDraft"
        @save="handleSave"
        @delete="handleDelete"
      />
      <TiledefEditor
        v-else
        v-model="tiledefDraft"
        :textures="tileTextures"
        @save="handleSave"
        @delete="handleDelete"
      />
      <div v-if="error" class="error">
        {{ error }}
      </div>
      <div v-if="status" class="status">
        {{ status }}
      </div>
    </div>
  </section>
</template>
