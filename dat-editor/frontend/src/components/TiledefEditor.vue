<script setup>
import { useId } from 'vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { textureUrl } from '../composables/client.js'

const model = defineModel({ type: Object, default: null })

defineProps({
  textures: { type: Array, default: () => [] },
})

defineEmits(['save', 'delete'])

const textureId = useId()

const traversableOptions = [
  { value: true, label: 'Yes' },
  { value: false, label: 'No' },
]
</script>

<template>
  <div v-if="!model" class="empty">
    Select a tile definition or create a new one.
  </div>
  <FormLayout v-else @submit.prevent="$emit('save')">
    <FormField label="Texture filename" :for="textureId">
      <select :id="textureId" v-model="model.texture">
        <option value="">
          Select texture…
        </option>
        <option v-for="name in textures" :key="name" :value="name">
          {{ name }}
        </option>
      </select>
    </FormField>

    <FormField label="Preview">
      <div class="preview-row">
        <img
          v-if="model.texture"
          :src="textureUrl('tiles', model.texture)"
          :alt="model.texture"
        >
        <div v-else class="empty">
          No texture selected
        </div>
      </div>
    </FormField>

    <FormField label="Traversable" component-has-label>
      <RadioGroup
        v-model="model.traversable"
        variant="boolean"
        :options="traversableOptions"
      />
    </FormField>

    <template #actions>
      <button type="submit">
        Save
      </button>
      <button type="button" class="danger" @click="$emit('delete')">
        Delete
      </button>
    </template>
  </FormLayout>
</template>
